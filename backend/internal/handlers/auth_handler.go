package handlers

import (
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"cloudprep-backend/internal/auth"
	"cloudprep-backend/internal/config"
	"cloudprep-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	if !emailRegex.MatchString(email) {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	return true
}

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAuthHandler(database *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		DB:  database,
		Cfg: cfg,
	}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type guestRequest struct {
	Name string `json:"name"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) issueTokens(user models.User) (string, string, error) {
	now := time.Now().UTC()
	accessClaims := auth.Claims{
		Sub:       user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		IsGuest:   user.IsGuest,
		TokenType: "access",
		Iat:       now.Unix(),
		Exp:       now.Add(auth.AccessTokenTTL).Unix(),
	}
	accessToken, err := auth.SignToken(accessClaims, h.Cfg.JWTAccessSecret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := auth.Claims{
		Sub:       user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		IsGuest:   user.IsGuest,
		TokenType: "refresh",
		Iat:       now.Unix(),
		Exp:       now.Add(auth.RefreshTokenTTL).Unix(),
	}
	refreshToken, err := auth.SignToken(refreshClaims, h.Cfg.JWTRefreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Register handles POST /api/auth/register (Zero-Friction Registration — instant JWT issuance, no email verification wall).
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Please provide valid registration fields (name, email, password).",
		})
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if len(name) < 2 || len(name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "VALIDATION_ERROR",
			"message": "Name must be between 2 and 100 characters.",
		})
		return
	}
	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "VALIDATION_ERROR",
			"message": "Please enter a valid email address (e.g. name@domain.com).",
		})
		return
	}
	passBytes := len([]byte(password))
	if passBytes < 8 || passBytes > 72 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "VALIDATION_ERROR",
			"message": "Password must be between 8 and 72 characters.",
		})
		return
	}

	// M-2 Constant-Time DX: Hash password before checking DB existence to eliminate
	// timing side-channel enumeration between registered and unregistered emails.
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to secure password.",
		})
		return
	}

	var existing models.User
	if err := h.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "EMAIL_ALREADY_REGISTERED",
			"message": "An account with this email already exists. Please sign in instead.",
		})
		return
	}

	now := time.Now().UTC()
	user := models.User{
		ID:           auth.GenerateID("usr"),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		IsGuest:      false,
		Role:         models.RoleLearner,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "EMAIL_ALREADY_REGISTERED",
				"message": "An account with this email already exists. Please sign in instead.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Could not create user account.",
		})
		return
	}

	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "TOKEN_ERROR",
			"message": "Could not issue authentication token.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":        accessToken,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	})
}

// Login handles POST /api/auth/login with constant-time password verification.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "Please provide email and password.",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "VALIDATION_ERROR",
			"message": "Please enter a valid email address.",
		})
		return
	}

	// Rejection for passwords longer than 72 bytes prevents bcrypt error / DoS
	if len([]byte(password)) > 72 || len([]byte(password)) < 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "INVALID_CREDENTIALS",
			"message": "Invalid email or password.",
		})
		return
	}

	var user models.User
	err := h.DB.Where("email = ? AND is_guest = ?", email, false).First(&user).Error
	if err != nil {
		// Execute dummy bcrypt comparison to prevent timing side-channel email enumeration
		_ = auth.CheckPassword("", password)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "INVALID_CREDENTIALS",
			"message": "Invalid email or password.",
		})
		return
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "INVALID_CREDENTIALS",
			"message": "Invalid email or password.",
		})
		return
	}

	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "TOKEN_ERROR",
			"message": "Could not issue authentication token.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        accessToken,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	})
}

// GuestLogin handles POST /api/auth/guest (1-Click Guest Learner creation with unique synthetic email).
func (h *AuthHandler) GuestLogin(c *gin.Context) {
	var req guestRequest
	_ = c.ShouldBindJSON(&req)

	guestID := auth.GenerateID("guest")
	displayName := strings.TrimSpace(req.Name)
	if displayName == "" || len(displayName) > 80 {
		suffix := strings.ToUpper(guestID[len(guestID)-4:])
		displayName = fmt.Sprintf("Guest Cloud Engineer #%s", suffix)
	}

	now := time.Now().UTC()
	syntheticEmail := fmt.Sprintf("%s@guest.cloudprep.local", guestID)

	user := models.User{
		ID:           guestID,
		Name:         displayName,
		Email:        syntheticEmail,
		PasswordHash: "",
		IsGuest:      true,
		Role:         models.RoleGuest,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to initialize guest session.",
		})
		return
	}

	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "TOKEN_ERROR",
			"message": "Failed to issue guest token.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":        accessToken,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	})
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_PAYLOAD",
			"message": "refreshToken is required.",
		})
		return
	}

	claims, err := auth.VerifyToken(strings.TrimSpace(req.RefreshToken), h.Cfg.JWTRefreshSecret, "refresh")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "INVALID_REFRESH_TOKEN",
			"message": "Refresh token is invalid or expired.",
		})
		return
	}

	var user models.User
	if err := h.DB.Where("id = ?", claims.Sub).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "USER_NOT_FOUND",
			"message": "Associated user account no longer exists.",
		})
		return
	}

	// M-4: Reject token refresh if user account is deactivated or suspended
	if user.Status != "ACTIVE" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "USER_INACTIVE",
			"message": "User account is suspended or inactive.",
		})
		return
	}

	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "TOKEN_ERROR",
			"message": "Could not refresh token.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        accessToken,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	})
}

// GetMe handles GET /api/me (and GET /api/auth/me).
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	var user models.User
	if err := h.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "USER_NOT_FOUND",
			"message": "User profile not found.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

