package users

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHandler_GenerateAndParseToken(t *testing.T) {
	h := NewHandler(nil, "test-secret-key")
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	role := RoleStudent

	token, err := h.generateToken(userID, role, false)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	parsedID, parsedRole, mustChange, expiresAt, err := h.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if parsedID != userID {
		t.Errorf("userID = %v, want %v", parsedID, userID)
	}
	if parsedRole != role {
		t.Errorf("role = %q, want %q", parsedRole, role)
	}
	if mustChange != false {
		t.Errorf("mustChangePassword = %v, want false", mustChange)
	}
	if expiresAt.IsZero() {
		t.Error("expiresAt should be set for valid token")
	}
}

func TestHandler_GenerateAndParseToken_MustChangePassword(t *testing.T) {
	h := NewHandler(nil, "test-secret-key")
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111112")

	token, err := h.generateToken(userID, RoleAdministrator, true)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	_, _, mustChange, _, err := h.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if !mustChange {
		t.Error("mustChangePassword = false, want true")
	}
}

func TestHandler_ParseToken_Invalid(t *testing.T) {
	h := NewHandler(nil, "test-secret")

	_, _, _, _, err := h.ParseToken("invalid")
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(invalid) = %v, want ErrInvalidToken", err)
	}

	_, _, _, _, err = h.ParseToken("")
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(empty) = %v, want ErrInvalidToken", err)
	}
}

func TestHandler_ParseToken_WrongSecret(t *testing.T) {
	h1 := NewHandler(nil, "secret-a")
	token, _ := h1.generateToken(uuid.New(), RoleTeacher, false)

	h2 := NewHandler(nil, "secret-b")
	_, _, _, _, err := h2.ParseToken(token)
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(wrong secret) = %v, want ErrInvalidToken", err)
	}
}

func TestHandler_NewToken(t *testing.T) {
	h := NewHandler(nil, "test-secret")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	role := RoleTeacher

	newToken, err := h.NewToken(userID, role, false)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	parsedID, parsedRole, mustChange, expiresAt, err := h.ParseToken(newToken)
	if err != nil {
		t.Fatalf("ParseToken(newToken): %v", err)
	}
	if parsedID != userID || parsedRole != role {
		t.Errorf("got userID=%v role=%q, want userID=%v role=%q", parsedID, parsedRole, userID, role)
	}
	if mustChange {
		t.Error("mustChangePassword should be false")
	}
	if expiresAt.IsZero() {
		t.Error("NewToken should produce token with non-zero expiresAt")
	}
}

func TestHandler_ParseToken_Expired(t *testing.T) {
	h := NewHandler(nil, "test-secret")
	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	// Build a token with exp in the past (same secret as handler).
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":              userID.String(),
		"role":                 RoleStudent,
		"must_change_password": false,
		"exp":                  time.Now().Add(-time.Minute).Unix(),
		"iat":                  time.Now().Add(-2 * time.Hour).Unix(),
		"iss":                  "learn_kids",
		"sub":                  userID.String(),
	})
	tokenStr, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	_, _, _, _, err = h.ParseToken(tokenStr)
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(expired) = %v, want ErrInvalidToken", err)
	}
}
