package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloudprep-backend/internal/ai"
	"cloudprep-backend/internal/auth"
	"cloudprep-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB     *gorm.DB
	Router *ai.MultiProviderAiRouter
	genMu  sync.Mutex // Serializes on-demand batch generation calls (Hardened V-4)
}

func NewAdminHandler(database *gorm.DB, aiRouter *ai.MultiProviderAiRouter) *AdminHandler {
	return &AdminHandler{
		DB:     database,
		Router: aiRouter,
	}
}

// ListQuestions handles GET /api/admin/questions (search & filter question bank).
func (h *AdminHandler) ListQuestions(c *gin.Context) {
	topicID := strings.TrimSpace(c.Query("topicId"))
	style := strings.ToUpper(strings.TrimSpace(c.Query("style")))
	diff := strings.ToUpper(strings.TrimSpace(c.Query("difficulty")))
	sourceType := strings.ToUpper(strings.TrimSpace(c.Query("sourceType")))
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	search := strings.TrimSpace(c.Query("q"))

	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		if l > 100 {
			l = 100
		}
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	// Base count query without preload overhead (Hardened V-5)
	baseQuery := h.DB.Model(&models.Question{})

	if topicID != "" {
		baseQuery = baseQuery.Where("topic_id = ?", topicID)
	}
	if style != "" && style != "ALL" {
		baseQuery = baseQuery.Where("question_style = ?", style)
	}
	if diff != "" && diff != "ALL" {
		baseQuery = baseQuery.Where("difficulty = ?", diff)
	}
	if sourceType != "" && sourceType != "ALL" {
		baseQuery = baseQuery.Where("source_type = ?", sourceType)
	}
	if status != "" && status != "ALL" {
		baseQuery = baseQuery.Where("status = ?", status)
	}
	if search != "" {
		baseQuery = baseQuery.Where("stem LIKE ?", "%"+search+"%")
	}

	var total int64
	baseQuery.Count(&total)

	var questions []models.Question
	if err := baseQuery.Preload("Options").Order("created_at DESC").Limit(limit).Offset(offset).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to query question bank.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"questions": questions,
	})
}

// ListGenerationRecords handles GET /api/admin/ai/records (inspect generation history and failovers).
func (h *AdminHandler) ListGenerationRecords(c *gin.Context) {
	topicID := strings.TrimSpace(c.Query("topicId"))
	provider := strings.ToLower(strings.TrimSpace(c.Query("provider")))

	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	query := h.DB.Model(&models.GenerationRecord{})
	if topicID != "" {
		query = query.Where("topic_id = ?", topicID)
	}
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}

	var total int64
	query.Count(&total)

	var records []models.GenerationRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to query generation records.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"records": records,
	})
}

type batchGenerateRequest struct {
	TopicID       string `json:"topicId" binding:"required"`
	QuestionStyle string `json:"questionStyle"`
	Difficulty    string `json:"difficulty"`
	Count         int    `json:"count"`
}

// BatchGenerate handles POST /api/admin/ai/generate (trigger on-demand AI question generation).
func (h *AdminHandler) BatchGenerate(c *gin.Context) {
	if h.Router == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "AI_ROUTER_UNAVAILABLE",
			"message": "Dual-AI generation router is not initialized.",
		})
		return
	}

	// Concurrency lock to prevent multiple batch calls from overwhelming AI providers (Hardened V-4)
	if !h.genMu.TryLock() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "GENERATION_IN_PROGRESS",
			"message": "Another AI generation task is currently running. Please wait for it to complete.",
		})
		return
	}
	defer h.genMu.Unlock()

	var req batchGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Please provide a valid topicId.",
		})
		return
	}

	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}

	// Normalize and sanitize style and difficulty against allowed enums (Hardened V-2)
	style := ai.NormalizeStyle(req.QuestionStyle, models.StyleScenario)
	diff := ai.NormalizeDifficulty(req.Difficulty)

	genReq := ai.GenerationRequest{
		TopicID:       strings.TrimSpace(req.TopicID),
		QuestionStyle: style,
		Difficulty:    diff,
		Count:         count,
	}

	accepted, record, err := h.Router.GenerateAndPersist(c.Request.Context(), genReq)
	if err != nil {
		// Scrub upstream error messages to prevent any internal credential leakage (Hardened V-1)
		cleanErr := strings.ReplaceAll(err.Error(), "\n", " ")
		if strings.Contains(strings.ToLower(cleanErr), "key") || strings.Contains(cleanErr, "http") {
			cleanErr = "Upstream AI provider error occurred during generation."
		}
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "GENERATION_FAILED",
			"message": fmt.Sprintf("AI question generation failed: %s", cleanErr),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"acceptedCount": len(accepted),
		"record":        record,
		"questions":     accepted,
	})
}

type updateQuestionStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateQuestionStatus handles PATCH /api/admin/questions/:questionId (Approve/Retire questions).
func (h *AdminHandler) UpdateQuestionStatus(c *gin.Context) {
	adminID, _ := c.Get(auth.ContextUserIDKey)
	adminIDStr := fmt.Sprintf("%v", adminID)
	questionID := c.Param("questionId")

	var req updateQuestionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Status is required (APPROVED or RETIRED).",
		})
		return
	}

	newStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if newStatus != models.QuestionStatusApproved && newStatus != "RETIRED" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_STATUS",
			"message": "Status must be APPROVED or RETIRED.",
		})
		return
	}

	var q models.Question
	if err := h.DB.Where("id = ?", questionID).First(&q).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "QUESTION_NOT_FOUND",
			"message": "Question not found in bank.",
		})
		return
	}

	prevStatus := q.Status
	h.DB.Model(&q).Update("status", newStatus)

	// Audit trail
	meta, _ := json.Marshal(map[string]string{
		"previousStatus": prevStatus,
		"newStatus":      newStatus,
	})
	_ = h.DB.Create(&models.AuditEvent{
		ID:           auth.GenerateID("aud"),
		ActorID:      adminIDStr,
		Action:       "QUESTION_STATUS_CHANGE",
		EntityType:   "Question",
		EntityID:     questionID,
		MetadataJSON: string(meta),
		CreatedAt:    time.Now().UTC(),
	})

	c.JSON(http.StatusOK, gin.H{
		"questionId": questionID,
		"status":     newStatus,
	})
}

// GetBankStats handles GET /api/admin/stats (curriculum question bank growth analytics).
func (h *AdminHandler) GetBankStats(c *gin.Context) {
	var totalQuestions int64
	h.DB.Model(&models.Question{}).Count(&totalQuestions)

	var seedCount int64
	h.DB.Model(&models.Question{}).Where("source_type = ?", models.SourceSeedBank).Count(&seedCount)

	var geminiCount int64
	h.DB.Model(&models.Question{}).Where("source_type = ?", models.SourceAIGemini).Count(&geminiCount)

	var groqCount int64
	h.DB.Model(&models.Question{}).Where("source_type = ?", models.SourceAIGroq).Count(&groqCount)

	var totalRecords int64
	h.DB.Model(&models.GenerationRecord{}).Count(&totalRecords)

	var failoverRecords int64
	h.DB.Model(&models.GenerationRecord{}).Where("failover_used = ?", true).Count(&failoverRecords)

	c.JSON(http.StatusOK, gin.H{
		"totalQuestions": totalQuestions,
		"sourceBreakdown": gin.H{
			"seedBank": seedCount,
			"aiGemini": geminiCount,
			"aiGroq":   groqCount,
		},
		"generationStats": gin.H{
			"totalRecords":  totalRecords,
			"failoverCount": failoverRecords,
		},
	})
}

