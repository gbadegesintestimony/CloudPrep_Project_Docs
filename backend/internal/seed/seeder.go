package seed

import (
	"fmt"
	"log"
	"time"

	"cloudprep-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetAllTopicBlueprints returns all 10 curriculum topic blueprints across Module 1 and Module 2.
func GetAllTopicBlueprints() []TopicBlueprint {
	all := make([]TopicBlueprint, 0, 10)
	all = append(all, GetModule1TopicsA()...)
	all = append(all, GetModule1TopicsB()...)
	all = append(all, GetModule2TopicsA()...)
	all = append(all, GetModule2TopicsB()...)
	all = append(all, GetModule2TopicsC()...)
	return all
}

// BuildFullQuestionBank constructs all 1,040 curriculum-grounded questions (104 per topic across all 10 topics).
func BuildFullQuestionBank() []models.Question {
	now := time.Now().UTC()
	blueprints := GetAllTopicBlueprints()
	allQuestions := make([]models.Question, 0, len(blueprints)*104)
	for _, bp := range blueprints {
		topicQuestions := BuildTopicQuestions(bp, now)
		allQuestions = append(allQuestions, topicQuestions...)
	}
	return allQuestions
}

// SeedQuestionBank idempotently seeds the 1,040 approved questions and 4,160 options into the database.
func SeedQuestionBank(database *gorm.DB) error {
	var existingCount int64
	if err := database.Model(&models.Question{}).Where("source_type = ?", models.SourceSeedBank).Count(&existingCount).Error; err != nil {
		return fmt.Errorf("failed counting existing seed bank questions: %w", err)
	}

	questions := BuildFullQuestionBank()

	// Separate options for fast batch insertion without nested association overhead
	bareQuestions := make([]models.Question, 0, len(questions))
	allOptions := make([]models.QuestionOption, 0, len(questions)*4)

	for _, q := range questions {
		allOptions = append(allOptions, q.Options...)
		qCopy := q
		qCopy.Options = nil
		bareQuestions = append(bareQuestions, qCopy)
	}

	err := database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"stem", "explanation", "difficulty", "cognitive_skill", "correct_option_key", "content_fingerprint", "updated_at"}),
		}).CreateInBatches(&bareQuestions, 100).Error; err != nil {
			return fmt.Errorf("failed batch upserting questions: %w", err)
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"text", "distractor_rationale", "ordering"}),
		}).CreateInBatches(&allOptions, 200).Error; err != nil {
			return fmt.Errorf("failed batch inserting question options: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var finalCount int64
	database.Model(&models.Question{}).Where("status = ?", models.QuestionStatusApproved).Count(&finalCount)
	log.Printf("[seed] Question Bank seeded: %d approved questions (%d options) across 10 topics", finalCount, len(allOptions))
	return nil
}

