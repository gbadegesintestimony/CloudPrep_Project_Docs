package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"cloudprep-backend/internal/curriculum"
	"cloudprep-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MultiProviderAiRouter struct {
	DB               *gorm.DB
	Primary          QuestionGenerationProvider
	Fallback         QuestionGenerationProvider
	CurriculumTopics []models.CurriculumTopic
}

func NewAiRouter(
	database *gorm.DB,
	primary, fallback QuestionGenerationProvider,
	topics []models.CurriculumTopic,
) *MultiProviderAiRouter {
	if len(topics) == 0 {
		topics = curriculum.GetCurriculumTopics()
	}
	return &MultiProviderAiRouter{
		DB:               database,
		Primary:          primary,
		Fallback:         fallback,
		CurriculumTopics: topics,
	}
}

func (r *MultiProviderAiRouter) findTopic(topicID string) (models.CurriculumTopic, bool) {
	for _, t := range r.CurriculumTopics {
		if t.ID == topicID || t.Slug == topicID {
			return t, true
		}
	}
	return models.CurriculumTopic{}, false
}

// GenerateQuestions routes the request to Primary and automatically fails over to Fallback on error.
func (r *MultiProviderAiRouter) GenerateQuestions(
	ctx context.Context,
	req GenerationRequest,
) (*GenerationResult, error) {
	topic, found := r.findTopic(req.TopicID)
	if !found {
		return nil, fmt.Errorf("curriculum topic %q not found", req.TopicID)
	}

	// Extract course excerpts from the topic's learning objectives
	excerpts := make([]string, 0, len(topic.Objectives))
	for _, obj := range topic.Objectives {
		if obj.CurriculumExcerpt != "" {
			excerpts = append(excerpts, fmt.Sprintf("[%s: %s]\n%s", obj.Code, obj.Title, obj.CurriculumExcerpt))
		}
	}

	// Fetch recent question stems for this topic to avoid duplicates
	if len(req.AvoidStems) == 0 && r.DB != nil {
		var recentStems []string
		r.DB.Model(&models.Question{}).
			Where("topic_id = ?", topic.ID).
			Order("created_at DESC").
			Limit(10).
			Pluck("stem", &recentStems)
		req.AvoidStems = recentStems
	}

	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 10 {
		req.Count = 10
	}

	var primaryErr error
	if r.Primary != nil && r.Primary.IsConfigured() {
		result, err := r.Primary.GenerateQuestions(ctx, req, topic, excerpts)
		if err == nil && len(result.Questions) > 0 {
			result.FailoverUsed = false
			return result, nil
		}
		primaryErr = err
		log.Printf("[ai-router] Primary provider (%s) failed: %v. Initiating failover to secondary provider...", r.Primary.ProviderName(), err)
	} else {
		primaryErr = ErrUnconfigured
	}

	// Automatic Failover to secondary provider
	if r.Fallback != nil && r.Fallback.IsConfigured() {
		result, err := r.Fallback.GenerateQuestions(ctx, req, topic, excerpts)
		if err == nil && len(result.Questions) > 0 {
			result.FailoverUsed = true
			log.Printf("[ai-router] Fallback provider (%s) succeeded in generating %d questions (FailoverUsed=true).", r.Fallback.ProviderName(), len(result.Questions))
			return result, nil
		}
		return nil, fmt.Errorf("both AI providers failed: primary error (%v); fallback error (%w)", primaryErr, err)
	}

	return nil, fmt.Errorf("primary AI provider failed (%v) and fallback provider is not configured or available", primaryErr)
}

// GenerateAndPersist generates questions via the dual-AI pipeline, validates quality/duplicates,
// saves accepted questions to the database, and records an audit GenerationRecord.
func (r *MultiProviderAiRouter) GenerateAndPersist(
	ctx context.Context,
	req GenerationRequest,
) ([]models.Question, *models.GenerationRecord, error) {
	topic, found := r.findTopic(req.TopicID)
	if !found {
		return nil, nil, fmt.Errorf("curriculum topic %q not found", req.TopicID)
	}

	result, err := r.GenerateQuestions(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Fetch existing fingerprints for this topic to reject near-duplicates
	existingFPMap := make(map[string]bool)
	if r.DB != nil {
		var fps []string
		r.DB.Model(&models.Question{}).Where("topic_id = ?", topic.ID).Pluck("content_fingerprint", &fps)
		for _, fp := range fps {
			existingFPMap[fp] = true
		}
	}

	accepted := make([]models.Question, 0, len(result.Questions))
	var rejectionLogs []string
	now := time.Now().UTC()

	sourceType := models.SourceAIGemini
	if result.Provider == "groq" {
		sourceType = models.SourceAIGroq
	}

	// Default objective fallback if not specified
	defaultObjID := fmt.Sprintf("%s-obj-1", topic.ID)
	if len(topic.Objectives) > 0 {
		defaultObjID = topic.Objectives[0].ID
	}

	for _, gq := range result.Questions {
		val := ValidateQuestion(&gq, topic.ID, req.QuestionStyle, existingFPMap)
		if !val.IsValid {
			rejectionLogs = append(rejectionLogs, val.RejectionReason)
			continue
		}

		// Register fingerprint in local map so duplicates within the same batch are also caught
		existingFPMap[val.ContentFingerprint] = true

		qShortID := generateRandomHex(4)
		stylePrefix := strings.ToLower(gq.QuestionStyle)
		if len(stylePrefix) > 3 {
			stylePrefix = stylePrefix[:3]
		}
		qID := fmt.Sprintf("q-%s-%s-ai-%s", topic.ID, stylePrefix, qShortID)

		// Construct canonical QuestionOptions
		opts := make([]models.QuestionOption, 0, 4)
		for order, opt := range gq.Options {
			opts = append(opts, models.QuestionOption{
				ID:                  fmt.Sprintf("%s-opt-%s", qID, opt.Key),
				QuestionID:          qID,
				Key:                 opt.Key,
				Text:                opt.Text,
				DistractorRationale: opt.DistractorRationale,
				Ordering:            order + 1,
			})
		}

		srcRef := gq.SourceReference
		if srcRef == "" {
			srcRef = fmt.Sprintf("%s, %s", topic.ModuleTitle, topic.WeekCode)
		}

		objID := req.ObjectiveID
		if objID == "" {
			objID = defaultObjID
		}

		questionModel := models.Question{
			ID:                 qID,
			TopicID:            topic.ID,
			ObjectiveID:        objID,
			Stem:               gq.Stem,
			QuestionStyle:      gq.QuestionStyle,
			Difficulty:         gq.Difficulty,
			CognitiveSkill:     gq.CognitiveSkill,
			ScenarioType:       gq.ScenarioType,
			CorrectOptionKey:   gq.CorrectOptionKey,
			Explanation:        gq.Explanation,
			SourceType:         sourceType,
			SourceReference:    srcRef,
			ContentFingerprint: val.ContentFingerprint,
			Status:             models.QuestionStatusApproved,
			Version:            1,
			Options:            opts,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		accepted = append(accepted, questionModel)
	}

	// Persist accepted questions into the Question table
	if len(accepted) > 0 && r.DB != nil {
		err := r.DB.Transaction(func(tx *gorm.DB) error {
			for _, q := range accepted {
				qCopy := q
				qOpts := q.Options
				qCopy.Options = nil

				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&qCopy).Error; err != nil {
					return err
				}
				if len(qOpts) > 0 {
					if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&qOpts).Error; err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("[ai-router] Error saving accepted questions to database: %v", err)
		}
	}

	// Create GenerationRecord audit trail
	genRecord := models.GenerationRecord{
		ID:                fmt.Sprintf("gen-%s", generateRandomHex(8)),
		Provider:          result.Provider,
		ModelID:           result.ModelID,
		FailoverUsed:      result.FailoverUsed,
		TopicID:           topic.ID,
		QuestionStyle:     req.QuestionStyle,
		RequestedCount:    req.Count,
		AcceptedCount:     len(accepted),
		RejectedCount:     len(result.Questions) - len(accepted),
		ValidationSummary: strings.Join(rejectionLogs, "; "),
		DurationMs:        result.DurationMs,
		CreatedAt:         now,
	}

	if r.DB != nil {
		_ = r.DB.Create(&genRecord)
	}

	return accepted, &genRecord, nil
}

func generateRandomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

