package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloudprep-backend/internal/auth"
	"cloudprep-backend/internal/curriculum"
	"cloudprep-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReplenishmentHook is an optional non-blocking callback invoked when a learner's unseen
// question pool for a topic drops below threshold (wired in Sprint 4 to Dual-AI Gemini+Groq).
type ReplenishmentHook func(topicID, questionStyle, difficulty string)

type SessionHandler struct {
	DB          *gorm.DB
	OnLowUnseen ReplenishmentHook
}

func NewSessionHandler(database *gorm.DB) *SessionHandler {
	return &SessionHandler{
		DB: database,
	}
}

type createSessionRequest struct {
	Mode                string   `json:"mode"`
	ExamPreset          string   `json:"examPreset"`
	Preset              string   `json:"preset"`
	Title               string   `json:"title"`
	TopicIDs            []string `json:"topicIds"`
	TopicID             string   `json:"topicId"`
	QuestionStyleFilter string   `json:"questionStyle"`
	DifficultyFilter    string   `json:"difficulty"`
	FeedbackMode        string   `json:"feedbackMode"`
	QuestionCount       int      `json:"questionCount"`
	TimeLimitSeconds    int      `json:"timeLimitSeconds"`
	CancelPrevious      bool     `json:"cancelPrevious"`
}

type submitAnswerRequest struct {
	SelectedOptionKey string `json:"selectedOptionKey"`
}

type flagQuestionRequest struct {
	IsFlagged *bool `json:"isFlagged"`
}

// ShuffledOptionDTO represents a single option (A, B, C, or D) in its per-session shuffled order.
// Notice DistractorRationale is omitted during active sessions and populated only in immediate feedback / post-submit review.
type ShuffledOptionDTO struct {
	Key                 string `json:"key"`
	Text                string `json:"text"`
	DistractorRationale string `json:"distractorRationale,omitempty"`
}

// ActiveSessionQuestionDTO is returned during an active session.
// CorrectOptionKey, Explanation, and DistractorRationale are strictly stripped on active Mock Exams.
type ActiveSessionQuestionDTO struct {
	QuestionID        string              `json:"questionId"`
	Ordering          int                 `json:"ordering"`
	TopicID           string              `json:"topicId"`
	ObjectiveID       string              `json:"objectiveId"`
	Stem              string              `json:"stem"`
	QuestionStyle     string              `json:"questionStyle"`
	Difficulty        string              `json:"difficulty"`
	SourceReference   string              `json:"sourceReference"`
	IsFlagged         bool                `json:"isFlagged"`
	IsBookmarked      bool                `json:"isBookmarked"`
	SelectedOptionKey string              `json:"selectedOptionKey,omitempty"`
	Options           []ShuffledOptionDTO `json:"options"`
	// Populated ONLY when already answered in IMMEDIATE feedback Practice Mode (never in ACTIVE Mock Exam)
	IsCorrect        *bool  `json:"isCorrect,omitempty"`
	CorrectOptionKey string `json:"correctOptionKey,omitempty"`
	Explanation      string `json:"explanation,omitempty"`
}

// ReviewedQuestionDTO is returned ONLY after a session is SUBMITTED (or for immediate-feedback practice).
type ReviewedQuestionDTO struct {
	QuestionID        string              `json:"questionId"`
	Ordering          int                 `json:"ordering"`
	TopicID           string              `json:"topicId"`
	TopicTitle        string              `json:"topicTitle"`
	ModuleNumber      int                 `json:"moduleNumber"`
	WeekCode          string              `json:"weekCode"`
	ObjectiveID       string              `json:"objectiveId"`
	Stem              string              `json:"stem"`
	QuestionStyle     string              `json:"questionStyle"`
	Difficulty        string              `json:"difficulty"`
	SourceType        string              `json:"sourceType"`
	SourceReference   string              `json:"sourceReference"`
	IsFlagged         bool                `json:"isFlagged"`
	IsBookmarked      bool                `json:"isBookmarked"`
	SelectedOptionKey string              `json:"selectedOptionKey"`
	CorrectOptionKey  string              `json:"correctOptionKey"`
	IsCorrect         bool                `json:"isCorrect"`
	IsAnswered        bool                `json:"isAnswered"`
	Explanation       string              `json:"explanation"`
	Options           []ShuffledOptionDTO `json:"options"`
}

func generateOptionShuffleMapping() (map[string]string, string) {
	displayKeys := []string{"A", "B", "C", "D"}
	canonicalKeys := []string{"A", "B", "C", "D"}
	rand.Shuffle(len(canonicalKeys), func(i, j int) {
		canonicalKeys[i], canonicalKeys[j] = canonicalKeys[j], canonicalKeys[i]
	})

	mapping := make(map[string]string, 4)
	for i, dKey := range displayKeys {
		mapping[dKey] = canonicalKeys[i]
	}
	raw, _ := json.Marshal(mapping)
	return mapping, string(raw)
}

func parseOptionMapping(raw string) map[string]string {
	m := make(map[string]string, 4)
	if err := json.Unmarshal([]byte(raw), &m); err != nil || len(m) != 4 {
		return map[string]string{"A": "A", "B": "B", "C": "C", "D": "D"}
	}
	return m
}

func invertOptionMapping(displayToCanonical map[string]string) map[string]string {
	inv := make(map[string]string, 4)
	for dKey, cKey := range displayToCanonical {
		inv[cKey] = dKey
	}
	return inv
}

// CreateSession handles POST /api/practice-sessions (and POST /api/sessions).
func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)

	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Invalid session configuration payload.",
		})
		return
	}

	mode := strings.ToUpper(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = models.ModeTopicPractice
	}
	validModes := map[string]bool{
		models.ModeTopicPractice:    true,
		models.ModeMixedPractice:    true,
		models.ModeMockExam:         true,
		models.ModeReviewMistakes:   true,
		models.ModeBookmarkRevision: true,
	}
	if !validModes[mode] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_MODE",
			"message": "Unsupported session mode.",
		})
		return
	}

	// If requested, cancel any previous active sessions before creating this new one
	if req.CancelPrevious {
		h.DB.Model(&models.PracticeSession{}).
			Where("user_id = ? AND status = ?", userIDStr, models.SessionStatusActive).
			Updates(map[string]interface{}{
				"status":     models.SessionStatusCancelled,
				"updated_at": time.Now().UTC(),
			})
	}

	// Resource Protection: Cap active sessions per user to 5 to prevent DB flooding
	var activeSessions []models.PracticeSession
	h.DB.Where("user_id = ? AND status = ?", userIDStr, models.SessionStatusActive).
		Order("created_at DESC").
		Find(&activeSessions)

	if len(activeSessions) >= 5 {
		c.JSON(http.StatusConflict, gin.H{
			"error":          "MAX_ACTIVE_SESSIONS_EXCEEDED",
			"message":        "You have 5 active sessions in progress. Please continue an existing session or cancel them to start fresh.",
			"activeSessions": activeSessions,
		})
		return
	}

	// Exam Integrity: Disallow starting multiple concurrent mock exams
	if mode == models.ModeMockExam {
		var activeMocks []models.PracticeSession
		h.DB.Where("user_id = ? AND mode = ? AND status = ?", userIDStr, models.ModeMockExam, models.SessionStatusActive).
			Order("created_at DESC").
			Find(&activeMocks)
		if len(activeMocks) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":          "ACTIVE_MOCK_EXAM_EXISTS",
				"message":        "You have an active mock exam in progress. Please complete, continue, or cancel your current mock exam before starting another.",
				"activeSessions": activeMocks,
			})
			return
		}
	}

	preset := strings.ToUpper(strings.TrimSpace(req.ExamPreset))
	if preset == "" {
		preset = strings.ToUpper(strings.TrimSpace(req.Preset))
	}
	if preset == "" {
		preset = models.PresetCustom
	}

	topicIDs := make([]string, 0)
	if req.TopicID != "" {
		topicIDs = append(topicIDs, strings.TrimSpace(req.TopicID))
	}
	for _, tid := range req.TopicIDs {
		trimmed := strings.TrimSpace(tid)
		if trimmed != "" {
			topicIDs = append(topicIDs, trimmed)
		}
	}

	questionCount := req.QuestionCount
	timeLimitSeconds := req.TimeLimitSeconds
	title := strings.TrimSpace(req.Title)

	// Apply Mock Exam Preset defaults if specified (Strictly enforces preset question count and timer, H-5)
	if mode == models.ModeMockExam && preset != models.PresetCustom {
		for _, p := range curriculum.GetExamPresets() {
			if p.ID == preset {
				topicIDs = p.TopicIDs
				questionCount = p.QuestionCount
				timeLimitSeconds = p.TimeLimitMinutes * 60
				title = p.Title
				break
			}
		}
	} else if mode == models.ModeMockExam && preset == models.PresetCustom {
		// Clamp custom mock exam time limits between 60s and 3 hours (10,800s)
		if timeLimitSeconds < 60 {
			timeLimitSeconds = 60
		} else if timeLimitSeconds > 10800 {
			timeLimitSeconds = 10800
		}
	}

	// Default topic selection if none provided
	if len(topicIDs) == 0 {
		for _, t := range curriculum.GetCurriculumTopics() {
			topicIDs = append(topicIDs, t.ID)
		}
	}

	if questionCount <= 0 {
		if mode == models.ModeMockExam {
			questionCount = 30
		} else {
			questionCount = 10
		}
	}
	if questionCount > 100 {
		questionCount = 100
	}

	if mode == models.ModeMockExam && timeLimitSeconds <= 0 {
		timeLimitSeconds = questionCount * 80 // ~1.33 min per question default
	}

	validStyles := map[string]bool{
		models.StyleDefinition:    true,
		models.StyleCommandSyntax: true,
		models.StyleComparison:   true,
		models.StyleScenario:     true,
		"ALL":                     true,
	}
	styleFilter := strings.ToUpper(strings.TrimSpace(req.QuestionStyleFilter))
	if !validStyles[styleFilter] || mode == models.ModeMockExam {
		styleFilter = "ALL"
	}

	validDiffs := map[string]bool{
		models.DifficultyFoundational: true,
		models.DifficultyIntermediate: true,
		models.DifficultyAdvanced:     true,
		"ALL":                         true,
	}
	diffFilter := strings.ToUpper(strings.TrimSpace(req.DifficultyFilter))
	if !validDiffs[diffFilter] {
		diffFilter = "ALL"
	}

	feedbackMode := strings.ToUpper(strings.TrimSpace(req.FeedbackMode))
	if mode == models.ModeMockExam {
		feedbackMode = "DEFERRED"
	} else if feedbackMode != "DEFERRED" {
		feedbackMode = "IMMEDIATE"
	}

	if title == "" {
		switch mode {
		case models.ModeMockExam:
			title = "Timed Cloud Engineering Mock Exam"
		case models.ModeBookmarkRevision:
			title = "Bookmarked Questions Revision"
		case models.ModeReviewMistakes:
			title = "Mistake Recovery Practice"
		default:
			title = "Cloud Engineering Practice Session"
		}
	}

	// Select eligible questions with Unseen-First prioritization and balanced style distribution
	selectedQuestions, unseenRemaining, err := h.selectQuestionsForSession(userIDStr, mode, topicIDs, styleFilter, diffFilter, questionCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SELECTION_ERROR",
			"message": "Failed to select questions from bank.",
		})
		return
	}
	if len(selectedQuestions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "NO_MATCHING_QUESTIONS",
			"message": "No questions match the selected filter criteria.",
		})
		return
	}

	// Trigger non-blocking background AI replenishment if unseen pool is low (< 25)
	if unseenRemaining < 25 && h.OnLowUnseen != nil && len(topicIDs) > 0 {
		targetTopic := topicIDs[rand.IntN(len(topicIDs))]
		go h.OnLowUnseen(targetTopic, styleFilter, diffFilter)
	}

	now := time.Now().UTC()
	var deadline *time.Time
	if timeLimitSeconds > 0 {
		dl := now.Add(time.Duration(timeLimitSeconds) * time.Second)
		deadline = &dl
	}

	topicIDsBytes, _ := json.Marshal(topicIDs)
	sessionID := auth.GenerateID("sess")

	session := models.PracticeSession{
		ID:                  sessionID,
		UserID:              userIDStr,
		Mode:                mode,
		ExamPreset:          preset,
		Title:               title,
		TopicIDsJSON:        string(topicIDsBytes),
		QuestionStyleFilter: styleFilter,
		DifficultyFilter:    diffFilter,
		FeedbackMode:        feedbackMode,
		TotalQuestions:      len(selectedQuestions),
		TimeLimitSeconds:    timeLimitSeconds,
		Status:              models.SessionStatusActive,
		StartedAt:           now,
		Deadline:            deadline,
		UnansweredCount:     len(selectedQuestions),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	sessionQuestions := make([]models.SessionQuestion, 0, len(selectedQuestions))
	for idx, q := range selectedQuestions {
		_, mappingJSON := generateOptionShuffleMapping()
		sessionQuestions = append(sessionQuestions, models.SessionQuestion{
			ID:                    fmt.Sprintf("%s-sq-%03d", sessionID, idx+1),
			SessionID:             sessionID,
			QuestionID:            q.ID,
			Ordering:              idx + 1,
			ShuffledOptionMapping: mappingJSON,
			IsFlagged:             false,
		})
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&session).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(&sessionQuestions, 100).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to create practice session.",
		})
		return
	}

	dtoQuestions := h.buildActiveQuestionDTOs(userIDStr, session, sessionQuestions, selectedQuestions, nil)

	c.JSON(http.StatusCreated, gin.H{
		"session":   session,
		"questions": dtoQuestions,
	})
}

func (h *SessionHandler) selectQuestionsForSession(
	userID, mode string,
	topicIDs []string,
	styleFilter, diffFilter string,
	count int,
) ([]models.Question, int, error) {
	// Query candidates without preloading options (M-7 memory & SQL optimization)
	query := h.DB.Model(&models.Question{}).Where("status = ?", models.QuestionStatusApproved)

	// C-2 Integrity: Exclude any questions that are in this user's currently active mock exam
	var activeMockQIDs []string
	h.DB.Table("session_questions").
		Joins("JOIN practice_sessions ON practice_sessions.id = session_questions.session_id").
		Where("practice_sessions.user_id = ? AND practice_sessions.mode = ? AND practice_sessions.status = ?",
			userID, models.ModeMockExam, models.SessionStatusActive).
		Pluck("session_questions.question_id", &activeMockQIDs)
	if len(activeMockQIDs) > 0 {
		query = query.Where("id NOT IN ?", activeMockQIDs)
	}

	if mode == models.ModeBookmarkRevision {
		var bms []models.QuestionBookmark
		h.DB.Where("user_id = ?", userID).Find(&bms)
		if len(bms) > 0 {
			bmIDs := make([]string, 0, len(bms))
			for _, b := range bms {
				bmIDs = append(bmIDs, b.QuestionID)
			}
			query = query.Where("id IN ?", bmIDs)
		} else {
			query = query.Where("topic_id IN ?", topicIDs)
		}
	} else if mode == models.ModeReviewMistakes {
		var failedExposures []models.QuestionExposure
		h.DB.Where("user_id = ? AND last_outcome = ?", userID, false).Find(&failedExposures)
		if len(failedExposures) > 0 {
			failedIDs := make([]string, 0, len(failedExposures))
			for _, exp := range failedExposures {
				failedIDs = append(failedIDs, exp.QuestionID)
			}
			query = query.Where("id IN ?", failedIDs)
		} else {
			query = query.Where("topic_id IN ?", topicIDs)
		}
	} else {
		query = query.Where("topic_id IN ?", topicIDs)
	}

	// Strictly enforce beginner and intermediate questions (0% Advanced across all modes)
	query = query.Where("difficulty != ?", models.DifficultyAdvanced)

	if styleFilter != "" && styleFilter != "ALL" {
		query = query.Where("question_style = ?", styleFilter)
	}
	if diffFilter != "" && diffFilter != "ALL" && diffFilter != models.DifficultyAdvanced {
		query = query.Where("difficulty = ?", diffFilter)
	}

	var candidates []models.Question
	if err := query.Find(&candidates).Error; err != nil {
		return nil, 0, err
	}
	if len(candidates) == 0 {
		return nil, 0, nil
	}

	// Lookup current user's QuestionExposure to prioritize unseen questions
	var exposures []models.QuestionExposure
	h.DB.Where("user_id = ?", userID).Find(&exposures)
	expMap := make(map[string]int, len(exposures))
	for _, e := range exposures {
		expMap[e.QuestionID] = e.ExposureCount
	}

	unseenCount := 0
	for _, c := range candidates {
		if expMap[c.ID] == 0 {
			unseenCount++
		}
	}

	// Shuffle candidates randomly first, then stable-sort by exposure_count ASC
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	sort.SliceStable(candidates, func(i, j int) bool {
		return expMap[candidates[i].ID] < expMap[candidates[j].ID]
	})

	var chosen []models.Question
	if len(candidates) <= count {
		chosen = candidates
	} else if styleFilter == "ALL" || styleFilter == "" {
		byStyle := map[string][]models.Question{
			models.StyleDefinition:    {},
			models.StyleCommandSyntax: {},
			models.StyleComparison:    {},
			models.StyleScenario:      {},
		}
		for _, q := range candidates {
			byStyle[q.QuestionStyle] = append(byStyle[q.QuestionStyle], q)
		}

		// Student-friendly weighted distribution:
		// Practice sessions heavily prioritize Foundational Definitions (55%) and practical Linux/Bash Commands (30%)
		// with gentle Comparisons (10%) and minimal Scenarios (5%)
		targetCounts := map[string]int{
			models.StyleDefinition:    int(math.Max(1, math.Round(float64(count)*0.55))),
			models.StyleCommandSyntax: int(math.Max(1, math.Round(float64(count)*0.30))),
			models.StyleComparison:    int(math.Max(0, math.Round(float64(count)*0.10))),
			models.StyleScenario:      int(math.Max(0, math.Round(float64(count)*0.05))),
		}
		if mode == models.ModeMockExam {
			// Welcoming, learner-friendly Mock Exam distribution:
			// 50% Definitions, 30% Hands-on Commands (Linux, Bash, Git, AWS), 15% Clear Comparisons, 5% Scenarios
			targetCounts[models.StyleDefinition] = int(math.Max(1, math.Round(float64(count)*0.50)))
			targetCounts[models.StyleCommandSyntax] = int(math.Max(1, math.Round(float64(count)*0.30)))
			targetCounts[models.StyleComparison] = int(math.Max(1, math.Round(float64(count)*0.15)))
			targetCounts[models.StyleScenario] = count - targetCounts[models.StyleDefinition] - targetCounts[models.StyleCommandSyntax] - targetCounts[models.StyleComparison]
			if targetCounts[models.StyleScenario] < 0 {
				targetCounts[models.StyleScenario] = 0
			}
		}

		styleOrder := []string{
			models.StyleDefinition,
			models.StyleCommandSyntax,
			models.StyleComparison,
			models.StyleScenario,
		}
		selected := make([]models.Question, 0, count)
		pickedIDs := make(map[string]bool, count)

		// First pass: fulfill weighted targets for each style
		for _, st := range styleOrder {
			target := targetCounts[st]
			bucket := byStyle[st]
			added := 0
			for _, q := range bucket {
				if added >= target || len(selected) >= count {
					break
				}
				if !pickedIDs[q.ID] {
					pickedIDs[q.ID] = true
					selected = append(selected, q)
					added++
				}
			}
		}

		// Second pass: top up any remaining slots
		for _, q := range candidates {
			if len(selected) >= count {
				break
			}
			if !pickedIDs[q.ID] {
				pickedIDs[q.ID] = true
				selected = append(selected, q)
			}
		}
		chosen = selected
	} else {
		chosen = candidates[:count]
	}

	// Pedagogical Progression: In practice sessions, order questions logically from Foundational -> Intermediate -> Advanced
	if mode != models.ModeMockExam {
		diffOrder := map[string]int{
			models.DifficultyFoundational: 0,
			models.DifficultyIntermediate: 1,
			models.DifficultyAdvanced:     2,
		}
		sort.SliceStable(chosen, func(i, j int) bool {
			return diffOrder[chosen[i].Difficulty] < diffOrder[chosen[j].Difficulty]
		})
	}

	// M-7: Preload Options ONLY for the final chosen questions
	chosenIDs := make([]string, 0, len(chosen))
	for _, q := range chosen {
		chosenIDs = append(chosenIDs, q.ID)
	}
	if len(chosenIDs) > 0 {
		var withOptions []models.Question
		h.DB.Preload("Options").Where("id IN ?", chosenIDs).Find(&withOptions)
		optMap := make(map[string][]models.QuestionOption, len(withOptions))
		for _, wo := range withOptions {
			optMap[wo.ID] = wo.Options
		}
		for i := range chosen {
			chosen[i].Options = optMap[chosen[i].ID]
		}
	}

	return chosen, unseenCount, nil
}

func (h *SessionHandler) buildActiveQuestionDTOs(
	userID string,
	session models.PracticeSession,
	sqs []models.SessionQuestion,
	questions []models.Question,
	responses []models.Response,
) []ActiveSessionQuestionDTO {
	qMap := make(map[string]models.Question, len(questions))
	for _, q := range questions {
		qMap[q.ID] = q
	}

	respMap := make(map[string]models.Response, len(responses))
	for _, r := range responses {
		respMap[r.QuestionID] = r
	}

	var bookmarks []models.QuestionBookmark
	h.DB.Where("user_id = ?", userID).Find(&bookmarks)
	bmMap := make(map[string]bool, len(bookmarks))
	for _, b := range bookmarks {
		bmMap[b.QuestionID] = true
	}

	dtos := make([]ActiveSessionQuestionDTO, 0, len(sqs))
	displayKeys := []string{"A", "B", "C", "D"}

	for _, sq := range sqs {
		q, ok := qMap[sq.QuestionID]
		if !ok {
			continue
		}

		displayToCanonical := parseOptionMapping(sq.ShuffledOptionMapping)
		canonicalToDisplay := invertOptionMapping(displayToCanonical)

		canonicalOptMap := make(map[string]models.QuestionOption, 4)
		for _, opt := range q.Options {
			canonicalOptMap[opt.Key] = opt
		}

		userResp, hasResp := respMap[q.ID]
		// In Practice mode, provide answer & explanation upfront for instant 0ms feedback.
		// In Mock Exam mode, withhold answers until the exam is submitted.
		isPracticeImmediate := session.Mode != models.ModeMockExam && session.FeedbackMode == "IMMEDIATE"
		revealDetails := isPracticeImmediate || session.Status == models.SessionStatusSubmitted || hasResp

		shuffledOpts := make([]ShuffledOptionDTO, 0, 4)
		for _, dKey := range displayKeys {
			cKey := displayToCanonical[dKey]
			cOpt := canonicalOptMap[cKey]
			optDTO := ShuffledOptionDTO{
				Key:  dKey,
				Text: cOpt.Text,
			}
			if revealDetails {
				optDTO.DistractorRationale = cOpt.DistractorRationale
			}
			shuffledOpts = append(shuffledOpts, optDTO)
		}

		dto := ActiveSessionQuestionDTO{
			QuestionID:      q.ID,
			Ordering:        sq.Ordering,
			TopicID:         q.TopicID,
			ObjectiveID:     q.ObjectiveID,
			Stem:            q.Stem,
			QuestionStyle:   q.QuestionStyle,
			Difficulty:      q.Difficulty,
			SourceReference: q.SourceReference,
			IsFlagged:       sq.IsFlagged,
			IsBookmarked:    bmMap[q.ID],
			Options:         shuffledOpts,
		}

		if hasResp {
			dto.SelectedOptionKey = userResp.SelectedDisplayKey
			isCorr := userResp.IsCorrect
			dto.IsCorrect = &isCorr
		}

		if revealDetails {
			dto.CorrectOptionKey = canonicalToDisplay[q.CorrectOptionKey]
			dto.Explanation = q.Explanation
		}

		dtos = append(dtos, dto)
	}

	return dtos
}

// ListSessions handles GET /api/practice-sessions (and GET /api/sessions).
func (h *SessionHandler) ListSessions(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	statusFilter := strings.ToUpper(strings.TrimSpace(c.Query("status")))

	query := h.DB.Where("user_id = ?", userID)
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var sessions []models.PracticeSession
	if err := query.Order("created_at DESC").Limit(50).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to load practice sessions.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// CancelSession handles POST /api/practice-sessions/:sessionId/cancel (and DELETE /api/practice-sessions/:sessionId).
func (h *SessionHandler) CancelSession(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	sessionID := c.Param("sessionId")

	var session models.PracticeSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "SESSION_NOT_FOUND",
			"message": "Practice session not found.",
		})
		return
	}

	if session.Status != models.SessionStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "SESSION_NOT_ACTIVE",
			"message": "Only active sessions can be cancelled.",
		})
		return
	}

	session.Status = models.SessionStatusCancelled
	session.UpdatedAt = time.Now().UTC()
	if err := h.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to cancel session.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Practice session cancelled successfully.",
		"sessionId": sessionID,
		"status":    models.SessionStatusCancelled,
	})
}

// CancelAllActiveSessions handles POST /api/practice-sessions/cancel-active.
func (h *SessionHandler) CancelAllActiveSessions(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)

	res := h.DB.Model(&models.PracticeSession{}).
		Where("user_id = ? AND status = ?", userIDStr, models.SessionStatusActive).
		Updates(map[string]interface{}{
			"status":     models.SessionStatusCancelled,
			"updated_at": time.Now().UTC(),
		})

	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to cancel active sessions.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "All active sessions cancelled.",
		"affectedCount": res.RowsAffected,
	})
}

// GetSession handles GET /api/practice-sessions/:sessionId.
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	sessionID := c.Param("sessionId")

	var session models.PracticeSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "SESSION_NOT_FOUND",
			"message": "Practice session not found.",
		})
		return
	}

	// M-8 Auto-finalize expired mock exams
	now := time.Now().UTC()
	if session.Mode == models.ModeMockExam && session.Status == models.SessionStatusActive && session.Deadline != nil {
		if now.After(session.Deadline.Add(15 * time.Second)) {
			if finalized, _, err := h.finalizeSession(sessionID, userIDStr); err == nil && finalized != nil {
				session = *finalized
			}
		}
	}

	var sqs []models.SessionQuestion
	h.DB.Where("session_id = ?", sessionID).Order("ordering ASC").Find(&sqs)

	qIDs := make([]string, 0, len(sqs))
	for _, sq := range sqs {
		qIDs = append(qIDs, sq.QuestionID)
	}

	var questions []models.Question
	if len(qIDs) > 0 {
		h.DB.Preload("Options").Where("id IN ?", qIDs).Find(&questions)
	}

	var responses []models.Response
	h.DB.Where("session_id = ? AND user_id = ?", sessionID, userIDStr).Find(&responses)

	dtoQuestions := h.buildActiveQuestionDTOs(userIDStr, session, sqs, questions, responses)

	c.JSON(http.StatusOK, gin.H{
		"session":    session,
		"questions":  dtoQuestions,
		"serverTime": time.Now().UTC().Format(time.RFC3339),
	})
}

// UpsertResponse handles PUT /api/practice-sessions/:sessionId/responses/:questionId.
func (h *SessionHandler) UpsertResponse(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	sessionID := c.Param("sessionId")
	questionID := c.Param("questionId")

	var req submitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Please provide selectedOptionKey (A, B, C, or D).",
		})
		return
	}

	selectedDisplayKey := strings.ToUpper(strings.TrimSpace(req.SelectedOptionKey))
	if selectedDisplayKey != "A" && selectedDisplayKey != "B" && selectedDisplayKey != "C" && selectedDisplayKey != "D" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_OPTION_KEY",
			"message": "selectedOptionKey must be A, B, C, or D.",
		})
		return
	}

	var session models.PracticeSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "SESSION_NOT_FOUND",
			"message": "Practice session not found.",
		})
		return
	}

	if session.Status != models.SessionStatusActive {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "SESSION_ALREADY_SUBMITTED",
			"message": "This session has already been submitted and cannot be modified.",
		})
		return
	}

	now := time.Now().UTC()
	if session.Deadline != nil && now.After(session.Deadline.Add(15*time.Second)) {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "SESSION_DEADLINE_EXPIRED",
			"message": "The time limit for this exam has expired. Please submit your session.",
		})
		return
	}

	// H-4 Integrity: In Immediate feedback mode, responses are locked after initial answer submission
	if session.FeedbackMode == "IMMEDIATE" {
		var existingResp models.Response
		if err := h.DB.Where("session_id = ? AND question_id = ?", sessionID, questionID).First(&existingResp).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "ANSWER_LOCKED",
				"message": "Immediate feedback question has already been answered and cannot be modified.",
			})
			return
		}
	}

	var sq models.SessionQuestion
	if err := h.DB.Where("session_id = ? AND question_id = ?", sessionID, questionID).First(&sq).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "QUESTION_NOT_IN_SESSION",
			"message": "Question does not belong to this session.",
		})
		return
	}

	var question models.Question
	if err := h.DB.Preload("Options").Where("id = ?", questionID).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "QUESTION_NOT_FOUND",
			"message": "Question not found.",
		})
		return
	}

	displayToCanonical := parseOptionMapping(sq.ShuffledOptionMapping)
	canonicalToDisplay := invertOptionMapping(displayToCanonical)
	canonicalSelectedKey := displayToCanonical[selectedDisplayKey]
	isCorrect := canonicalSelectedKey == question.CorrectOptionKey

	respRecord := models.Response{
		ID:                 fmt.Sprintf("resp-%s-%s", sessionID, questionID),
		SessionID:          sessionID,
		UserID:             userIDStr,
		QuestionID:         questionID,
		SelectedDisplayKey: selectedDisplayKey,
		CanonicalOptionKey: canonicalSelectedKey,
		IsCorrect:          isCorrect,
		AnsweredAt:         now,
	}

	if err := h.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}, {Name: "question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"selected_display_key", "canonical_option_key", "is_correct", "answered_at"}),
	}).Create(&respRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to save response.",
		})
		return
	}

	// Strictly conceal correctness and rationales on Mock Exams or Deferred feedback sessions
	if session.Mode == models.ModeMockExam || session.FeedbackMode == "DEFERRED" {
		c.JSON(http.StatusOK, gin.H{
			"saved":             true,
			"questionId":        questionID,
			"selectedOptionKey": selectedDisplayKey,
			"answeredAt":        now.Format(time.RFC3339),
		})
		return
	}

	// Immediate Feedback Mode (Practice): return display-mapped rationales and explanation
	distractorRationales := make(map[string]string, 4)
	for _, opt := range question.Options {
		dKey := canonicalToDisplay[opt.Key]
		if opt.DistractorRationale != "" {
			distractorRationales[dKey] = opt.DistractorRationale
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"saved":                true,
		"questionId":           questionID,
		"selectedOptionKey":    selectedDisplayKey,
		"isCorrect":            isCorrect,
		"correctOptionKey":     canonicalToDisplay[question.CorrectOptionKey],
		"explanation":          question.Explanation,
		"distractorRationales": distractorRationales,
		"sourceReference":      question.SourceReference,
		"answeredAt":           now.Format(time.RFC3339),
	})
}

// ToggleFlag handles PUT /api/practice-sessions/:sessionId/questions/:questionId/flag.
func (h *SessionHandler) ToggleFlag(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	sessionID := c.Param("sessionId")
	questionID := c.Param("questionId")

	var session models.PracticeSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "SESSION_NOT_FOUND",
			"message": "Session not found.",
		})
		return
	}

	// M-8: Disallow flagging on non-ACTIVE sessions
	if session.Status != models.SessionStatusActive {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "SESSION_NOT_ACTIVE",
			"message": "Flagging is only permitted on active sessions.",
		})
		return
	}

	var sq models.SessionQuestion
	if err := h.DB.Where("session_id = ? AND question_id = ?", sessionID, questionID).First(&sq).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "QUESTION_NOT_IN_SESSION",
			"message": "Question not in session.",
		})
		return
	}

	var req flagQuestionRequest
	_ = c.ShouldBindJSON(&req)
	newFlag := !sq.IsFlagged
	if req.IsFlagged != nil {
		newFlag = *req.IsFlagged
	}

	h.DB.Model(&sq).Update("is_flagged", newFlag)

	c.JSON(http.StatusOK, gin.H{
		"questionId": questionID,
		"isFlagged":  newFlag,
	})
}

// finalizeSession handles atomic transition of a session from ACTIVE -> SUBMITTED (H-1).
func (h *SessionHandler) finalizeSession(sessionID, userIDStr string) (*models.PracticeSession, bool, error) {
	var session models.PracticeSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&session).Error; err != nil {
		return nil, false, err
	}

	if session.Status == models.SessionStatusSubmitted {
		return &session, true, nil
	}

	var sqs []models.SessionQuestion
	h.DB.Where("session_id = ?", sessionID).Order("ordering ASC").Find(&sqs)

	var responses []models.Response
	h.DB.Where("session_id = ? AND user_id = ?", sessionID, userIDStr).Find(&responses)
	respMap := make(map[string]models.Response, len(responses))
	for _, r := range responses {
		respMap[r.QuestionID] = r
	}

	correctCount := 0
	incorrectCount := 0
	unansweredCount := 0
	now := time.Now().UTC()

	for _, sq := range sqs {
		r, answered := respMap[sq.QuestionID]
		if !answered {
			unansweredCount++
		} else if r.IsCorrect {
			correctCount++
		} else {
			incorrectCount++
		}
	}

	totalQ := len(sqs)
	scorePercent := 0.0
	if totalQ > 0 {
		scorePercent = math.Round((float64(correctCount)/float64(totalQ))*10000) / 100
	}

	alreadySubmitted := false
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Atomic conditional update ensures only ONE concurrent request can transition ACTIVE -> SUBMITTED
		res := tx.Model(&models.PracticeSession{}).
			Where("id = ? AND user_id = ? AND status = ?", sessionID, userIDStr, models.SessionStatusActive).
			Updates(map[string]any{
				"status":           models.SessionStatusSubmitted,
				"submitted_at":     &now,
				"score_percent":    scorePercent,
				"correct_count":    correctCount,
				"incorrect_count":  incorrectCount,
				"unanswered_count": unansweredCount,
				"updated_at":       now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			alreadySubmitted = true
			return nil
		}

		session.Status = models.SessionStatusSubmitted
		session.SubmittedAt = &now
		session.ScorePercent = scorePercent
		session.ScorePercentage = scorePercent
		session.CorrectCount = correctCount
		session.IncorrectCount = incorrectCount
		session.UnansweredCount = unansweredCount
		session.UpdatedAt = now

		// Update QuestionExposure for all questions in the submitted session
		for _, sq := range sqs {
			r, answered := respMap[sq.QuestionID]
			isCorr := answered && r.IsCorrect
			correctInc := 0
			if isCorr {
				correctInc = 1
			}

			var existingExp models.QuestionExposure
			err := tx.Where("user_id = ? AND question_id = ?", userIDStr, sq.QuestionID).First(&existingExp).Error
			if err != nil {
				newExp := models.QuestionExposure{
					ID:            fmt.Sprintf("exp-%s-%s", userIDStr, sq.QuestionID),
					UserID:        userIDStr,
					QuestionID:    sq.QuestionID,
					ExposureCount: 1,
					CorrectCount:  correctInc,
					LastOutcome:   isCorr,
					LastSeenAt:    now,
				}
				if err := tx.Create(&newExp).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&existingExp).Updates(map[string]any{
					"exposure_count": existingExp.ExposureCount + 1,
					"correct_count":  existingExp.CorrectCount + correctInc,
					"last_outcome":   isCorr,
					"last_seen_at":   now,
				}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, false, err
	}

	if alreadySubmitted {
		var existing models.PracticeSession
		h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&existing)
		existing.ScorePercentage = existing.ScorePercent
		return &existing, true, nil
	}

	return &session, false, nil
}

// SubmitSession handles POST /api/practice-sessions/:sessionId/submit (Transactional & Idempotent).
func (h *SessionHandler) SubmitSession(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	sessionID := c.Param("sessionId")

	session, alreadySubmitted, err := h.finalizeSession(sessionID, userIDStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "SESSION_NOT_FOUND",
				"message": "Practice session not found.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to finalize session submission.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alreadySubmitted": alreadySubmitted,
		"session":          session,
	})
}

// GetSessionReview handles GET /api/practice-sessions/:sessionId/review.
func (h *SessionHandler) GetSessionReview(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	userIDStr := fmt.Sprintf("%v", userID)
	sessionID := c.Param("sessionId")

	var session models.PracticeSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userIDStr).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "SESSION_NOT_FOUND",
			"message": "Session not found.",
		})
		return
	}

	if session.Status != models.SessionStatusSubmitted {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "SESSION_NOT_SUBMITTED",
			"message": "Full exam review and explanations are only available after submitting the session.",
		})
		return
	}
	session.ScorePercentage = session.ScorePercent

	var sqs []models.SessionQuestion
	h.DB.Where("session_id = ?", sessionID).Order("ordering ASC").Find(&sqs)

	qIDs := make([]string, 0, len(sqs))
	for _, sq := range sqs {
		qIDs = append(qIDs, sq.QuestionID)
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

	var responses []models.Response
	h.DB.Where("session_id = ? AND user_id = ?", sessionID, userIDStr).Find(&responses)
	respMap := make(map[string]models.Response, len(responses))
	for _, r := range responses {
		respMap[r.QuestionID] = r
	}

	var bookmarks []models.QuestionBookmark
	h.DB.Where("user_id = ?", userIDStr).Find(&bookmarks)
	bmMap := make(map[string]bool, len(bookmarks))
	for _, b := range bookmarks {
		bmMap[b.QuestionID] = true
	}

	type breakdownStat struct {
		Key       string  `json:"key"`
		Label     string  `json:"label"`
		Total     int     `json:"total"`
		Correct   int     `json:"correct"`
		Accuracy  float64 `json:"accuracyPercent"`
	}
	topicStats := make(map[string]*breakdownStat)
	styleStats := make(map[string]*breakdownStat)

	reviewedQuestions := make([]ReviewedQuestionDTO, 0, len(sqs))
	displayKeys := []string{"A", "B", "C", "D"}

	for _, sq := range sqs {
		q, ok := qMap[sq.QuestionID]
		if !ok {
			continue
		}
		topic := topicMap[q.TopicID]
		displayToCanonical := parseOptionMapping(sq.ShuffledOptionMapping)
		canonicalToDisplay := invertOptionMapping(displayToCanonical)

		canonicalOptMap := make(map[string]models.QuestionOption, 4)
		for _, opt := range q.Options {
			canonicalOptMap[opt.Key] = opt
		}

		shuffledOpts := make([]ShuffledOptionDTO, 0, 4)
		for _, dKey := range displayKeys {
			cKey := displayToCanonical[dKey]
			cOpt := canonicalOptMap[cKey]
			shuffledOpts = append(shuffledOpts, ShuffledOptionDTO{
				Key:                 dKey,
				Text:                cOpt.Text,
				DistractorRationale: cOpt.DistractorRationale,
			})
		}

		userResp, isAnswered := respMap[q.ID]
		isCorr := isAnswered && userResp.IsCorrect
		correctDisplayKey := canonicalToDisplay[q.CorrectOptionKey]

		reviewedQuestions = append(reviewedQuestions, ReviewedQuestionDTO{
			QuestionID:        q.ID,
			Ordering:          sq.Ordering,
			TopicID:           q.TopicID,
			TopicTitle:        topic.Title,
			ModuleNumber:      topic.ModuleNumber,
			WeekCode:          topic.WeekCode,
			ObjectiveID:       q.ObjectiveID,
			Stem:              q.Stem,
			QuestionStyle:     q.QuestionStyle,
			Difficulty:        q.Difficulty,
			SourceType:        q.SourceType,
			SourceReference:   q.SourceReference,
			IsFlagged:         sq.IsFlagged,
			IsBookmarked:      bmMap[q.ID],
			SelectedOptionKey: userResp.SelectedDisplayKey,
			CorrectOptionKey:  correctDisplayKey,
			IsCorrect:         isCorr,
			IsAnswered:        isAnswered,
			Explanation:       q.Explanation,
			Options:           shuffledOpts,
		})

		if _, exists := topicStats[q.TopicID]; !exists {
			topicStats[q.TopicID] = &breakdownStat{Key: q.TopicID, Label: topic.Title}
		}
		topicStats[q.TopicID].Total++
		if isCorr {
			topicStats[q.TopicID].Correct++
		}

		if _, exists := styleStats[q.QuestionStyle]; !exists {
			styleStats[q.QuestionStyle] = &breakdownStat{Key: q.QuestionStyle, Label: q.QuestionStyle}
		}
		styleStats[q.QuestionStyle].Total++
		if isCorr {
			styleStats[q.QuestionStyle].Correct++
		}
	}

	topicBreakdown := make([]breakdownStat, 0, len(topicStats))
	for _, st := range topicStats {
		if st.Total > 0 {
			st.Accuracy = math.Round((float64(st.Correct)/float64(st.Total))*10000) / 100
		}
		topicBreakdown = append(topicBreakdown, *st)
	}

	styleBreakdown := make([]breakdownStat, 0, len(styleStats))
	for _, st := range styleStats {
		if st.Total > 0 {
			st.Accuracy = math.Round((float64(st.Correct)/float64(st.Total))*10000) / 100
		}
		styleBreakdown = append(styleBreakdown, *st)
	}

	c.JSON(http.StatusOK, gin.H{
		"session":        session,
		"questions":      reviewedQuestions,
		"topicBreakdown": topicBreakdown,
		"styleBreakdown": styleBreakdown,
	})
}

