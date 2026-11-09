package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type ptOpt struct {
	Key                 string `json:"key"`
	Text                string `json:"text"`
	DistractorRationale string `json:"distractorRationale,omitempty"`
}

type ptQuestion struct {
	QuestionID string  `json:"questionId"`
	Options    []ptOpt `json:"options"`
}

type ptCreate struct {
	Session   map[string]any `json:"session"`
	Questions []ptQuestion   `json:"questions"`
}

func ptGuest(t *testing.T, r http.Handler) string {
	w := performJSONRequest(r, http.MethodPost, "/api/auth/guest", "", nil)
	var g struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &g)
	if g.Token == "" {
		t.Fatalf("guest failed: %s", w.Body.String())
	}
	return g.Token
}

func ptCreateSession(t *testing.T, r http.Handler, tok string, body map[string]any) ptCreate {
	w := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", tok, body)
	if w.Code != http.StatusCreated {
		t.Fatalf("create session failed %d: %s", w.Code, w.Body.String())
	}
	var cs ptCreate
	_ = json.Unmarshal(w.Body.Bytes(), &cs)
	return cs
}

func ptSubmitScore(r http.Handler, tok, sid string) float64 {
	w := performJSONRequest(r, http.MethodPost, "/api/practice-sessions/"+sid+"/submit", tok, nil)
	var s struct {
		Session struct {
			ScorePercent float64 `json:"scorePercent"`
		} `json:"session"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &s)
	return s.Session.ScorePercent
}

// TestHardened_C1_BookmarkMockExamAnswerRedacted verifies that bookmarking an active mock exam
// question NEVER reveals the correctOptionKey, explanation, or distractor rationales via GET /api/bookmarks.
func TestHardened_C1_BookmarkMockExamAnswerRedacted(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	cs := ptCreateSession(t, r, tok, map[string]any{
		"mode":       "MOCK_EXAM",
		"examPreset": "MODULE_1",
	})
	if len(cs.Questions) == 0 {
		t.Fatal("expected questions in mock exam session")
	}

	// Bookmark all questions from the active mock exam
	for _, q := range cs.Questions {
		wBm := performJSONRequest(r, http.MethodPost, "/api/bookmarks/"+q.QuestionID, tok, nil)
		if wBm.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for bookmarking served question, got %d: %s", wBm.Code, wBm.Body.String())
		}
	}

	// Fetch bookmarks while the mock exam is still ACTIVE
	wb := performJSONRequest(r, http.MethodGet, "/api/bookmarks", tok, nil)
	if wb.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /api/bookmarks, got %d", wb.Code)
	}

	var bl struct {
		Bookmarks []struct {
			QuestionID       string  `json:"questionId"`
			CorrectOptionKey string  `json:"correctOptionKey"`
			Explanation      string  `json:"explanation"`
			Options          []ptOpt `json:"options"`
		} `json:"bookmarks"`
	}
	_ = json.Unmarshal(wb.Body.Bytes(), &bl)

	for _, b := range bl.Bookmarks {
		if b.CorrectOptionKey != "" {
			t.Fatalf("SECURITY VIOLATION C-1: active mock exam question %s leaked correctOptionKey=%q in bookmarks", b.QuestionID, b.CorrectOptionKey)
		}
		if b.Explanation != "" {
			t.Fatalf("SECURITY VIOLATION C-1: active mock exam question %s leaked explanation in bookmarks", b.QuestionID)
		}
		for _, opt := range b.Options {
			if opt.DistractorRationale != "" {
				t.Fatalf("SECURITY VIOLATION C-1: active mock exam question %s leaked distractor rationale in bookmarks", b.QuestionID)
			}
		}
	}
}

// TestHardened_C1_UnservedQuestionBookmarkRejected verifies that an attacker cannot enumerate
// predictable question IDs to dump the question bank without being legitimately served the question.
func TestHardened_C1_UnservedQuestionBookmarkRejected(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	w := performJSONRequest(r, http.MethodPost, "/api/bookmarks/q-topic-1-cloud-intro-def-001", tok, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("SECURITY VIOLATION C-1: expected 403 Forbidden for unserved question bookmark, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "QUESTION_NOT_SERVED") {
		t.Fatalf("expected QUESTION_NOT_SERVED error code, got %s", w.Body.String())
	}
}

// TestHardened_C2_ActiveMockExamIntegrity verifies that a user cannot start a second mock exam
// while one is active, and questions on the active mock exam cannot be leaked into concurrent practice.
func TestHardened_C2_ActiveMockExamIntegrity(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	// Create 1st mock exam
	cs1 := ptCreateSession(t, r, tok, map[string]any{
		"mode":       "MOCK_EXAM",
		"examPreset": "MODULE_1",
	})
	if len(cs1.Questions) == 0 {
		t.Fatal("expected questions in mock exam session")
	}

	// Attempt to start a 2nd mock exam while 1st is ACTIVE -> must be rejected 409
	wSecondMock := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", tok, map[string]any{
		"mode":       "MOCK_EXAM",
		"examPreset": "MODULE_2",
	})
	if wSecondMock.Code != http.StatusConflict {
		t.Fatalf("SECURITY VIOLATION C-2: expected 409 Conflict when creating second concurrent mock exam, got %d: %s", wSecondMock.Code, wSecondMock.Body.String())
	}

	// Starting a practice session must exclude the active mock exam's questions
	wPractice := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", tok, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-1-cloud-intro",
		"questionCount": 10,
	})
	if wPractice.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for practice session, got %d: %s", wPractice.Code, wPractice.Body.String())
	}

	var csPractice ptCreate
	_ = json.Unmarshal(wPractice.Body.Bytes(), &csPractice)

	mockQIDs := make(map[string]bool)
	for _, q := range cs1.Questions {
		mockQIDs[q.QuestionID] = true
	}

	for _, pq := range csPractice.Questions {
		if mockQIDs[pq.QuestionID] {
			t.Fatalf("SECURITY VIOLATION C-2: practice session served question %s which is currently on an active mock exam", pq.QuestionID)
		}
	}
}

// TestHardened_H1_ConcurrentSubmitRaceResolved verifies that multiple concurrent submit calls
// execute atomically without duplicate exposure increments or multiple 200 firstSubmits.
func TestHardened_H1_ConcurrentSubmitRaceResolved(t *testing.T) {
	r := newTestRouter(t, 100000, 100000)
	tok := ptGuest(t, r)

	cs := ptCreateSession(t, r, tok, map[string]any{
		"mode":       "MOCK_EXAM",
		"examPreset": "MODULE_1",
	})
	sid := cs.Session["id"].(string)

	var wg sync.WaitGroup
	var mu sync.Mutex
	firstSubmits := 0
	alreadySubmits := 0
	errors := 0

	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := performJSONRequest(r, http.MethodPost, "/api/practice-sessions/"+sid+"/submit", tok, nil)
			mu.Lock()
			defer mu.Unlock()
			if w.Code != http.StatusOK {
				errors++
				return
			}
			if strings.Contains(w.Body.String(), `"alreadySubmitted":false`) {
				firstSubmits++
			} else if strings.Contains(w.Body.String(), `"alreadySubmitted":true`) {
				alreadySubmits++
			}
		}()
	}
	wg.Wait()

	if errors > 0 {
		t.Fatalf("expected 0 errors during concurrent submits, got %d", errors)
	}
	if firstSubmits != 1 {
		t.Fatalf("SECURITY VIOLATION H-1: expected exactly 1 firstSubmit=false, got %d (alreadySubmits=%d)", firstSubmits, alreadySubmits)
	}
	if alreadySubmits != 11 {
		t.Fatalf("expected 11 alreadySubmitted=true, got %d", alreadySubmits)
	}
}

// TestHardened_H4_ImmediatePracticeAnswerLocked verifies that once an answer is revealed
// in IMMEDIATE feedback mode, the learner cannot overwrite their answer to inflate their score.
func TestHardened_H4_ImmediatePracticeAnswerLocked(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	cs := ptCreateSession(t, r, tok, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-1-cloud-intro",
		"questionCount": 5,
		"feedbackMode":  "IMMEDIATE",
	})
	sid := cs.Session["id"].(string)
	targetQID := cs.Questions[0].QuestionID

	// 1st answer submission -> returns 200 OK with correct answer revealed
	w1 := performJSONRequest(r, http.MethodPut, "/api/practice-sessions/"+sid+"/responses/"+targetQID, tok, map[string]any{
		"selectedOptionKey": "A",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on first answer, got %d: %s", w1.Code, w1.Body.String())
	}

	// 2nd answer submission for the same question -> MUST BE REJECTED 409 ANSWER_LOCKED
	w2 := performJSONRequest(r, http.MethodPut, "/api/practice-sessions/"+sid+"/responses/"+targetQID, tok, map[string]any{
		"selectedOptionKey": "B",
	})
	if w2.Code != http.StatusConflict {
		t.Fatalf("SECURITY VIOLATION H-4: expected 409 Conflict when changing answer after reveal, got %d: %s", w2.Code, w2.Body.String())
	}
	if !strings.Contains(w2.Body.String(), "ANSWER_LOCKED") {
		t.Fatalf("expected ANSWER_LOCKED error code, got %s", w2.Body.String())
	}
}

// TestHardened_H5_MockExamPresetTamperingBlocked verifies that clients cannot bypass official
// preset question count or time limits.
func TestHardened_H5_MockExamPresetTamperingBlocked(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	cs := ptCreateSession(t, r, tok, map[string]any{
		"mode":             "MOCK_EXAM",
		"examPreset":       "MODULE_1",
		"questionCount":    3,
		"timeLimitSeconds": 864000, // 10 days
	})

	tl := cs.Session["timeLimitSeconds"].(float64)
	tq := cs.Session["totalQuestions"].(float64)

	// MODULE_1 preset is 30 questions and 40 minutes (2400 seconds)
	if tl != 2400 {
		t.Fatalf("SECURITY VIOLATION H-5: expected server-enforced timeLimitSeconds=2400, got %.0f", tl)
	}
	if tq != 30 {
		t.Fatalf("SECURITY VIOLATION H-5: expected server-enforced totalQuestions=30, got %.0f", tq)
	}
}

// TestHardened_M1_PasswordLengthValidation verifies that passwords > 72 bytes and < 8 bytes
// are rejected cleanly with 400 Validation Error, preventing bcrypt truncation and crashes.
func TestHardened_M1_PasswordLengthValidation(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)

	// Password too short (< 8 chars)
	wShort := performJSONRequest(r, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "Short",
		"email":    "short@cloudprep.io",
		"password": "12345",
	})
	if wShort.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for short password, got %d: %s", wShort.Code, wShort.Body.String())
	}

	// Password too long (> 72 bytes)
	wLong := performJSONRequest(r, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "Long",
		"email":    "long@cloudprep.io",
		"password": strings.Repeat("a", 80),
	})
	if wLong.Code != http.StatusBadRequest {
		t.Fatalf("SECURITY VIOLATION M-1: expected 400 Bad Request for 80-byte password, got %d: %s", wLong.Code, wLong.Body.String())
	}

	// Login with > 72 byte password -> rejected with 401 without bcrypt crash
	wLoginLong := performJSONRequest(r, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email":    "anyone@cloudprep.io",
		"password": strings.Repeat("a", 80),
	})
	if wLoginLong.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for oversized login password, got %d: %s", wLoginLong.Code, wLoginLong.Body.String())
	}
}

// TestHardened_M8_ToggleFlagOnSubmittedSessionBlocked verifies that flags cannot be toggled
// once a session has been finalized.
func TestHardened_M8_ToggleFlagOnSubmittedSessionBlocked(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	cs := ptCreateSession(t, r, tok, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-1-cloud-intro",
		"questionCount": 5,
	})
	sid := cs.Session["id"].(string)
	qid := cs.Questions[0].QuestionID

	// Submit the session
	ptSubmitScore(r, tok, sid)

	// Attempt to toggle flag on SUBMITTED session
	wFlag := performJSONRequest(r, http.MethodPut, "/api/practice-sessions/"+sid+"/questions/"+qid+"/flag", tok, map[string]any{
		"isFlagged": true,
	})
	if wFlag.Code != http.StatusConflict {
		t.Fatalf("SECURITY VIOLATION M-8: expected 409 Conflict when toggling flag on submitted session, got %d: %s", wFlag.Code, wFlag.Body.String())
	}
}

// TestHardened_M5_BearerOnlyAuth verifies that cookies alone do not authenticate API calls,
// preventing CSRF attacks.
func TestHardened_M5_BearerOnlyAuth(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	// Send request with token ONLY in cookie
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "cloudprep_token",
		Value: tok,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("SECURITY VIOLATION M-5: cookie-based authentication was accepted, expected 401: got %d", w.Code)
	}

	// Send request with Bearer header -> 200 OK
	wBearer := performJSONRequest(r, http.MethodGet, "/api/me", tok, nil)
	if wBearer.Code != http.StatusOK {
		t.Fatalf("expected 200 OK with Bearer header, got %d: %s", wBearer.Code, wBearer.Body.String())
	}
}

// TestHardened_ResourceProtection_ActiveSessionCap verifies that a single user cannot create
// more than 5 active sessions concurrently, preventing database bloat.
func TestHardened_ResourceProtection_ActiveSessionCap(t *testing.T) {
	r := newTestRouter(t, 1000, 5000)
	tok := ptGuest(t, r)

	for i := 0; i < 5; i++ {
		w := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", tok, map[string]any{
			"mode":          "TOPIC_PRACTICE",
			"topicId":       "topic-1-cloud-intro",
			"questionCount": 5,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for session %d, got %d", i+1, w.Code)
		}
	}

	// 6th session must be rejected with 409 MAX_ACTIVE_SESSIONS_EXCEEDED
	w6 := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", tok, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-1-cloud-intro",
		"questionCount": 5,
	})
	if w6.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when exceeding 5 active sessions, got %d: %s", w6.Code, w6.Body.String())
	}
	if !strings.Contains(w6.Body.String(), "MAX_ACTIVE_SESSIONS_EXCEEDED") {
		t.Fatalf("expected MAX_ACTIVE_SESSIONS_EXCEEDED error, got %s", w6.Body.String())
	}
}

// TestHardened_M4_InactiveUserRefreshRejected verifies that a suspended/inactive user
// cannot refresh their access token (M-4).
func TestHardened_M4_InactiveUserRefreshRejected(t *testing.T) {
	r, database := newTestRouterWithDB(t, 1000, 5000)

	// Register user
	wReg := performJSONRequest(r, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "Suspended User",
		"email":    "suspended@cloudprep.io",
		"password": "strongPassword123",
	})
	if wReg.Code != http.StatusCreated {
		t.Fatalf("register failed: %s", wReg.Body.String())
	}
	var regResp struct {
		RefreshToken string `json:"refreshToken"`
		User         struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &regResp)

	// Deactivate user in database
	database.Table("users").Where("id = ?", regResp.User.ID).Update("status", "SUSPENDED")

	// Attempt token refresh
	wRef := performJSONRequest(r, http.MethodPost, "/api/auth/refresh", "", map[string]any{
		"refreshToken": regResp.RefreshToken,
	})
	if wRef.Code != http.StatusUnauthorized {
		t.Fatalf("SECURITY VIOLATION M-4: expected 401 Unauthorized when refreshing token for suspended user, got %d: %s", wRef.Code, wRef.Body.String())
	}
	if !strings.Contains(wRef.Body.String(), "USER_INACTIVE") {
		t.Fatalf("expected USER_INACTIVE error code, got %s", wRef.Body.String())
	}
}
