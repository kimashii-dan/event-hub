package jwt_test

import (
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/pkg/jwt"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"
	userID := "123"
	email := "test@example.com"
	role := "user"

	token, err := jwt.GenerateToken(userID, email, role, secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := jwt.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	secret := "test-secret"
	userID := "123"
	email := "test@example.com"
	role := "user"

	token, err := jwt.GenerateToken(userID, email, role, secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	_, err = jwt.ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Errorf("expected error when validating token with wrong secret, got nil")
	}
}
