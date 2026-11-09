package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloudprep-backend/internal/curriculum"
	"cloudprep-backend/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// mockProvider implements QuestionGenerationProvider for deterministic unit testing.
type mockProvider struct {
	name       string
	model      string
	configured bool
	shouldFail bool
	errToFail  error
	questions  []GeneratedQuestion
}

func (m *mockProvider) ProviderName() string { return m.name }
func (m *mockProvider) ModelID() string      { return m.model }
func (m *mockProvider) IsConfigured() bool   { return m.configured }
func (m *mockProvider) GenerateQuestions(ctx context.Context, req GenerationRequest, topic models.CurriculumTopic, excerpts []string) (*GenerationResult, error) {
	if !m.configured {
		return nil, ErrUnconfigured
	}
	if m.shouldFail {
		if m.errToFail != nil {
			return nil, m.errToFail
		}
		return nil, errors.New("simulated provider network timeout")
	}
	return &GenerationResult{
		Provider:     m.name,
		ModelID:      m.model,
		FailoverUsed: false,
		Questions:    m.questions,
		DurationMs:   120,
	}, nil
}

func sampleValidQuestion() GeneratedQuestion {
	return GeneratedQuestion{
		Stem:             "In a multi-tier web application on AWS, which subnet topology is required to secure backend database instances from public ingress?",
		QuestionStyle:    models.StyleScenario,
		Difficulty:       models.DifficultyIntermediate,
		CognitiveSkill:   "APPLICATION",
		ScenarioType:     "vpc-security",
		CorrectOptionKey: "B",
		Explanation:      "Database instances must reside in private subnets with no route to an Internet Gateway; ingress is strictly limited to application tier security groups.",
		SourceReference:  "Module 2: Cloud Architecture, Services & AI, M2_W2",
		Options: []GeneratedOption{
			{Key: "A", Text: "Deploy databases in public subnets with 0.0.0.0/0 allowed on port 5432.", DistractorRationale: "Exposes database directly to public internet."},
			{Key: "B", Text: "Deploy databases in isolated private subnets with ingress restricted to application tier security groups.", DistractorRationale: ""},
			{Key: "C", Text: "Deploy databases across an unpeered VPC without route tables.", DistractorRationale: "Prevents legitimate application tier communication."},
			{Key: "D", Text: "Attach an Elastic IP directly to each database instance.", DistractorRationale: "Directly contravenes private isolation requirements."},
		},
	}
}

func newTestAiDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}
	err = db.AutoMigrate(
		&models.CurriculumTopic{},
		&models.LearningObjective{},
		&models.Question{},
		&models.QuestionOption{},
		&models.GenerationRecord{},
	)
	if err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}
	return db
}

func TestPromptAssembly(t *testing.T) {
	topics := curriculum.GetCurriculumTopics()
	topic := topics[0]

	req := GenerationRequest{
		TopicID:       topic.ID,
		QuestionStyle: models.StyleDefinition,
		Difficulty:    models.DifficultyFoundational,
		Count:         2,
		AvoidStems:    []string{"What is virtualization?"},
	}

	excerpts := []string{"Virtualization abstracts hardware into software instances."}
	prompt := BuildUserPrompt(req, topic, excerpts)

	if !containsString(prompt, topic.Title) {
		t.Errorf("expected prompt to contain topic title %s", topic.Title)
	}
	if !containsString(prompt, "Virtualization abstracts hardware") {
		t.Errorf("expected prompt to contain curriculum excerpt")
	}
	if !containsString(prompt, "What is virtualization?") {
		t.Errorf("expected prompt to contain avoidStems")
	}
}

func TestValidator_ValidQuestion(t *testing.T) {
	q := sampleValidQuestion()
	res := ValidateQuestion(&q, "topic-6-networking", models.StyleScenario, nil)
	if !res.IsValid {
		t.Fatalf("expected question to be valid, got rejection: %s", res.RejectionReason)
	}
	if res.ContentFingerprint == "" {
		t.Fatal("expected non-empty content fingerprint")
	}
}

func TestValidator_InvalidOptions(t *testing.T) {
	// Fewer than 4 options
	qShort := sampleValidQuestion()
	qShort.Options = qShort.Options[:3]
	resShort := ValidateQuestion(&qShort, "topic-6-networking", models.StyleScenario, nil)
	if resShort.IsValid {
		t.Fatal("expected rejection for < 4 options")
	}

	// Duplicate option keys
	qDupKey := sampleValidQuestion()
	qDupKey.Options[1].Key = "A"
	resDupKey := ValidateQuestion(&qDupKey, "topic-6-networking", models.StyleScenario, nil)
	if resDupKey.IsValid {
		t.Fatal("expected rejection for duplicate option keys")
	}

	// Duplicate option texts
	qDupText := sampleValidQuestion()
	qDupText.Options[2].Text = qDupText.Options[0].Text
	resDupText := ValidateQuestion(&qDupText, "topic-6-networking", models.StyleScenario, nil)
	if resDupText.IsValid {
		t.Fatal("expected rejection for duplicate option texts")
	}
}

func TestValidator_DuplicateFingerprint(t *testing.T) {
	q := sampleValidQuestion()
	fp := ComputeFingerprint("topic-6-networking", models.StyleScenario, q.Stem)
	existing := map[string]bool{fp: true}

	res := ValidateQuestion(&q, "topic-6-networking", models.StyleScenario, existing)
	if res.IsValid {
		t.Fatal("expected rejection for duplicate content fingerprint")
	}
	if !containsString(res.RejectionReason, "Duplicate question detected") {
		t.Fatalf("unexpected rejection reason: %s", res.RejectionReason)
	}
}

func TestMultiProviderAiRouter_PrimarySuccess(t *testing.T) {
	db := newTestAiDB(t)
	q := sampleValidQuestion()

	primary := &mockProvider{
		name:       "gemini",
		model:      "gemini-3.8-flash",
		configured: true,
		questions:  []GeneratedQuestion{q},
	}
	fallback := &mockProvider{
		name:       "groq",
		model:      "openai/gpt-oss-120b",
		configured: true,
		questions:  []GeneratedQuestion{q},
	}

	router := NewAiRouter(db, primary, fallback, curriculum.GetCurriculumTopics())
	result, err := router.GenerateQuestions(context.Background(), GenerationRequest{
		TopicID:       "topic-6-networking",
		QuestionStyle: models.StyleScenario,
		Count:         1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provider != "gemini" {
		t.Errorf("expected provider gemini, got %s", result.Provider)
	}
	if result.FailoverUsed {
		t.Errorf("expected failoverUsed=false when primary succeeds")
	}
}

func TestMultiProviderAiRouter_FailoverToGroq(t *testing.T) {
	db := newTestAiDB(t)
	q := sampleValidQuestion()

	// Primary fails (e.g. rate limit HTTP 429)
	primary := &mockProvider{
		name:       "gemini",
		model:      "gemini-3.8-flash",
		configured: true,
		shouldFail: true,
		errToFail:  ErrRateLimited,
	}
	// Fallback succeeds
	fallback := &mockProvider{
		name:       "groq",
		model:      "openai/gpt-oss-120b",
		configured: true,
		questions:  []GeneratedQuestion{q},
	}

	router := NewAiRouter(db, primary, fallback, curriculum.GetCurriculumTopics())
	result, err := router.GenerateQuestions(context.Background(), GenerationRequest{
		TopicID:       "topic-6-networking",
		QuestionStyle: models.StyleScenario,
		Count:         1,
	})
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}
	if result.Provider != "groq" {
		t.Errorf("expected provider groq after failover, got %s", result.Provider)
	}
	if !result.FailoverUsed {
		t.Errorf("expected failoverUsed=true when secondary was used")
	}
}

func TestMultiProviderAiRouter_BothFailGraceful(t *testing.T) {
	db := newTestAiDB(t)

	primary := &mockProvider{name: "gemini", configured: true, shouldFail: true}
	fallback := &mockProvider{name: "groq", configured: true, shouldFail: true}

	router := NewAiRouter(db, primary, fallback, curriculum.GetCurriculumTopics())
	result, err := router.GenerateQuestions(context.Background(), GenerationRequest{
		TopicID:       "topic-6-networking",
		QuestionStyle: models.StyleScenario,
		Count:         1,
	})
	if err == nil {
		t.Fatal("expected error when both AI providers fail")
	}
	if result != nil {
		t.Fatal("expected nil result on total provider failure")
	}
}

func TestGenerateAndPersist_BankGrowth(t *testing.T) {
	db := newTestAiDB(t)
	q := sampleValidQuestion()

	primary := &mockProvider{
		name:       "gemini",
		model:      "gemini-3.8-flash",
		configured: true,
		questions:  []GeneratedQuestion{q},
	}

	router := NewAiRouter(db, primary, nil, curriculum.GetCurriculumTopics())
	accepted, record, err := router.GenerateAndPersist(context.Background(), GenerationRequest{
		TopicID:       "topic-6-networking",
		QuestionStyle: models.StyleScenario,
		Count:         1,
	})
	if err != nil {
		t.Fatalf("unexpected error in GenerateAndPersist: %v", err)
	}
	if len(accepted) != 1 {
		t.Fatalf("expected 1 accepted question, got %d", len(accepted))
	}
	if record == nil {
		t.Fatal("expected non-nil GenerationRecord")
	}
	if record.AcceptedCount != 1 {
		t.Errorf("expected record.AcceptedCount=1, got %d", record.AcceptedCount)
	}

	// Verify question exists in SQLite
	var dbQ models.Question
	if err := db.Preload("Options").Where("id = ?", accepted[0].ID).First(&dbQ).Error; err != nil {
		t.Fatalf("failed to query persisted question: %v", err)
	}
	if dbQ.SourceType != models.SourceAIGemini {
		t.Errorf("expected sourceType=AI_GEMINI, got %s", dbQ.SourceType)
	}
	if len(dbQ.Options) != 4 {
		t.Errorf("expected 4 options in database, got %d", len(dbQ.Options))
	}
}

func TestReplenishmentWorker(t *testing.T) {
	db := newTestAiDB(t)
	q := sampleValidQuestion()

	primary := &mockProvider{
		name:       "gemini",
		model:      "gemini-3.8-flash",
		configured: true,
		questions:  []GeneratedQuestion{q},
	}

	router := NewAiRouter(db, primary, nil, curriculum.GetCurriculumTopics())
	worker := NewReplenishmentWorker(router)

	worker.TriggerReplenishment("topic-6-networking", models.StyleScenario, models.DifficultyIntermediate)
	time.Sleep(100 * time.Millisecond)

	var count int64
	db.Model(&models.Question{}).Where("topic_id = ?", "topic-6-networking").Count(&count)
	if count != 1 {
		t.Fatalf("expected background replenishment to persist 1 question, got %d", count)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr))))
}

