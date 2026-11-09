package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"cloudprep-backend/internal/curriculum"
	"cloudprep-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	healthCacheTTL     = 15 * time.Second
	curriculumCacheTTL = 30 * time.Second
	dbQueryTimeout     = 3 * time.Second
)

type cachedHealthSnapshot struct {
	dbActive       bool
	topicCount     int64
	objectiveCount int64
	bankCount      int64
	fetchedAt      time.Time
}

type HealthHandler struct {
	DB       *gorm.DB
	DBDriver string

	healthMu    sync.Mutex
	healthCache cachedHealthSnapshot

	currMu        sync.Mutex
	currCacheBody gin.H
	currCachedAt  time.Time
}

func NewHealthHandler(database *gorm.DB, dbDriver string) *HealthHandler {
	return &HealthHandler{
		DB:       database,
		DBDriver: dbDriver,
	}
}

// InvalidateCache clears the in-memory health and curriculum caches when questions are seeded or generated.
func (h *HealthHandler) InvalidateCache() {
	h.healthMu.Lock()
	h.healthCache = cachedHealthSnapshot{}
	h.healthMu.Unlock()

	h.currMu.Lock()
	h.currCacheBody = nil
	h.currCachedAt = time.Time{}
	h.currMu.Unlock()
}

// GetHealth handles GET /api/health.
// Dual-purpose:
//  1. Wakes the Render free-tier Golang/Gin web service when a user lands on / or /login.
//  2. Executes a lightweight database query (SELECT 1 + counts) with a 15s singleflight cache TTL
//     so Supabase PostgreSQL registers active database traffic (preventing 7-day inactivity auto-pause)
//     while completely preventing SQL connection-pool exhaustion under high concurrency or flood attacks.
func (h *HealthHandler) GetHealth(c *gin.Context) {
	h.healthMu.Lock()
	now := time.Now().UTC()
	snap := h.healthCache
	if now.Sub(snap.fetchedAt) >= healthCacheTTL {
		ctx, cancel := context.WithTimeout(c.Request.Context(), dbQueryTimeout)
		dbCtx := h.DB.WithContext(ctx)

		var probe int
		dbActive := true
		if err := dbCtx.Raw("SELECT 1").Scan(&probe).Error; err != nil || probe != 1 {
			dbActive = false
		}

		var topicCount, objectiveCount, bankCount int64
		if dbActive {
			dbCtx.Model(&models.CurriculumTopic{}).Count(&topicCount)
			dbCtx.Model(&models.LearningObjective{}).Count(&objectiveCount)
			dbCtx.Model(&models.Question{}).Where("status = ?", models.QuestionStatusApproved).Count(&bankCount)
		}
		cancel()

		snap = cachedHealthSnapshot{
			dbActive:       dbActive,
			topicCount:     topicCount,
			objectiveCount: objectiveCount,
			bankCount:      bankCount,
			fetchedAt:      now,
		}
		h.healthCache = snap
	}
	h.healthMu.Unlock()

	status := http.StatusOK
	statusLabel := "ok"
	if !snap.dbActive {
		status = http.StatusServiceUnavailable
		statusLabel = "degraded"
	}

	c.JSON(status, gin.H{
		"status":         statusLabel,
		"dbActive":       snap.dbActive,
		"topicCount":     snap.topicCount,
		"objectiveCount": snap.objectiveCount,
		"bankCount":      snap.bankCount,
		"timestamp":      now.Format(time.RFC3339),
	})
}

type TopicResponseDTO struct {
	ID            string                     `json:"id"`
	Slug          string                     `json:"slug"`
	Name          string                     `json:"name"`
	Title         string                     `json:"title"`
	ModuleNumber  int                        `json:"moduleNumber"`
	ModuleTitle   string                     `json:"moduleTitle"`
	ModuleCode    string                     `json:"moduleCode"`
	WeekCode      string                     `json:"weekCode"`
	SourcePDF     string                     `json:"sourcePdf"`
	Description   string                     `json:"description"`
	DisplayRank   int                        `json:"displayRank"`
	Ordering      int                        `json:"ordering"`
	QuestionCount int64                      `json:"questionCount"`
	StyleCounts   map[string]int64           `json:"styleCounts"`
	Objectives    []models.LearningObjective `json:"objectives"`
}

// GetCurriculum handles GET /api/curriculum.
// Returns the 10 course curriculum topics (with learning objectives, excluding internal AI grounding excerpts)
// plus the 3 built-in Timed Mock Exam presets, protected by a 30s singleflight cache and 3s DB timeout.
func (h *HealthHandler) GetCurriculum(c *gin.Context) {
	h.currMu.Lock()
	now := time.Now().UTC()
	if h.currCacheBody != nil && now.Sub(h.currCachedAt) < curriculumCacheTTL {
		cached := h.currCacheBody
		h.currMu.Unlock()
		c.JSON(http.StatusOK, cached)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), dbQueryTimeout)
	defer cancel()
	dbCtx := h.DB.WithContext(ctx)

	var topics []models.CurriculumTopic
	if err := dbCtx.Preload("Objectives").Order("ordering ASC").Find(&topics).Error; err != nil {
		h.currMu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load curriculum topics",
		})
		return
	}

	type countRow struct {
		TopicID       string
		QuestionStyle string
		Total         int64
	}
	var rows []countRow
	dbCtx.Model(&models.Question{}).
		Select("topic_id, question_style, COUNT(*) as total").
		Where("status = ?", models.QuestionStatusApproved).
		Group("topic_id, question_style").
		Scan(&rows)

	topicTotals := make(map[string]int64)
	topicStyles := make(map[string]map[string]int64)
	var totalApproved int64

	for _, r := range rows {
		topicTotals[r.TopicID] += r.Total
		totalApproved += r.Total
		if _, ok := topicStyles[r.TopicID]; !ok {
			topicStyles[r.TopicID] = map[string]int64{
				models.StyleDefinition:    0,
				models.StyleCommandSyntax: 0,
				models.StyleComparison:    0,
				models.StyleScenario:      0,
			}
		}
		topicStyles[r.TopicID][r.QuestionStyle] = r.Total
	}

	dtoList := make([]TopicResponseDTO, 0, len(topics))
	for _, t := range topics {
		styles, ok := topicStyles[t.ID]
		if !ok {
			styles = map[string]int64{
				models.StyleDefinition:    0,
				models.StyleCommandSyntax: 0,
				models.StyleComparison:    0,
				models.StyleScenario:      0,
			}
		}
		dtoList = append(dtoList, TopicResponseDTO{
			ID:            t.ID,
			Slug:          t.Slug,
			Name:          t.Title,
			Title:         t.Title,
			ModuleNumber:  t.ModuleNumber,
			ModuleTitle:   t.ModuleTitle,
			ModuleCode:    fmt.Sprintf("Module %d", t.ModuleNumber),
			WeekCode:      t.WeekCode,
			SourcePDF:     t.SourcePdf,
			Description:   t.Description,
			DisplayRank:   t.Ordering,
			Ordering:      t.Ordering,
			QuestionCount: topicTotals[t.ID],
			StyleCounts:   styles,
			Objectives:    t.Objectives,
		})
	}

	payload := gin.H{
		"topics":             dtoList,
		"examPresets":        curriculum.GetExamPresets(),
		"totalBankQuestions": totalApproved,
		"supportedStyles": []string{
			models.StyleDefinition,
			models.StyleCommandSyntax,
			models.StyleComparison,
			models.StyleScenario,
		},
		"supportedDifficulties": []string{
			models.DifficultyFoundational,
			models.DifficultyIntermediate,
			models.DifficultyAdvanced,
		},
	}

	h.currCacheBody = payload
	h.currCachedAt = now
	h.currMu.Unlock()

	c.JSON(http.StatusOK, payload)
}
