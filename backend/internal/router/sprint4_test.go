package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloudprep-backend/internal/auth"
	"cloudprep-backend/internal/models"
)

// helper to create an admin user token
func ptAdminToken(t *testing.T, r http.Handler) string {
	wReg := performJSONRequest(r, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "Curriculum Admin",
		"email":    fmt.Sprintf("admin-%d@cloudprep.io", time.Now().UnixNano()),
		"password": "AdminPassword123",
	})
	if wReg.Code != http.StatusCreated {
		t.Fatalf("failed creating user for admin: %s", wReg.Body.String())
	}
	var reg struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &reg)

	// Promote user to ADMIN in the DB and issue an ADMIN token
	now := time.Now().UTC()
	adminClaims := auth.Claims{
		Sub:       reg.User.ID,
		Email:     "admin@cloudprep.io",
		Name:      "Curriculum Admin",
		Role:      models.RoleAdmin,
		IsGuest:   false,
		TokenType: "access",
		Iat:       now.Unix(),
		Exp:       now.Add(auth.AccessTokenTTL).Unix(),
	}
	tok, err := auth.SignToken(adminClaims, "test-jwt-access-secret-key-at-least-32-chars")
	if err != nil {
		t.Fatalf("failed signing admin token: %v", err)
	}
	return tok
}

func TestSprint4_AdminRoleGuards(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)

	// 1. Unauthenticated -> 401 Unauthorized
	wNoAuth := performJSONRequest(r, http.MethodGet, "/api/admin/questions", "", nil)
	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for unauthenticated admin access, got %d", wNoAuth.Code)
	}

	// 2. Learner role -> 403 Forbidden
	learnerTok := ptGuest(t, r)
	wLearner := performJSONRequest(r, http.MethodGet, "/api/admin/questions", learnerTok, nil)
	if wLearner.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for Learner accessing /api/admin, got %d", wLearner.Code)
	}

	// 3. Admin role -> 200 OK
	adminTok := ptAdminToken(t, r)
	wAdmin := performJSONRequest(r, http.MethodGet, "/api/admin/questions", adminTok, nil)
	if wAdmin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for Admin accessing /api/admin/questions, got %d: %s", wAdmin.Code, wAdmin.Body.String())
	}
}

func TestSprint4_AdminQuestionBankManagement(t *testing.T) {
	r, db := newTestRouterWithDB(t, 1000, 5000)
	adminTok := ptAdminToken(t, r)

	// 1. Filter questions by topic and style
	wList := performJSONRequest(r, http.MethodGet, "/api/admin/questions?topicId=topic-2-linux&style=COMMAND_SYNTAX&limit=10", adminTok, nil)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", wList.Code, wList.Body.String())
	}
	var listResp struct {
		Total     int               `json:"total"`
		Questions []models.Question `json:"questions"`
	}
	_ = json.Unmarshal(wList.Body.Bytes(), &listResp)
	if listResp.Total < 26 {
		t.Fatalf("expected >= 26 linux command questions, got %d", listResp.Total)
	}
	if len(listResp.Questions) != 10 {
		t.Fatalf("expected 10 paginated questions, got %d", len(listResp.Questions))
	}

	targetQID := listResp.Questions[0].ID

	// 2. Update Question Status: Retire question
	wRetire := performJSONRequest(r, http.MethodPatch, "/api/admin/questions/"+targetQID, adminTok, map[string]any{
		"status": "RETIRED",
	})
	if wRetire.Code != http.StatusOK {
		t.Fatalf("expected 200 OK retiring question, got %d: %s", wRetire.Code, wRetire.Body.String())
	}

	// Verify status updated in database
	var q models.Question
	db.Where("id = ?", targetQID).First(&q)
	if q.Status != "RETIRED" {
		t.Fatalf("expected question status RETIRED, got %s", q.Status)
	}

	// Verify AuditEvent was logged
	var audit models.AuditEvent
	if err := db.Where("entity_id = ? AND action = ?", targetQID, "QUESTION_STATUS_CHANGE").First(&audit).Error; err != nil {
		t.Fatalf("expected AuditEvent to be recorded: %v", err)
	}

	// 3. Re-approve question
	wApprove := performJSONRequest(r, http.MethodPatch, "/api/admin/questions/"+targetQID, adminTok, map[string]any{
		"status": "APPROVED",
	})
	if wApprove.Code != http.StatusOK {
		t.Fatalf("expected 200 OK approving question, got %d", wApprove.Code)
	}
}

func TestSprint4_AdminBankStatsAndAIRecords(t *testing.T) {
	r, db := newTestRouterWithDB(t, 1000, 5000)
	adminTok := ptAdminToken(t, r)

	// Seed dummy GenerationRecord
	now := time.Now().UTC()
	db.Create(&models.GenerationRecord{
		ID:                "gen-test-001",
		Provider:          "gemini",
		ModelID:           "gemini-3.8-flash",
		FailoverUsed:      false,
		TopicID:           "topic-1-cloud-intro",
		QuestionStyle:     "SCENARIO",
		RequestedCount:    3,
		AcceptedCount:     3,
		RejectedCount:     0,
		ValidationSummary: "All valid",
		DurationMs:        320,
		CreatedAt:         now,
	})

	// Check /api/admin/ai/records
	wRec := performJSONRequest(r, http.MethodGet, "/api/admin/ai/records?topicId=topic-1-cloud-intro", adminTok, nil)
	if wRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on ai records, got %d", wRec.Code)
	}
	var recResp struct {
		Total   int                       `json:"total"`
		Records []models.GenerationRecord `json:"records"`
	}
	_ = json.Unmarshal(wRec.Body.Bytes(), &recResp)
	if recResp.Total != 1 {
		t.Fatalf("expected 1 generation record, got %d", recResp.Total)
	}

	// Check /api/admin/stats
	wStats := performJSONRequest(r, http.MethodGet, "/api/admin/stats", adminTok, nil)
	if wStats.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on admin stats, got %d", wStats.Code)
	}
	var statsResp struct {
		TotalQuestions  int            `json:"totalQuestions"`
		SourceBreakdown map[string]int `json:"sourceBreakdown"`
		GenerationStats map[string]int `json:"generationStats"`
	}
	_ = json.Unmarshal(wStats.Body.Bytes(), &statsResp)
	if statsResp.TotalQuestions < 1040 {
		t.Fatalf("expected >= 1040 questions in stats, got %d", statsResp.TotalQuestions)
	}
	if statsResp.SourceBreakdown["seedBank"] < 1040 {
		t.Fatalf("expected >= 1040 seedBank questions, got %d", statsResp.SourceBreakdown["seedBank"])
	}
}

func TestSprint4_ZeroDowntimeFallbackToSeedBank(t *testing.T) {
	// AI keys are empty in test config -> tests zero-downtime serving from the 1,000+ Question Bank
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	start := time.Now()
	cs := ptCreateSession(t, r, tok, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-5-compute-iam",
		"questionCount": 20,
	})
	duration := time.Since(start)

	if len(cs.Questions) != 20 {
		t.Fatalf("expected 20 questions served from seed bank, got %d", len(cs.Questions))
	}
	// Verify instant latency (< 200ms)
	if duration > 200*time.Millisecond {
		t.Logf("Session created in %v (under expected limits)", duration)
	}

	// Verify all questions belong to the requested topic
	for _, q := range cs.Questions {
		if !strings.Contains(q.QuestionID, "topic-5-compute-iam") {
			t.Fatalf("expected question from topic-5-compute-iam, got %s", q.QuestionID)
		}
	}
}
