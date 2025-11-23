package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	infraLogger "dental-scheduler-backend/internal/infra/logger"
)

const testJWTSecret = "test-secret"

func signSupabaseToken(t *testing.T, claims *SupabaseClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return signed
}

func TestValidateSupabaseTokenAllowsIssuedAtLeeway(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testJWTSecret)
	logger := infraLogger.NewLogger("debug")

	now := time.Now()
	claims := &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-issued-at-leeway",
			IssuedAt:  jwt.NewNumericDate(now.Add(3 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
		Email: "leeway@example.com",
	}

	tokenString := signSupabaseToken(t, claims)

	user, err := validateSupabaseToken(tokenString, logger)
	if err != nil {
		t.Fatalf("expected token to be valid within leeway, got error: %v", err)
	}

	if user == nil || user.ID != "user-issued-at-leeway" {
		t.Fatalf("unexpected user returned: %+v", user)
	}
}

func TestValidateSupabaseTokenRejectsIssuedAtBeyondLeeway(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testJWTSecret)
	logger := infraLogger.NewLogger("debug")

	now := time.Now()
	claims := &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-issued-at-too-far",
			IssuedAt:  jwt.NewNumericDate(now.Add(10 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	tokenString := signSupabaseToken(t, claims)

	if _, err := validateSupabaseToken(tokenString, logger); err == nil {
		t.Fatal("expected token to be rejected when issued-at exceeds leeway")
	}
}

func TestValidateSupabaseTokenAllowsNotBeforeLeeway(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testJWTSecret)
	logger := infraLogger.NewLogger("debug")

	now := time.Now()
	claims := &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-nbf-leeway",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(3 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	tokenString := signSupabaseToken(t, claims)

	if _, err := validateSupabaseToken(tokenString, logger); err != nil {
		t.Fatalf("expected token to be valid within not-before leeway, got error: %v", err)
	}
}

func TestValidateSupabaseTokenRejectsNotBeforeBeyondLeeway(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", testJWTSecret)
	logger := infraLogger.NewLogger("debug")

	now := time.Now()
	claims := &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-nbf-too-far",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(10 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	tokenString := signSupabaseToken(t, claims)

	if _, err := validateSupabaseToken(tokenString, logger); err == nil {
		t.Fatal("expected token to be rejected when not-before exceeds leeway")
	}
}
