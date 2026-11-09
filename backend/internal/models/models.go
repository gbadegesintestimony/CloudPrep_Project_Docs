package models

import "time"

// Role constants
const (
	RoleLearner = "LEARNER"
	RoleGuest   = "GUEST"
	RoleAdmin   = "ADMIN"
)

// QuestionStyle constants
const (
	StyleDefinition    = "DEFINITION"
	StyleCommandSyntax = "COMMAND_SYNTAX"
	StyleComparison    = "COMPARISON"
	StyleScenario      = "SCENARIO"
)

// Difficulty constants
const (
	DifficultyFoundational = "FOUNDATIONAL"
	DifficultyIntermediate = "INTERMEDIATE"
	DifficultyAdvanced     = "ADVANCED"
)

// QuestionStatus constants
const (
	QuestionStatusApproved = "APPROVED"
	QuestionStatusDraft    = "DRAFT"
	QuestionStatusRetired  = "RETIRED"
)

// SourceType constants
const (
	SourceSeedBank = "SEED_BANK"
	SourceAIGemini = "AI_GEMINI"
	SourceAIGroq   = "AI_GROQ"
	SourceAdmin    = "ADMIN"
)

// SessionMode constants
const (
	ModeTopicPractice    = "TOPIC_PRACTICE"
	ModeMixedPractice    = "MIXED_PRACTICE"
	ModeReviewMistakes   = "REVIEW_MISTAKES"
	ModeBookmarkRevision = "BOOKMARK_REVISION"
	ModeMockExam         = "MOCK_EXAM"
)

// ExamPreset constants
const (
	PresetModule1            = "MODULE_1"
	PresetModule2            = "MODULE_2"
	PresetComprehensiveFinal = "COMPREHENSIVE_FINAL"
	PresetCustom             = "CUSTOM"
)

// SessionStatus constants
const (
	SessionStatusActive    = "ACTIVE"
	SessionStatusSubmitted = "SUBMITTED"
	SessionStatusCancelled = "CANCELLED"
	SessionStatusAbandoned = "ABANDONED"
)

// User represents a registered learner, 1-click guest learner, or content administrator.
type User struct {
	ID           string    `gorm:"primaryKey;size:64" json:"id"`
	Name         string    `gorm:"size:120;not null" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	IsGuest      bool      `gorm:"default:false;index" json:"isGuest"`
	Role         string    `gorm:"size:32;default:'LEARNER';index" json:"role"`
	Status       string    `gorm:"size:32;default:'ACTIVE'" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CurriculumTopic represents one of the 10 core Cloud Engineering curriculum areas across Module 1 and Module 2.
type CurriculumTopic struct {
	ID            string              `gorm:"primaryKey;size:64" json:"id"`
	Slug          string              `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	ModuleNumber  int                 `gorm:"not null;index" json:"moduleNumber"`
	ModuleTitle   string              `gorm:"size:160;not null" json:"moduleTitle"`
	WeekCode      string              `gorm:"size:32;not null" json:"weekCode"`
	Title         string              `gorm:"size:160;not null" json:"title"`
	Description   string              `gorm:"type:text;not null" json:"description"`
	SourcePdf     string              `gorm:"size:255;not null" json:"sourcePdf"`
	Ordering      int                 `gorm:"not null;index" json:"ordering"`
	Objectives    []LearningObjective `gorm:"foreignKey:TopicID" json:"objectives,omitempty"`
	QuestionCount int64               `gorm:"-" json:"questionCount"`
	CreatedAt     time.Time           `json:"createdAt"`
	UpdatedAt     time.Time           `json:"updatedAt"`
}

// LearningObjective represents a specific learning objective and extracted course notes within a CurriculumTopic.
type LearningObjective struct {
	ID                string    `gorm:"primaryKey;size:64" json:"id"`
	TopicID           string    `gorm:"size:64;not null;index" json:"topicId"`
	Code              string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Title             string    `gorm:"size:200;not null" json:"title"`
	Statement         string    `gorm:"type:text;not null" json:"statement"`
	CurriculumExcerpt string    `gorm:"type:text;not null" json:"-"`
	SourceReference   string    `gorm:"size:255;not null" json:"sourceReference"`
	Ordering          int       `gorm:"not null" json:"ordering"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Question represents a multiple-choice question in the 1,000+ pre-seeded bank or generated via Gemini/Groq.
type Question struct {
	ID                 string           `gorm:"primaryKey;size:64" json:"id"`
	TopicID            string           `gorm:"size:64;not null;index:idx_question_lookup" json:"topicId"`
	ObjectiveID        string           `gorm:"size:64;not null;index" json:"objectiveId"`
	Stem               string           `gorm:"type:text;not null" json:"stem"`
	QuestionStyle      string           `gorm:"size:32;not null;index:idx_question_lookup" json:"questionStyle"`
	Difficulty         string           `gorm:"size:32;not null;index:idx_question_lookup" json:"difficulty"`
	CognitiveSkill     string           `gorm:"size:32;not null;not null" json:"cognitiveSkill"`
	ScenarioType       string           `gorm:"size:255" json:"scenarioType,omitempty"`
	CorrectOptionKey   string           `gorm:"size:8;not null" json:"-"`
	Explanation        string           `gorm:"type:text;not null" json:"-"`
	SourceType         string           `gorm:"size:32;not null;default:'SEED_BANK';index" json:"sourceType"`
	SourceReference    string           `gorm:"size:255;not null" json:"sourceReference"`
	ContentFingerprint string           `gorm:"size:128;uniqueIndex;not null" json:"-"`
	Status             string           `gorm:"size:32;not null;default:'APPROVED';index:idx_question_lookup" json:"status"`
	Version            int              `gorm:"not null;default:1" json:"version"`
	Options            []QuestionOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"options"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
}

// QuestionOption represents one of the 4 canonical options (A, B, C, D) for a Question.
type QuestionOption struct {
	ID                  string `gorm:"primaryKey;size:64" json:"id"`
	QuestionID          string `gorm:"size:64;not null;uniqueIndex:idx_question_opt_key" json:"questionId"`
	Key                 string `gorm:"size:8;not null;uniqueIndex:idx_question_opt_key" json:"key"`
	Text                string `gorm:"type:text;not null" json:"text"`
	DistractorRationale string `gorm:"type:text" json:"-"`
	Ordering            int    `gorm:"not null" json:"ordering"`
}

// PracticeSession represents a topic practice, mixed practice, mistake review, bookmark revision, or timed mock exam session.
type PracticeSession struct {
	ID                  string            `gorm:"primaryKey;size:64" json:"id"`
	UserID              string            `gorm:"size:64;not null;index" json:"userId"`
	Mode                string            `gorm:"size:32;not null;index" json:"mode"`
	ExamPreset          string            `gorm:"size:32;not null;default:'CUSTOM'" json:"examPreset"`
	Title               string            `gorm:"size:200;not null" json:"title"`
	TopicIDsJSON        string            `gorm:"type:text;not null" json:"topicIdsJson"`
	QuestionStyleFilter string            `gorm:"size:32;not null;default:'ALL'" json:"questionStyleFilter"`
	DifficultyFilter    string            `gorm:"size:32;not null;default:'ALL'" json:"difficultyFilter"`
	FeedbackMode        string            `gorm:"size:32;not null;default:'IMMEDIATE'" json:"feedbackMode"`
	TotalQuestions      int               `gorm:"not null" json:"totalQuestions"`
	TimeLimitSeconds    int               `gorm:"not null;default:0" json:"timeLimitSeconds"`
	Status              string            `gorm:"size:32;not null;default:'ACTIVE';index" json:"status"`
	StartedAt           time.Time         `gorm:"not null" json:"startedAt"`
	Deadline            *time.Time        `json:"deadline,omitempty"`
	SubmittedAt         *time.Time        `json:"submittedAt,omitempty"`
	ScorePercent        float64           `gorm:"default:0" json:"scorePercent"`
	ScorePercentage     float64           `gorm:"-" json:"scorePercentage"`
	CorrectCount        int               `gorm:"default:0" json:"correctCount"`
	IncorrectCount      int               `gorm:"default:0" json:"incorrectCount"`
	UnansweredCount     int               `gorm:"default:0" json:"unansweredCount"`
	SessionQuestions    []SessionQuestion `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"sessionQuestions,omitempty"`
	Responses           []Response        `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"responses,omitempty"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// SessionQuestion binds a Question to a PracticeSession with a per-session shuffled option order (A, B, C, D).
type SessionQuestion struct {
	ID                    string `gorm:"primaryKey;size:64" json:"id"`
	SessionID             string `gorm:"size:64;not null;uniqueIndex:idx_session_question" json:"sessionId"`
	QuestionID            string `gorm:"size:64;not null;uniqueIndex:idx_session_question" json:"questionId"`
	Ordering              int    `gorm:"not null;index" json:"ordering"`
	ShuffledOptionMapping string `gorm:"type:text;not null" json:"-"` // JSON map e.g. {"A":"C","B":"A","C":"D","D":"B"}
	IsFlagged             bool   `gorm:"default:false" json:"isFlagged"`
}

// Response stores a learner's selected option for a specific question within a session.
type Response struct {
	ID                 string    `gorm:"primaryKey;size:64" json:"id"`
	SessionID          string    `gorm:"size:64;not null;uniqueIndex:idx_session_response" json:"sessionId"`
	UserID             string    `gorm:"size:64;not null;index" json:"userId"`
	QuestionID         string    `gorm:"size:64;not null;uniqueIndex:idx_session_response" json:"questionId"`
	SelectedDisplayKey string    `gorm:"size:8;not null" json:"selectedDisplayKey"`
	CanonicalOptionKey string    `gorm:"size:8;not null" json:"canonicalOptionKey"`
	IsCorrect          bool      `gorm:"not null" json:"isCorrect"`
	AnsweredAt         time.Time `gorm:"not null" json:"answeredAt"`
}

// QuestionExposure tracks how many times a specific learner has seen and answered a Question.
type QuestionExposure struct {
	ID            string    `gorm:"primaryKey;size:64" json:"id"`
	UserID        string    `gorm:"size:64;not null;uniqueIndex:idx_user_question_exposure" json:"userId"`
	QuestionID    string    `gorm:"size:64;not null;uniqueIndex:idx_user_question_exposure" json:"questionId"`
	ExposureCount int       `gorm:"not null;default:1" json:"exposureCount"`
	CorrectCount  int       `gorm:"not null;default:0" json:"correctCount"`
	LastOutcome   bool      `gorm:"not null;default:false" json:"lastOutcome"`
	LastSeenAt    time.Time `gorm:"not null" json:"lastSeenAt"`
}

// QuestionBookmark stores a learner's saved/bookmarked question for exam revision.
type QuestionBookmark struct {
	ID         string    `gorm:"primaryKey;size:64" json:"id"`
	UserID     string    `gorm:"size:64;not null;uniqueIndex:idx_user_question_bookmark" json:"userId"`
	QuestionID string    `gorm:"size:64;not null;uniqueIndex:idx_user_question_bookmark" json:"questionId"`
	Note       string    `gorm:"type:text" json:"note,omitempty"`
	Question   Question  `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// GenerationRecord logs metadata and validation outcomes for Gemini and Groq AI question generation calls.
type GenerationRecord struct {
	ID                string    `gorm:"primaryKey;size:64" json:"id"`
	Provider          string    `gorm:"size:32;not null;index" json:"provider"`
	ModelID           string    `gorm:"size:64;not null" json:"modelId"`
	FailoverUsed      bool      `gorm:"default:false" json:"failoverUsed"`
	TopicID           string    `gorm:"size:64;not null;index" json:"topicId"`
	QuestionStyle     string    `gorm:"size:32;not null" json:"questionStyle"`
	RequestedCount    int       `gorm:"not null" json:"requestedCount"`
	AcceptedCount     int       `gorm:"not null" json:"acceptedCount"`
	RejectedCount     int       `gorm:"not null" json:"rejectedCount"`
	ValidationSummary string    `gorm:"type:text" json:"validationSummary"`
	DurationMs        int64     `gorm:"not null" json:"durationMs"`
	CreatedAt         time.Time `json:"createdAt"`
}

// AuditEvent logs administrative or privileged actions.
type AuditEvent struct {
	ID           string    `gorm:"primaryKey;size:64" json:"id"`
	ActorID      string    `gorm:"size:64;not null;index" json:"actorId"`
	Action       string    `gorm:"size:64;not null;index" json:"action"`
	EntityType   string    `gorm:"size:64;not null" json:"entityType"`
	EntityID     string    `gorm:"size:64;not null" json:"entityId"`
	MetadataJSON string    `gorm:"type:text" json:"metadataJson"`
	CreatedAt    time.Time `json:"createdAt"`
}

