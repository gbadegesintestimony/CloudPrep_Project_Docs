package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	ContextUserIDKey    = "auth_user_id"
	ContextUserEmailKey = "auth_user_email"
	ContextUserRoleKey  = "auth_user_role"
	ContextIsGuestKey   = "auth_is_guest"

	AccessTokenTTL  = 1 * time.Hour
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// Precomputed dummy bcrypt hash used when an email is not found during login
// to prevent timing side-channel email enumeration attacks.
var dummyBcryptHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("cloudprep-constant-time-dummy-password"), bcrypt.DefaultCost)
	if err == nil {
		dummyBcryptHash = h
	}
}

type Claims struct {
	Sub       string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	IsGuest   bool   `json:"isGuest"`
	TokenType string `json:"type"` // "access" or "refresh"
	Iat       int64  `json:"iat"`
	Exp       int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// GenerateID creates a cryptographically random prefixed identifier.
func GenerateID(prefix string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword Constant-time password comparison with dummy fallback when hash is empty.
func CheckPassword(passwordHash, candidatePassword string) bool {
	if passwordHash == "" {
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(candidatePassword))
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(candidatePassword))
	return err == nil
}

// SignToken signs an HS256 JWT token using the provided secret.
func SignToken(claims Claims, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("JWT secret cannot be empty")
	}
	headerBytes, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := encodedHeader + "." + encodedPayload

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

// VerifyToken verifies an HS256 JWT token's format, algorithm, HMAC signature, and expiration.
func VerifyToken(tokenStr, secret, expectedType string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed JWT token")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("invalid JWT header encoding")
	}
	var hdr jwtHeader
	if err := json.Unmarshal(headerBytes, &hdr); err != nil || hdr.Alg != "HS256" {
		return nil, errors.New("unsupported JWT signing algorithm")
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid JWT signature encoding")
	}
	if !hmac.Equal(actualSig, expectedSig) {
		return nil, errors.New("invalid JWT signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid JWT payload encoding")
	}
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("invalid JWT claims JSON")
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s", expectedType)
	}
	if time.Now().UTC().Unix() > claims.Exp {
		return nil, errors.New("JWT token has expired")
	}
	if claims.Sub == "" {
		return nil, errors.New("missing subject claim")
	}

	return &claims, nil
}

// AuthMiddleware verifies the Bearer JWT token and populates user identity in Gin context.
func AuthMiddleware(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Authentication required. Please sign in or continue as guest.",
			})
			return
		}

		claims, err := VerifyToken(tokenStr, accessSecret, "access")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN",
				"message": "Invalid or expired authentication token.",
			})
			return
		}

		c.Set(ContextUserIDKey, claims.Sub)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextUserRoleKey, claims.Role)
		c.Set(ContextIsGuestKey, claims.IsGuest)
		c.Next()
	}
}

// RequireRoles restricts an endpoint to users holding one of the allowed roles.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextUserRoleKey)
		roleStr, _ := roleVal.(string)
		if !exists || !allowed[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": "You do not have permission to access this resource.",
			})
			return
		}
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}

