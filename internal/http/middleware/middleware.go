package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequestLogger creates a middleware that logs HTTP requests
func RequestLogger(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		entry := logger.Logger.WithFields(map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"duration":    duration.String(),
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error("Request completed with errors")
		} else {
			entry.Info("Request completed")
		}
	}
}

// Recovery creates a middleware that recovers from panics
func Recovery(logger *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Logger.WithField("panic", recovered).Error("Panic recovered")
		c.JSON(500, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal server error",
			},
		})
	})
}

// CORS creates a middleware that handles CORS
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID creates a middleware that adds a request ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return strings.ReplaceAll(time.Now().Format("20060102150405.000000"), ".", "")
}

// SupabaseUser represents the user information from Supabase JWT
type SupabaseUser struct {
	ID    string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SupabaseClaims represents the JWT claims structure from Supabase
type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SupabaseAuth creates a middleware that validates Supabase JWT tokens
func SupabaseAuth(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Logger.Debug("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authorization token",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.Logger.Debug("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		// Validate and parse the JWT token
		user, err := validateSupabaseToken(tokenString, logger)
		if err != nil {
			logger.Logger.WithError(err).Debug("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid token",
				},
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// validateSupabaseToken validates a Supabase JWT token
func validateSupabaseToken(tokenString string, logger *logger.Logger) (*SupabaseUser, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET not configured")
	}

	logger.Logger.WithFields(map[string]interface{}{
		"token_length":  len(tokenString),
		"secret_length": len(jwtSecret),
		"token_prefix":  tokenString[:min(20, len(tokenString))],
	}).Debug("Validating JWT token")

	// Parse and validate the token using ParseWithClaims and raw JWT secret
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Logger.WithField("signing_method", token.Header["alg"]).Error("Unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use raw JWT secret as bytes (no base64 decoding)
		return []byte(jwtSecret), nil
	})

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to parse JWT token")
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		logger.Logger.Error("JWT token is invalid")
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok {
		logger.Logger.Error("Failed to parse token claims")
		return nil, fmt.Errorf("failed to parse token claims")
	}

	logger.Logger.WithField("subject", claims.Subject).Debug("Successfully parsed JWT claims")

	// Check if token is expired using time comparison
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		logger.Logger.WithField("exp_time", claims.ExpiresAt.Time).Error("JWT token has expired")
		return nil, fmt.Errorf("token is expired")
	}

	// Check if token is issued in the future
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(time.Now()) {
		logger.Logger.WithField("issued_at", claims.IssuedAt.Time).Error("JWT token used before issued")
		return nil, fmt.Errorf("token used before issued")
	}

	// Check if token is not valid yet
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		logger.Logger.WithField("not_before", claims.NotBefore.Time).Error("JWT token used before valid")
		return nil, fmt.Errorf("token used before valid")
	}

	// Create user from claims
	user := &SupabaseUser{
		ID:    claims.Subject,
		Email: claims.Email,
		Role:  claims.Role,
	}

	// Set default role if not present
	if user.Role == "" {
		user.Role = "authenticated"
		logger.Logger.Debug("No role claim found, using default 'authenticated'")
	}

	logger.Logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	}).Info("Successfully validated JWT token")

	return user, nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// OptionalAuth creates a middleware that optionally validates Supabase JWT tokens
// If a token is present, it validates it and sets user context
// If no token is present, it continues without setting user context
func OptionalAuth(logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// Invalid format, but continue without authentication
			c.Next()
			return
		}

		// Try to validate the token
		user, err := validateSupabaseToken(tokenString, logger)
		if err != nil {
			logger.Logger.WithError(err).Debug("Optional auth token validation failed")
			// Continue without authentication
			c.Next()
			return
		}

		// Set user information in context if validation succeeded
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
// This should be used after SupabaseAuth middleware
func RequireRole(role string, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			logger.Logger.Debug("No user role found in context")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		if userRole != role {
			logger.Logger.WithFields(map[string]interface{}{
				"required_role": role,
				"user_role":     userRole,
			}).Debug("Insufficient permissions")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from the Gin context
func GetUserFromContext(c *gin.Context) (*SupabaseUser, bool) {
	if user, exists := c.Get("user"); exists {
		if supabaseUser, ok := user.(*SupabaseUser); ok {
			return supabaseUser, true
		}
	}
	return nil, false
}

// GetUserIDFromContext retrieves the authenticated user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id, true
		}
	}
	return "", false
}
