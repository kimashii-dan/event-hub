package service

import (
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

// TestAuthService_ValidateToken verifies that a token generated with a known secret
// is correctly validated and parsed by the AuthService.

func TestAuthService_ValidateToken(t *testing.T) {
	secret := "test-secret"
	expiry := time.Hour

	// Create a token manually
	token, err := jwt.GenerateToken("user-123", "test@example.com", "user", secret, expiry)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	service := NewAuthService(nil, secret, expiry)

	claims, err := service.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}
