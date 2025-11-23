package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
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

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
	ID    string   `json:"sub"`
	Email string   `json:"email"`
	Roles []string `json:"roles,omitempty"`
}

// SupabaseClaims represents the JWT claims structure from Supabase
type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email string   `json:"email"`
	Roles []string `json:"roles,omitempty"`
}

const supabaseTokenTimeLeeway = 5 * time.Second

// SupabaseAuth creates a middleware that validates Supabase JWT tokens
// and enriches the context with full user profile from database
func SupabaseAuth(logger *logger.Logger, userRepo repositories.UserRepository) gin.HandlerFunc {
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
		jwtUser, err := validateSupabaseToken(tokenString, logger)
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

		// Fetch full user profile from database if repository is provided
		if userRepo != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()

			userProfile, err := userRepo.GetProfileBySupabaseID(ctx, jwtUser.ID)
			if err != nil {
				logger.Logger.WithError(err).Warn("Failed to fetch user profile from database, using JWT data only")
				// Continue with JWT data only, don't abort
			} else {
				// Use database data and set additional context
				c.Set("user_profile", userProfile)
				c.Set("organization", userProfile.Organization)
				if userProfile.Profile.OrganizationID != nil {
					c.Set("organization_id", userProfile.Profile.OrganizationID.String())
				}
			}
		}

		// Set basic user information in context (from JWT)
		c.Set("user", jwtUser)
		c.Set("user_id", jwtUser.ID)
		c.Set("user_email", jwtUser.Email)
		c.Set("user_roles", jwtUser.Roles)

		c.Next()
	}
}

// SupabaseAuthSimple creates a basic middleware that only validates JWT tokens
// without database lookup for organization data
func SupabaseAuthSimple(logger *logger.Logger) gin.HandlerFunc {
	return SupabaseAuth(logger, nil)
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

	parser := jwt.NewParser(jwt.WithLeeway(supabaseTokenTimeLeeway))

	// Parse and validate the token using ParseWithClaims and raw JWT secret
	token, err := parser.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	now := time.Now()

	// Check if token is expired using time comparison
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
		logger.Logger.WithField("exp_time", claims.ExpiresAt.Time).Error("JWT token has expired")
		return nil, fmt.Errorf("token is expired")
	}

	// Check if token is issued in the future
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(now.Add(supabaseTokenTimeLeeway)) {
		logger.Logger.WithField("issued_at", claims.IssuedAt.Time).Error("JWT token used before issued")
		return nil, fmt.Errorf("token used before issued")
	}

	// Check if token is not valid yet
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now.Add(supabaseTokenTimeLeeway)) {
		logger.Logger.WithField("not_before", claims.NotBefore.Time).Error("JWT token used before valid")
		return nil, fmt.Errorf("token used before valid")
	}

	// Create user from claims
	user := &SupabaseUser{
		ID:    claims.Subject,
		Email: claims.Email,
		Roles: claims.Roles,
	}

	// Set default roles if not present
	if len(user.Roles) == 0 {
		user.Roles = []string{"authenticated"}
		logger.Logger.Debug("No roles claim found, using default 'authenticated'")
	}

	logger.Logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"roles":   user.Roles,
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
		c.Set("user_roles", user.Roles)

		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
// This should be used after SupabaseAuth middleware
func RequireRole(role string, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			logger.Logger.Debug("No user roles found in context")
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

		roles, ok := userRoles.([]string)
		if !ok {
			logger.Logger.Debug("Invalid roles format in context")
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

		// Check if user has the required role
		hasRole := false
		for _, userRole := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			logger.Logger.WithFields(map[string]interface{}{
				"required_role": role,
				"user_roles":    roles,
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

// GetUserRolesFromContext retrieves the authenticated user roles from the Gin context
func GetUserRolesFromContext(c *gin.Context) ([]string, bool) {
	if userRoles, exists := c.Get("user_roles"); exists {
		if roles, ok := userRoles.([]string); ok {
			return roles, true
		}
	}
	return nil, false
}

// HasRole checks if the user has a specific role
func HasRole(c *gin.Context, role string) bool {
	roles, exists := GetUserRolesFromContext(c)
	if !exists {
		return false
	}

	for _, userRole := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// GetOrganizationIDFromContext retrieves the organization ID from the Gin context
func GetOrganizationIDFromContext(c *gin.Context) (string, bool) {
	if orgID, exists := c.Get("organization_id"); exists {
		if id, ok := orgID.(string); ok && id != "" {
			return id, true
		}
	}
	return "", false
}

// GetUserProfileFromContext retrieves the full user profile from the Gin context
func GetUserProfileFromContext(c *gin.Context) (*entities.UserProfile, bool) {
	if profile, exists := c.Get("user_profile"); exists {
		if userProfile, ok := profile.(*entities.UserProfile); ok {
			return userProfile, true
		}
	}
	return nil, false
}

// GetOrganizationFromContext retrieves the organization from the Gin context
func GetOrganizationFromContext(c *gin.Context) (*entities.Organization, bool) {
	if org, exists := c.Get("organization"); exists {
		if organization, ok := org.(*entities.Organization); ok {
			return organization, true
		}
	}
	return nil, false
}

// CustomCORS is a middleware to handle dynamic CORS origins
func CustomCORS(allowedOrigins []string) gin.HandlerFunc {
	// Compile regex for dynamic subdomains
	dynamicOriginRegex := regexp.MustCompile(`^https://dental-scheduler-front-[a-zA-Z0-9-]+\.vercel\.app$`)

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))

		regexMatched := dynamicOriginRegex.MatchString(origin)
		explicitMatch := false

		// Check if the origin matches the dynamic pattern
		if regexMatched {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Check if the origin is in the allowed list
			for _, o := range allowedOrigins {
				if strings.TrimSpace(o) == origin {
					explicitMatch = true
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// Log the origin evaluation for easier debugging during deployments
		log.Printf("CustomCORS origin=%q regexMatch=%t explicitMatch=%t", origin, regexMatched, explicitMatch)

		// Set other CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
