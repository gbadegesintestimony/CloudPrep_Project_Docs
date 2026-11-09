package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	devDefaultAccessSecret  = "cloudprep-dev-access-secret-2026"
	devDefaultRefreshSecret = "cloudprep-dev-refresh-secret-2026"
)

type Config struct {
	Port               string
	GinMode            string
	FrontendOrigins    []string
	TrustedProxies     []string
	DatabaseURL        string
	JWTAccessSecret    string
	JWTRefreshSecret   string
	AIPrimaryProvider  string
	AIFallbackProvider string
	GeminiAPIKey       string
	GeminiModelID      string
	GroqAPIKey         string
	GroqModelID        string
	RateLimitRPS       float64
	RateLimitBurst     int
	RateLimitAuthRPS   float64
	RateLimitAuthBurst int
}

func Load() *Config {
	loadDotEnv(".env", "../.env")

	originsRaw := getEnv("FRONTEND_ORIGIN", "http://localhost:3000,http://127.0.0.1:3000")
	var origins []string
	for _, o := range strings.Split(originsRaw, ",") {
		trimmed := strings.TrimSpace(o)
		// Never allow wildcard "*" in credentialed CORS allowlist
		if trimmed != "" && trimmed != "*" {
			origins = append(origins, trimmed)
		}
	}

	var trustedProxies []string
	if rawProxies := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); rawProxies != "" {
		for _, p := range strings.Split(rawProxies, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				trustedProxies = append(trustedProxies, trimmed)
			}
		}
	} else if os.Getenv("RENDER") != "" {
		// Render internal proxy network
		trustedProxies = []string{"10.0.0.0/8", "127.0.0.1/32"}
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		FrontendOrigins:    origins,
		TrustedProxies:     trustedProxies,
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTAccessSecret:    getEnv("JWT_ACCESS_SECRET", devDefaultAccessSecret),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", devDefaultRefreshSecret),
		AIPrimaryProvider:  strings.ToLower(getEnv("AI_PRIMARY_PROVIDER", "gemini")),
		AIFallbackProvider: strings.ToLower(getEnv("AI_FALLBACK_PROVIDER", "groq")),
		GeminiAPIKey:       strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		GeminiModelID:      getEnv("GEMINI_MODEL_ID", "gemini-3.8-flash"),
		GroqAPIKey:         strings.TrimSpace(os.Getenv("GROQ_API_KEY")),
		GroqModelID:        getEnv("GROQ_MODEL_ID", "openai/gpt-oss-120b"),
		RateLimitRPS:       20.0,
		RateLimitBurst:     40,
		RateLimitAuthRPS:   0.2, // ~1 request every 5s per IP
		RateLimitAuthBurst: 5,   // burst of 5 auth attempts
	}
}

// Validate enforces production security invariants so the server refuses to start
// in release/production mode with weak or default JWT secrets.
func (c *Config) Validate() error {
	isProd := strings.EqualFold(c.GinMode, "release") ||
		strings.EqualFold(os.Getenv("APP_ENV"), "production") ||
		os.Getenv("RENDER") != ""

	if isProd {
		if c.JWTAccessSecret == "" || c.JWTAccessSecret == devDefaultAccessSecret || len(c.JWTAccessSecret) < 32 {
			return errors.New("CRITICAL: JWT_ACCESS_SECRET must be set to a strong secret (>= 32 chars) in production/release mode")
		}
		if c.JWTRefreshSecret == "" || c.JWTRefreshSecret == devDefaultRefreshSecret || len(c.JWTRefreshSecret) < 32 {
			return errors.New("CRITICAL: JWT_REFRESH_SECRET must be set to a strong secret (>= 32 chars) in production/release mode")
		}
		if c.JWTAccessSecret == c.JWTRefreshSecret {
			return fmt.Errorf("CRITICAL: JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be distinct in production/release mode")
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}

// loadDotEnv scans candidate file paths for a .env file and sets unpopulated environment variables.
func loadDotEnv(paths ...string) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		break
	}
}
