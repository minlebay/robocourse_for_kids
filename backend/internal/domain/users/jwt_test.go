package users

import (
	"testing"

	"github.com/google/uuid"
)

func TestHandler_GenerateAndParseToken(t *testing.T) {
	h := NewHandler(nil, "test-secret-key", "")
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	role := RoleStudent

	token, err := h.generateToken(userID, role)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	parsedID, parsedRole, err := h.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if parsedID != userID {
		t.Errorf("userID = %v, want %v", parsedID, userID)
	}
	if parsedRole != role {
		t.Errorf("role = %q, want %q", parsedRole, role)
	}
}

func TestHandler_ParseToken_Invalid(t *testing.T) {
	h := NewHandler(nil, "test-secret", "")

	_, _, err := h.ParseToken("invalid")
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(invalid) = %v, want ErrInvalidToken", err)
	}

	_, _, err = h.ParseToken("")
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(empty) = %v, want ErrInvalidToken", err)
	}
}

func TestHandler_ParseToken_WrongSecret(t *testing.T) {
	h1 := NewHandler(nil, "secret-a", "")
	token, _ := h1.generateToken(uuid.New(), RoleTeacher)

	h2 := NewHandler(nil, "secret-b", "")
	_, _, err := h2.ParseToken(token)
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(wrong secret) = %v, want ErrInvalidToken", err)
	}
}
