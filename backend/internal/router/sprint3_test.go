package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func performJSONRequest(r http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestZeroFrictionAuthAndMultipleGuests(t *testing.T) {
	r := newTestRouter(t, 500, 1000)

	// 1. Register new learner (instant token, no email verification wall)
	wReg := performJSONRequest(r, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name":     "Ada Lovelace",
		"email":    "ada@cloudprep.io",
		"password": "securePassword123",
	})
	if wReg.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on /api/auth/register, got %d: %s", wReg.Code, wReg.Body.String())
	}

	var regResp struct {
		Token string         `json:"token"`
		User  map[string]any `json:"user"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &regResp)
	if regResp.Token == "" {
		t.Fatal("expected non-empty JWT token on registration")
	}
	if regResp.User["isGuest"] != false {
		t.Errorf("expected isGuest=false for registered user, got %v", regResp.User["isGuest"])
	}

	// 2. Login with valid and invalid credentials
	wLoginBad := performJSONRequest(r, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email":    "ada@cloudprep.io",
		"password": "wrongPassword",
	})
	if wLoginBad.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong password, got %d", wLoginBad.Code)
	}

	wLoginOk := performJSONRequest(r, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email":    "ada@cloudprep.io",
		"password": "securePassword123",
	})
	if wLoginOk.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on valid login, got %d: %s", wLoginOk.Code, wLoginOk.Body.String())
	}

	// 3. Create multiple 1-Click Guest accounts (verifies no unique email index collision)
	for i := 1; i <= 3; i++ {
		wGuest := performJSONRequest(r, http.MethodPost, "/api/auth/guest", "", map[string]any{})
		if wGuest.Code != http.StatusCreated {
			t.Fatalf("guest #%d expected 201 Created, got %d: %s", i, wGuest.Code, wGuest.Body.String())
		}
		var gResp struct {
			Token string         `json:"token"`
			User  map[string]any `json:"user"`
		}
		_ = json.Unmarshal(wGuest.Body.Bytes(), &gResp)
		if gResp.Token == "" || gResp.User["isGuest"] != true {
			t.Fatalf("guest #%d invalid response: %v", i, gResp)
		}

		// Verify GET /api/me works for guest token
		wMe := performJSONRequest(r, http.MethodGet, "/api/me", gResp.Token, nil)
		if wMe.Code != http.StatusOK {
			t.Fatalf("guest #%d GET /api/me expected 200, got %d", i, wMe.Code)
		}
	}
}

func TestPracticeSessionImmediateFeedbackAndOptionShuffling(t *testing.T) {
	r := newTestRouter(t, 500, 1000)

	wGuest := performJSONRequest(r, http.MethodPost, "/api/auth/guest", "", nil)
	var gResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wGuest.Body.Bytes(), &gResp)

	// Start a Practice Session filtered by Topic 2 (Linux) and Style COMMAND_SYNTAX
	wCreate := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", gResp.Token, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-2-linux",
		"questionStyle": "COMMAND_SYNTAX",
		"questionCount": 5,
		"feedbackMode":  "IMMEDIATE",
	})
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for practice session, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var createResp struct {
		Session   map[string]any   `json:"session"`
		Questions []map[string]any `json:"questions"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	sessionID := createResp.Session["id"].(string)
	if len(createResp.Questions) != 5 {
		t.Fatalf("expected 5 questions, got %d", len(createResp.Questions))
	}

	firstQ := createResp.Questions[0]
	firstQID := firstQ["questionId"].(string)
	if firstQ["questionStyle"] != "COMMAND_SYNTAX" {
		t.Errorf("expected questionStyle=COMMAND_SYNTAX, got %v", firstQ["questionStyle"])
	}

	// Answer the first question in IMMEDIATE feedback mode
	wAns := performJSONRequest(r, http.MethodPut, "/api/practice-sessions/"+sessionID+"/responses/"+firstQID, gResp.Token, map[string]any{
		"selectedOptionKey": "B",
	})
	if wAns.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on answer submission, got %d: %s", wAns.Code, wAns.Body.String())
	}
	var ansResp map[string]any
	_ = json.Unmarshal(wAns.Body.Bytes(), &ansResp)
	if ansResp["explanation"] == nil || ansResp["explanation"] == "" {
		t.Errorf("expected immediate feedback explanation in Practice Mode, got %v", ansResp)
	}
	if ansResp["correctOptionKey"] == nil || ansResp["correctOptionKey"] == "" {
		t.Errorf("expected correctOptionKey in immediate Practice Mode feedback")
	}

	// Flag the question for review
	wFlag := performJSONRequest(r, http.MethodPut, "/api/practice-sessions/"+sessionID+"/questions/"+firstQID+"/flag", gResp.Token, nil)
	if wFlag.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on flag toggle, got %d", wFlag.Code)
	}
}

func TestMockExamAnswerConcealmentAndIdempotentSubmit(t *testing.T) {
	r := newTestRouter(t, 500, 1000)

	wGuest := performJSONRequest(r, http.MethodPost, "/api/auth/guest", "", nil)
	var gResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wGuest.Body.Bytes(), &gResp)

	// Start a MODULE_1 Timed Mock Exam
	wCreate := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", gResp.Token, map[string]any{
		"mode":          "MOCK_EXAM",
		"examPreset":    "MODULE_1",
		"questionCount": 8,
	})
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for MODULE_1 Mock Exam, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	// Verify active mock exam payload NEVER leaks correctOptionKey, explanation, or distractorRationale
	rawBody := wCreate.Body.String()
	if strings.Contains(rawBody, "correctOptionKey") || strings.Contains(rawBody, "explanation") || strings.Contains(rawBody, "distractorRationale") {
		t.Fatalf("SECURITY FAILURE: active Mock Exam leaked answer/explanation metadata in CreateSession response: %s", rawBody)
	}

	var createResp struct {
		Session   map[string]any   `json:"session"`
		Questions []map[string]any `json:"questions"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	sessionID := createResp.Session["id"].(string)
	firstQID := createResp.Questions[0]["questionId"].(string)

	// Verify saving an answer during an active Mock Exam ALSO conceals correctness and explanations
	wAns := performJSONRequest(r, http.MethodPut, "/api/practice-sessions/"+sessionID+"/responses/"+firstQID, gResp.Token, map[string]any{
		"selectedOptionKey": "A",
	})
	if wAns.Code != http.StatusOK {
		t.Fatalf("expected 200 OK saving mock exam answer, got %d", wAns.Code)
	}
	if strings.Contains(wAns.Body.String(), "correctOptionKey") || strings.Contains(wAns.Body.String(), "explanation") || strings.Contains(wAns.Body.String(), "isCorrect") {
		t.Fatalf("SECURITY FAILURE: active Mock Exam response save leaked correctness: %s", wAns.Body.String())
	}

	// Verify calling /review BEFORE submitting the Mock Exam is blocked with 403 Forbidden
	wEarlyReview := performJSONRequest(r, http.MethodGet, "/api/practice-sessions/"+sessionID+"/review", gResp.Token, nil)
	if wEarlyReview.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when requesting /review on ACTIVE Mock Exam, got %d", wEarlyReview.Code)
	}

	// Submit the Mock Exam (1st call)
	wSub1 := performJSONRequest(r, http.MethodPost, "/api/practice-sessions/"+sessionID+"/submit", gResp.Token, nil)
	if wSub1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on submit #1, got %d: %s", wSub1.Code, wSub1.Body.String())
	}
	var sub1Resp map[string]any
	_ = json.Unmarshal(wSub1.Body.Bytes(), &sub1Resp)
	if sub1Resp["alreadySubmitted"] != false {
		t.Errorf("expected alreadySubmitted=false on first submit, got %v", sub1Resp["alreadySubmitted"])
	}

	// Submit the Mock Exam a 2nd time (Idempotency verification)
	wSub2 := performJSONRequest(r, http.MethodPost, "/api/practice-sessions/"+sessionID+"/submit", gResp.Token, nil)
	if wSub2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on idempotent submit #2, got %d", wSub2.Code)
	}
	var sub2Resp map[string]any
	_ = json.Unmarshal(wSub2.Body.Bytes(), &sub2Resp)
	if sub2Resp["alreadySubmitted"] != true {
		t.Errorf("expected alreadySubmitted=true on second submit, got %v", sub2Resp["alreadySubmitted"])
	}

	// Now /review must succeed (200 OK) and include explanations, distractor rationales, and sourceReferences
	wReview := performJSONRequest(r, http.MethodGet, "/api/practice-sessions/"+sessionID+"/review", gResp.Token, nil)
	if wReview.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on post-submit /review, got %d: %s", wReview.Code, wReview.Body.String())
	}
	if !strings.Contains(wReview.Body.String(), "explanation") || !strings.Contains(wReview.Body.String(), "sourceReference") {
		t.Errorf("expected post-submit review to contain explanation and sourceReference")
	}
}

func TestBookmarksAndProgressAnalytics(t *testing.T) {
	r := newTestRouter(t, 500, 1000)

	wGuest := performJSONRequest(r, http.MethodPost, "/api/auth/guest", "", nil)
	var gResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(wGuest.Body.Bytes(), &gResp)

	// 1. Verify C-1 Hardening: Attempting to bookmark an arbitrary unserved question is blocked
	wUnserved := performJSONRequest(r, http.MethodPost, "/api/bookmarks/q-topic-2-linux-cmd-001", gResp.Token, map[string]any{
		"note": "Trying to dump answer key",
	})
	if wUnserved.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when bookmarking unserved question, got %d: %s", wUnserved.Code, wUnserved.Body.String())
	}

	// 2. Start a legitimate session so questions are legitimately served to the user
	wSession := performJSONRequest(r, http.MethodPost, "/api/practice-sessions", gResp.Token, map[string]any{
		"mode":          "TOPIC_PRACTICE",
		"topicId":       "topic-2-linux",
		"questionCount": 5,
	})
	if wSession.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for practice session, got %d: %s", wSession.Code, wSession.Body.String())
	}
	var sResp struct {
		Questions []map[string]any `json:"questions"`
	}
	_ = json.Unmarshal(wSession.Body.Bytes(), &sResp)
	targetQID := sResp.Questions[0]["questionId"].(string)

	// 3. Bookmark the legitimately served question
	wBmAdd := performJSONRequest(r, http.MethodPost, "/api/bookmarks/"+targetQID, gResp.Token, map[string]any{
		"note": "Review chmod octal triads before exam",
	})
	if wBmAdd.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on bookmark add, got %d: %s", wBmAdd.Code, wBmAdd.Body.String())
	}

	// 4. List bookmarks
	wBmList := performJSONRequest(r, http.MethodGet, "/api/bookmarks", gResp.Token, nil)
	if wBmList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on bookmark list, got %d", wBmList.Code)
	}
	var bmListResp struct {
		Total     int              `json:"total"`
		Bookmarks []map[string]any `json:"bookmarks"`
	}
	_ = json.Unmarshal(wBmList.Body.Bytes(), &bmListResp)
	if bmListResp.Total != 1 || bmListResp.Bookmarks[0]["questionId"] != targetQID {
		t.Fatalf("expected 1 bookmark for %s, got %+v", targetQID, bmListResp)
	}

	// 5. Check Progress Summary
	wProg := performJSONRequest(r, http.MethodGet, "/api/progress/summary", gResp.Token, nil)
	if wProg.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /api/progress/summary, got %d: %s", wProg.Code, wProg.Body.String())
	}
	var progResp struct {
		BookmarkCount  int              `json:"bookmarkCount"`
		TopicBreakdown []map[string]any `json:"topicBreakdown"`
		StyleBreakdown []map[string]any `json:"styleBreakdown"`
	}
	_ = json.Unmarshal(wProg.Body.Bytes(), &progResp)
	if progResp.BookmarkCount != 1 {
		t.Errorf("expected bookmarkCount=1 in progress summary, got %d", progResp.BookmarkCount)
	}
	if len(progResp.TopicBreakdown) != 10 {
		t.Errorf("expected 10 topics in topicBreakdown, got %d", len(progResp.TopicBreakdown))
	}
	if len(progResp.StyleBreakdown) != 4 {
		t.Errorf("expected 4 styles in styleBreakdown, got %d", len(progResp.StyleBreakdown))
	}

	// 6. Delete bookmark
	wBmDel := performJSONRequest(r, http.MethodDelete, "/api/bookmarks/"+targetQID, gResp.Token, nil)
	if wBmDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on bookmark delete, got %d", wBmDel.Code)
	}
}

