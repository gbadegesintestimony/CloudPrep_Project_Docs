package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudprep-backend/internal/config"
	"cloudprep-backend/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func newTestRouterWithDB(t *testing.T, rps float64, burst int) (*gin.Engine, *gorm.DB) {
	t.Helper()
	authRPS := rps / 2
	if authRPS <= 0 {
		authRPS = 5.0
	}
	authBurst := burst / 2
	if authBurst <= 0 {
		authBurst = 10
	}
	cfg := &config.Config{
		Port:               "8080",
		GinMode:            gin.TestMode,
		DatabaseURL:        ":memory:",
		FrontendOrigins:    []string{"http://localhost:3000"},
		JWTAccessSecret:    "test-jwt-access-secret-key-at-least-32-chars",
		JWTRefreshSecret:   "test-jwt-refresh-secret-key-at-least-32-chars",
		RateLimitRPS:       rps,
		RateLimitBurst:     burst,
		RateLimitAuthRPS:   authRPS,
		RateLimitAuthBurst: authBurst,
	}

	database, dbDriver, err := db.ConnectAndMigrate(cfg)
	if err != nil {
		t.Fatalf("failed to connect and migrate test db: %v", err)
	}

	return SetupRouter(cfg, database, dbDriver), database
}

func newTestRouter(t *testing.T, rps float64, burst int) *gin.Engine {
	r, _ := newTestRouterWithDB(t, rps, burst)
	return r
}

func TestHealthAndCurriculumEndpoints(t *testing.T) {
	r := newTestRouter(t, 100, 200)

	// 1. Test GET /api/health
	reqHealth := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	wHealth := httptest.NewRecorder()
	r.ServeHTTP(wHealth, reqHealth)

	if wHealth.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /api/health, got %d: %s", wHealth.Code, wHealth.Body.String())
	}

	// Verify defensive security headers
	if got := wHealth.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("expected X-Content-Type-Options=nosniff, got %q", got)
	}
	if got := wHealth.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("expected X-Frame-Options=DENY, got %q", got)
	}
	if got := wHealth.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary=Origin, got %q", got)
	}

	var healthResp map[string]any
	if err := json.Unmarshal(wHealth.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("failed to parse health JSON: %v", err)
	}
	if healthResp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", healthResp["status"])
	}
	if healthResp["dbActive"] != true {
		t.Errorf("expected dbActive=true, got %v", healthResp["dbActive"])
	}
	if _, leaked := healthResp["dbDriver"]; leaked {
		t.Errorf("expected dbDriver to be omitted from public /api/health response")
	}
	if int(healthResp["topicCount"].(float64)) != 10 {
		t.Errorf("expected topicCount=10, got %v", healthResp["topicCount"])
	}
	if int(healthResp["objectiveCount"].(float64)) != 31 {
		t.Errorf("expected objectiveCount=31, got %v", healthResp["objectiveCount"])
	}
	if int(healthResp["bankCount"].(float64)) < 1000 {
		t.Errorf("expected bankCount >= 1000, got %v", healthResp["bankCount"])
	}

	// 2. Test GET /api/curriculum and verify internal curriculumExcerpt is not leaked
	reqCurr := httptest.NewRequest(http.MethodGet, "/api/curriculum", nil)
	wCurr := httptest.NewRecorder()
	r.ServeHTTP(wCurr, reqCurr)

	if wCurr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /api/curriculum, got %d: %s", wCurr.Code, wCurr.Body.String())
	}
	if strings.Contains(wCurr.Body.String(), "curriculumExcerpt") {
		t.Errorf("expected curriculumExcerpt to be hidden from public /api/curriculum response")
	}

	var currResp struct {
		Topics             []map[string]any `json:"topics"`
		ExamPresets        []map[string]any `json:"examPresets"`
		TotalBankQuestions int              `json:"totalBankQuestions"`
	}
	if err := json.Unmarshal(wCurr.Body.Bytes(), &currResp); err != nil {
		t.Fatalf("failed to parse curriculum JSON: %v", err)
	}
	if len(currResp.Topics) != 10 {
		t.Errorf("expected 10 curriculum topics, got %d", len(currResp.Topics))
	}
	if len(currResp.ExamPresets) != 3 {
		t.Errorf("expected 3 exam presets, got %d", len(currResp.ExamPresets))
	}
	if currResp.TotalBankQuestions < 1000 {
		t.Errorf("expected totalBankQuestions >= 1000, got %d", currResp.TotalBankQuestions)
	}
}

func TestStrictCORSPolicy(t *testing.T) {
	r := newTestRouter(t, 100, 200)

	// Allowed origin preflight
	reqAllowed := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	reqAllowed.Header.Set("Origin", "http://localhost:3000")
	wAllowed := httptest.NewRecorder()
	r.ServeHTTP(wAllowed, reqAllowed)
	if wAllowed.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for allowed origin OPTIONS, got %d", wAllowed.Code)
	}
	if got := wAllowed.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin=http://localhost:3000, got %q", got)
	}

	// Untrusted origin preflight must be rejected with 403
	reqEvil := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	reqEvil.Header.Set("Origin", "https://evil-attacker.example.com")
	wEvil := httptest.NewRecorder()
	r.ServeHTTP(wEvil, reqEvil)
	if wEvil.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for untrusted origin OPTIONS, got %d", wEvil.Code)
	}
	if got := wEvil.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin for untrusted origin, got %q", got)
	}
}

func TestBruteForceRateLimitingAndSpoofedXFF(t *testing.T) {
	// Configure strict burst of 5 requests
	r := newTestRouter(t, 1, 5)

	// First 5 requests from client IP should succeed
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		req.RemoteAddr = "203.0.113.42:54321"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d expected 200 OK, got %d", i, w.Code)
		}
	}

	// 6th request with a SPOOFED X-Forwarded-For header must STILL be blocked (429)
	reqBlocked := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	reqBlocked.RemoteAddr = "203.0.113.42:54321"
	reqBlocked.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", 99))
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)

	if wBlocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests when exceeding burst (even with spoofed X-Forwarded-For), got %d", wBlocked.Code)
	}
	if got := wBlocked.Header().Get("Retry-After"); got != "1" {
		t.Errorf("expected Retry-After=1 header on 429 response, got %q", got)
	}
}

func TestOversizedBodyRejection(t *testing.T) {
	r := newTestRouter(t, 100, 200)

	oversized := bytes.Repeat([]byte("A"), MaxRequestBodyBytes+1024)
	req := httptest.NewRequest(http.MethodGet, "/api/health", bytes.NewReader(oversized))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 Request Entity Too Large for >64KB payload, got %d", w.Code)
	}
}

func TestProductionConfigSecretValidation(t *testing.T) {
	cfgWeak := config.Load()
	cfgWeak.GinMode = "release"
	if err := cfgWeak.Validate(); err == nil {
		t.Fatal("expected Validate() to fail when release mode uses default dev JWT secrets")
	}

	cfgSame := config.Load()
	cfgSame.GinMode = "release"
	cfgSame.JWTAccessSecret = "0123456789abcdef0123456789abcdef0123456789abcdef"
	cfgSame.JWTRefreshSecret = "0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := cfgSame.Validate(); err == nil {
		t.Fatal("expected Validate() to fail when access and refresh secrets are identical")
	}

	cfgStrong := config.Load()
	cfgStrong.GinMode = "release"
	cfgStrong.JWTAccessSecret = "access-secret-9f8e7d6c5b4a3928172635445566778899"
	cfgStrong.JWTRefreshSecret = "refresh-secret-112233445566778899aabbccddeeff0011"
	if err := cfgStrong.Validate(); err != nil {
		t.Fatalf("expected Validate() to pass with strong distinct secrets, got error: %v", err)
	}
}
