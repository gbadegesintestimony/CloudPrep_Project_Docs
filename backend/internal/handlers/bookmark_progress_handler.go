package handlers

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloudprep-backend/internal/auth"
	"cloudprep-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookmarkProgressHandler struct {
	DB *gorm.DB
}

func NewBookmarkProgressHandler(database *gorm.DB) *BookmarkProgressHandler {
	return &BookmarkProgressHandler{
		DB: database,
	}
}

type bookmarkRequest struct {
	Note string `json:"note"`
}

type BookmarkItemDTO struct {
	ID               string              `json:"id"`
	QuestionID       string              `json:"questionId"`
	Note             string              `json:"note"`
	CreatedAt        time.Time           `json:"createdAt"`
	TopicID          string              `json:"topicId"`
	TopicTitle       string              `json:"topicTitle"`
	ModuleNumber     int                 `json:"moduleNumber"`
	WeekCode         string              `json:"weekCode"`
	Stem             string              `json:"stem"`
	QuestionStyle    string              `json:"questionStyle"`
	Difficulty       string              `json:"difficulty"`
	CorrectOptionKey string              `json:"correctOptionKey"`
	Explanation      string              `json:"explanation"`
	SourceReference  string              `json:"sourceReference"`
	Options          []ShuffledOptionDTO `json:"options"`
}

// ListBookmarks handles GET /api/bookmarks.
func (h *BookmarkProgressHandler) ListBookmarks(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)

	var bookmarks []models.QuestionBookmark
	if err := h.DB.Where("user_id = ?", userIDStr).Order("created_at DESC").Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to load bookmarked questions.",
		})
		return
	}

	qIDs := make([]string, 0, len(bookmarks))
	for _, b := range bookmarks {
		qIDs = append(qIDs, b.QuestionID)
	}

	var questions []models.Question
	if len(qIDs) > 0 {
		h.DB.Preload("Options").Where("id IN ?", qIDs).Find(&questions)
	}
	qMap := make(map[string]models.Question, len(questions))
	for _, q := range questions {
		qMap[q.ID] = q
	}

	var topics []models.CurriculumTopic
	h.DB.Find(&topics)
	topicMap := make(map[string]models.CurriculumTopic, len(topics))
	for _, t := range topics {
		topicMap[t.ID] = t
	}

	// C-1 / C-2 Integrity: Identify questions in active mock exams for this user
	var activeMockQIDs []string
	h.DB.Table("session_questions").
		Joins("JOIN practice_sessions ON practice_sessions.id = session_questions.session_id").
		Where("practice_sessions.user_id = ? AND practice_sessions.mode = ? AND practice_sessions.status = ?",
			userIDStr, models.ModeMockExam, models.SessionStatusActive).
		Pluck("session_questions.question_id", &activeMockQIDs)
	activeMockMap := make(map[string]bool, len(activeMockQIDs))
	for _, qid := range activeMockQIDs {
		activeMockMap[qid] = true
	}

	// Identify legitimately revealed questions (either submitted sessions or answered in immediate practice)
	var revealedQIDs []string
	h.DB.Table("session_questions").
		Joins("JOIN practice_sessions ON practice_sessions.id = session_questions.session_id").
		Where("practice_sessions.user_id = ? AND practice_sessions.status = ?", userIDStr, models.SessionStatusSubmitted).
		Pluck("DISTINCT session_questions.question_id", &revealedQIDs)

	var immediateAnsweredQIDs []string
	h.DB.Table("responses").
		Joins("JOIN practice_sessions ON practice_sessions.id = responses.session_id").
		Where("responses.user_id = ? AND practice_sessions.feedback_mode = ?", userIDStr, "IMMEDIATE").
		Pluck("DISTINCT responses.question_id", &immediateAnsweredQIDs)

	revealedMap := make(map[string]bool, len(revealedQIDs)+len(immediateAnsweredQIDs))
	for _, id := range revealedQIDs {
		revealedMap[id] = true
	}
	for _, id := range immediateAnsweredQIDs {
		revealedMap[id] = true
	}

	items := make([]BookmarkItemDTO, 0, len(bookmarks))
	for _, b := range bookmarks {
		q, ok := qMap[b.QuestionID]
		if !ok {
			continue
		}
		topic := topicMap[q.TopicID]

		// Redact answers if question is in an active mock exam OR has not been legitimately revealed yet
		isRedacted := activeMockMap[q.ID] || !revealedMap[q.ID]

		correctKey := q.CorrectOptionKey
		explanation := q.Explanation
		if isRedacted {
			correctKey = ""
			explanation = ""
		}

		opts := make([]ShuffledOptionDTO, 0, len(q.Options))
		for _, o := range q.Options {
			distractorRationale := o.DistractorRationale
			if isRedacted {
				distractorRationale = ""
			}
			opts = append(opts, ShuffledOptionDTO{
				Key:                 o.Key,
				Text:                o.Text,
				DistractorRationale: distractorRationale,
			})
		}
		items = append(items, BookmarkItemDTO{
			ID:               b.ID,
			QuestionID:       q.ID,
			Note:             b.Note,
			CreatedAt:        b.CreatedAt,
			TopicID:          q.TopicID,
			TopicTitle:       topic.Title,
			ModuleNumber:     topic.ModuleNumber,
			WeekCode:         topic.WeekCode,
			Stem:             q.Stem,
			QuestionStyle:    q.QuestionStyle,
			Difficulty:       q.Difficulty,
			CorrectOptionKey: correctKey,
			Explanation:      explanation,
			SourceReference:  q.SourceReference,
			Options:          opts,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": items,
		"total":     len(items),
	})
}

// UpsertBookmark handles POST /api/bookmarks/:questionId.
func (h *BookmarkProgressHandler) UpsertBookmark(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	questionID := strings.TrimSpace(c.Param("questionId"))

	var q models.Question
	if err := h.DB.Where("id = ?", questionID).First(&q).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "QUESTION_NOT_FOUND",
			"message": "Question not found.",
		})
		return
	}

	// C-1 Integrity: Ensure question has been served in at least one session for this user
	var servedCount int64
	h.DB.Table("session_questions").
		Joins("JOIN practice_sessions ON practice_sessions.id = session_questions.session_id").
		Where("practice_sessions.user_id = ? AND session_questions.question_id = ?", userIDStr, questionID).
		Count(&servedCount)

	if servedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "QUESTION_NOT_SERVED",
			"message": "You can only bookmark questions that have appeared in one of your practice sessions or mock exams.",
		})
		return
	}

	var req bookmarkRequest
	_ = c.ShouldBindJSON(&req)
	note := strings.TrimSpace(req.Note)
	runes := []rune(note)
	if len(runes) > 1000 {
		note = string(runes[:1000])
	}

	now := time.Now().UTC()
	bm := models.QuestionBookmark{
		ID:         fmt.Sprintf("bm-%s-%s", userIDStr, questionID),
		UserID:     userIDStr,
		QuestionID: questionID,
		Note:       note,
		CreatedAt:  now,
	}

	if err := h.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"note"}),
	}).Create(&bm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to save bookmark.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"bookmarked": true,
		"questionId": questionID,
		"note":       note,
		"createdAt":  now,
	})
}

// DeleteBookmark handles DELETE /api/bookmarks/:questionId.
func (h *BookmarkProgressHandler) DeleteBookmark(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	questionID := strings.TrimSpace(c.Param("questionId"))

	h.DB.Where("user_id = ? AND question_id = ?", userIDStr, questionID).Delete(&models.QuestionBookmark{})

	c.JSON(http.StatusOK, gin.H{
		"bookmarked": false,
		"questionId": questionID,
	})
}

type TopicProgressDTO struct {
	TopicID         string  `json:"topicId"`
	Slug            string  `json:"slug"`
	Title           string  `json:"title"`
	ModuleNumber    int     `json:"moduleNumber"`
	WeekCode        string  `json:"weekCode"`
	Attempted       int     `json:"attempted"`
	Total           int     `json:"total"`
	Correct         int     `json:"correct"`
	AccuracyPercent float64 `json:"accuracyPercent"`
	Accuracy        float64 `json:"accuracy"`
	MasteryStatus   string  `json:"masteryStatus"`
}

type StyleProgressDTO struct {
	QuestionStyle   string  `json:"questionStyle"`
	Attempted       int     `json:"attempted"`
	Total           int     `json:"total"`
	Correct         int     `json:"correct"`
	AccuracyPercent float64 `json:"accuracyPercent"`
	Accuracy        float64 `json:"accuracy"`
}

type ModuleProgressDTO struct {
	ModuleNumber    int     `json:"moduleNumber"`
	ModuleTitle     string  `json:"moduleTitle"`
	Attempted       int     `json:"attempted"`
	Total           int     `json:"total"`
	Correct         int     `json:"correct"`
	AccuracyPercent float64 `json:"accuracyPercent"`
	Accuracy        float64 `json:"accuracy"`
}

// GetProgressSummary handles GET /api/progress/summary.
func (h *BookmarkProgressHandler) GetProgressSummary(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)

	var totalBank int64
	h.DB.Model(&models.Question{}).Where("status = ?", models.QuestionStatusApproved).Count(&totalBank)

	var uniqueSeen int64
	h.DB.Model(&models.QuestionExposure{}).Where("user_id = ?", userIDStr).Count(&uniqueSeen)

	var bookmarkCount int64
	h.DB.Model(&models.QuestionBookmark{}).Where("user_id = ?", userIDStr).Count(&bookmarkCount)

	var completedSessions, activeSessions int64
	h.DB.Model(&models.PracticeSession{}).Where("user_id = ? AND status = ?", userIDStr, models.SessionStatusSubmitted).Count(&completedSessions)
	h.DB.Model(&models.PracticeSession{}).Where("user_id = ? AND status = ?", userIDStr, models.SessionStatusActive).Count(&activeSessions)

	var recentSessions []models.PracticeSession
	h.DB.Where("user_id = ?", userIDStr).Order("created_at DESC").Limit(10).Find(&recentSessions)

	// Load all user responses joined with Question metadata
	type responseJoinRow struct {
		QuestionID    string
		IsCorrect     bool
		TopicID       string
		QuestionStyle string
	}
	var rows []responseJoinRow
	h.DB.Table("responses").
		Select("responses.question_id, responses.is_correct, questions.topic_id, questions.question_style").
		Joins("JOIN questions ON questions.id = responses.question_id").
		Where("responses.user_id = ?", userIDStr).
		Scan(&rows)

	var topics []models.CurriculumTopic
	h.DB.Order("ordering ASC").Find(&topics)

	topicStats := make(map[string]*TopicProgressDTO, len(topics))
	for _, t := range topics {
		topicStats[t.ID] = &TopicProgressDTO{
			TopicID:       t.ID,
			Slug:          t.Slug,
			Title:         t.Title,
			ModuleNumber:  t.ModuleNumber,
			WeekCode:      t.WeekCode,
			MasteryStatus: "NOT_STARTED",
		}
	}

	styleStats := map[string]*StyleProgressDTO{
		models.StyleDefinition:    {QuestionStyle: models.StyleDefinition},
		models.StyleCommandSyntax: {QuestionStyle: models.StyleCommandSyntax},
		models.StyleComparison:    {QuestionStyle: models.StyleComparison},
		models.StyleScenario:      {QuestionStyle: models.StyleScenario},
	}

	mod1 := ModuleProgressDTO{ModuleNumber: 1, ModuleTitle: "Module 1: Cloud & DevOps Foundations"}
	mod2 := ModuleProgressDTO{ModuleNumber: 2, ModuleTitle: "Module 2: Cloud Architecture, Services & AI"}

	totalAnswered := len(rows)
	totalCorrect := 0

	for _, r := range rows {
		if r.IsCorrect {
			totalCorrect++
		}
		if ts, ok := topicStats[r.TopicID]; ok {
			ts.Attempted++
			if r.IsCorrect {
				ts.Correct++
			}
			if ts.ModuleNumber == 1 {
				mod1.Attempted++
				if r.IsCorrect {
					mod1.Correct++
				}
			} else {
				mod2.Attempted++
				if r.IsCorrect {
					mod2.Correct++
				}
			}
		}
		if ss, ok := styleStats[r.QuestionStyle]; ok {
			ss.Attempted++
			if r.IsCorrect {
				ss.Correct++
			}
		}
	}

	overallAccuracy := 0.0
	if totalAnswered > 0 {
		overallAccuracy = math.Round((float64(totalCorrect)/float64(totalAnswered))*10000) / 100
	}
	mod1.Total = mod1.Attempted
	if mod1.Attempted > 0 {
		mod1.AccuracyPercent = math.Round((float64(mod1.Correct)/float64(mod1.Attempted))*10000) / 100
		mod1.Accuracy = mod1.AccuracyPercent
	}
	mod2.Total = mod2.Attempted
	if mod2.Attempted > 0 {
		mod2.AccuracyPercent = math.Round((float64(mod2.Correct)/float64(mod2.Attempted))*10000) / 100
		mod2.Accuracy = mod2.AccuracyPercent
	}

	topicProgressList := make([]TopicProgressDTO, 0, len(topics))
	for _, t := range topics {
		ts := topicStats[t.ID]
		ts.Total = ts.Attempted
		if ts.Attempted > 0 {
			ts.AccuracyPercent = math.Round((float64(ts.Correct)/float64(ts.Attempted))*10000) / 100
			ts.Accuracy = ts.AccuracyPercent
			switch {
			case ts.AccuracyPercent >= 80 && ts.Attempted >= 5:
				ts.MasteryStatus = "MASTERED"
			case ts.AccuracyPercent >= 65:
				ts.MasteryStatus = "PROFICIENT"
			default:
				ts.MasteryStatus = "NEEDS_REVIEW"
			}
		}
		topicProgressList = append(topicProgressList, *ts)
	}

	// Recommend up to 3 weak topics based strictly on actual student attempts (never show weak topics to new accounts with 0 answers)
	recommendedWeakTopics := make([]TopicProgressDTO, 0, 3)
	if totalAnswered > 0 {
		weakCandidates := make([]TopicProgressDTO, 0, len(topicProgressList))
		for _, tp := range topicProgressList {
			if tp.Attempted > 0 && tp.AccuracyPercent < 80 {
				weakCandidates = append(weakCandidates, tp)
			}
		}
		sort.SliceStable(weakCandidates, func(i, j int) bool {
			if weakCandidates[i].AccuracyPercent != weakCandidates[j].AccuracyPercent {
				return weakCandidates[i].AccuracyPercent < weakCandidates[j].AccuracyPercent
			}
			return weakCandidates[i].Attempted > weakCandidates[j].Attempted
		})
		for _, wc := range weakCandidates {
			if len(recommendedWeakTopics) >= 3 {
				break
			}
			recommendedWeakTopics = append(recommendedWeakTopics, wc)
		}
	}

	styleOrder := []string{
		models.StyleDefinition,
		models.StyleCommandSyntax,
		models.StyleComparison,
		models.StyleScenario,
	}
	styleProgressList := make([]StyleProgressDTO, 0, 4)
	for _, st := range styleOrder {
		ss := styleStats[st]
		ss.Total = ss.Attempted
		if ss.Attempted > 0 {
			ss.AccuracyPercent = math.Round((float64(ss.Correct)/float64(ss.Attempted))*10000) / 100
			ss.Accuracy = ss.AccuracyPercent
		}
		styleProgressList = append(styleProgressList, *ss)
	}

	unseenRemaining := totalBank - uniqueSeen
	if unseenRemaining < 0 {
		unseenRemaining = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"completedSessions":        completedSessions,
		"totalCompletedSessions":   completedSessions,
		"activeSessions":           activeSessions,
		"activeSessionsCount":      activeSessions,
		"totalQuestionsAnswered":   totalAnswered,
		"totalAnswered":            totalAnswered,
		"totalCorrect":             totalCorrect,
		"overallAccuracyPercent":   overallAccuracy,
		"overallAccuracy":          overallAccuracy,
		"totalBankQuestions":       totalBank,
		"uniqueQuestionsSeen":      uniqueSeen,
		"unseenQuestionsRemaining": unseenRemaining,
		"bookmarkCount":            bookmarkCount,
		"totalBookmarks":           bookmarkCount,
		"moduleBreakdown":          []ModuleProgressDTO{mod1, mod2},
		"topicBreakdown":           topicProgressList,
		"styleBreakdown":           styleProgressList,
		"recommendedWeakTopics":    recommendedWeakTopics,
		"recentSessions":           recentSessions,
	})
}

