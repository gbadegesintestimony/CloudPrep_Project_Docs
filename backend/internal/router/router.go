package router

import (
	"cloudprep-backend/internal/ai"
	"cloudprep-backend/internal/auth"
	"cloudprep-backend/internal/config"
	"cloudprep-backend/internal/curriculum"
	"cloudprep-backend/internal/handlers"
	"cloudprep-backend/internal/middleware"
	"cloudprep-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const MaxRequestBodyBytes = 64 * 1024 // 64 KB cap per request

func SetupRouter(cfg *config.Config, database *gorm.DB, dbDriver string) *gin.Engine {
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Prevent X-Forwarded-For spoofing unless explicit trusted proxy CIDRs are configured
	if len(cfg.TrustedProxies) > 0 {
		_ = r.SetTrustedProxies(cfg.TrustedProxies)
	} else {
		_ = r.SetTrustedProxies(nil)
	}

	// Global defensive middlewares
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.StrictCORSMiddleware(cfg.FrontendOrigins))
	r.Use(middleware.MaxBodyBytesMiddleware(MaxRequestBodyBytes))

	rateLimiter := middleware.NewIPRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	r.Use(rateLimiter.Middleware())

	// Initialize Dual-AI Generation Engine (Gemini + Groq Failover)
	geminiProvider := ai.NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiModelID)
	groqProvider := ai.NewGroqProvider(cfg.GroqAPIKey, cfg.GroqModelID)

	var primaryProvider, fallbackProvider ai.QuestionGenerationProvider
	if cfg.AIPrimaryProvider == "groq" {
		primaryProvider = groqProvider
		fallbackProvider = geminiProvider
	} else {
		primaryProvider = geminiProvider
		fallbackProvider = groqProvider
	}

	aiRouter := ai.NewAiRouter(database, primaryProvider, fallbackProvider, curriculum.GetCurriculumTopics())
	replenishmentWorker := ai.NewReplenishmentWorker(aiRouter)

	healthHandler := handlers.NewHealthHandler(database, dbDriver)
	authHandler := handlers.NewAuthHandler(database, cfg)
	sessionHandler := handlers.NewSessionHandler(database)
	sessionHandler.OnLowUnseen = replenishmentWorker.Hook()
	bookmarkProgressHandler := handlers.NewBookmarkProgressHandler(database)
	adminHandler := handlers.NewAdminHandler(database, aiRouter)

	api := r.Group("/api")
	{
		// Public endpoints
		api.GET("/health", healthHandler.GetHealth)
		api.GET("/curriculum", healthHandler.GetCurriculum)

		// Public Zero-Friction Auth & 1-Click Guest endpoints with dedicated brute-force rate limiting (H-3)
		authGroup := api.Group("/auth")
		{
			authRateLimiter := middleware.NewIPRateLimiter(cfg.RateLimitAuthRPS, cfg.RateLimitAuthBurst)
			authGroup.Use(authRateLimiter.Middleware())

			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/guest", authHandler.GuestLogin)
			authGroup.POST("/refresh", authHandler.Refresh)
		}

		// Authenticated endpoints (Learner, Guest, Admin)
		protected := api.Group("")
		protected.Use(auth.AuthMiddleware(cfg.JWTAccessSecret))
		{
			protected.GET("/me", authHandler.GetMe)
			protected.GET("/auth/me", authHandler.GetMe)

			// Practice & Mock Exam Session routes (/api/practice-sessions and /api/sessions alias)
			for _, prefix := range []string{"/practice-sessions", "/sessions"} {
				sg := protected.Group(prefix)
				{
					sg.POST("", sessionHandler.CreateSession)
					sg.GET("", sessionHandler.ListSessions)
					sg.POST("/cancel-active", sessionHandler.CancelAllActiveSessions)
					sg.GET("/:sessionId", sessionHandler.GetSession)
					sg.POST("/:sessionId/cancel", sessionHandler.CancelSession)
					sg.DELETE("/:sessionId", sessionHandler.CancelSession)
					sg.PUT("/:sessionId/responses/:questionId", sessionHandler.UpsertResponse)
					sg.POST("/:sessionId/responses/:questionId", sessionHandler.UpsertResponse)
					sg.PUT("/:sessionId/questions/:questionId/flag", sessionHandler.ToggleFlag)
					sg.POST("/:sessionId/submit", sessionHandler.SubmitSession)
					sg.GET("/:sessionId/review", sessionHandler.GetSessionReview)
				}
			}

			// Bookmarks & Progress Analytics routes
			protected.GET("/bookmarks", bookmarkProgressHandler.ListBookmarks)
			protected.POST("/bookmarks/:questionId", bookmarkProgressHandler.UpsertBookmark)
			protected.DELETE("/bookmarks/:questionId", bookmarkProgressHandler.DeleteBookmark)
			protected.GET("/progress/summary", bookmarkProgressHandler.GetProgressSummary)

			// Admin Question Bank & Batch AI Generation routes
			admin := protected.Group("/admin")
			admin.Use(auth.RequireRoles(models.RoleAdmin))
			{
				admin.GET("/questions", adminHandler.ListQuestions)
				admin.PATCH("/questions/:questionId", adminHandler.UpdateQuestionStatus)
				admin.GET("/ai/records", adminHandler.ListGenerationRecords)
				admin.POST("/ai/generate", adminHandler.BatchGenerate)
				admin.GET("/stats", adminHandler.GetBankStats)
			}
		}
	}

	return r
}
