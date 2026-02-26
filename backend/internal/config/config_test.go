package config

import (
	"testing"
)

func TestLoad_Production_RequiresJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("FRONTEND_ORIGIN", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail when APP_ENV=production and JWT_SECRET is unset")
	}
	if err.Error() != "JWT_SECRET must be set when APP_ENV=production" {
		t.Errorf("err = %v; want JWT_SECRET must be set when APP_ENV=production", err)
	}
}

func TestLoad_Production_RequiresFrontendOrigin(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "a-secret-with-at-least-32-characters-long")
	t.Setenv("FRONTEND_ORIGIN", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail when APP_ENV=production and FRONTEND_ORIGIN is unset")
	}
	if err.Error() != "FRONTEND_ORIGIN must be set when APP_ENV=production" {
		t.Errorf("err = %v; want FRONTEND_ORIGIN must be set when APP_ENV=production", err)
	}
}

func TestLoad_Production_RejectsShortJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "short")
	t.Setenv("FRONTEND_ORIGIN", "https://example.com")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail when APP_ENV=production and JWT_SECRET is shorter than 32 chars")
	}
	if err.Error() != "JWT_SECRET must be at least 32 characters in production" {
		t.Errorf("err = %v; want JWT_SECRET must be at least 32 characters in production", err)
	}
}

func TestLoad_Production_SucceedsWithRequiredVars(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "a-secret-with-at-least-32-characters-long")
	t.Setenv("FRONTEND_ORIGIN", "https://app.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() err = %v", err)
	}
	if cfg.JWTSecret != "a-secret-with-at-least-32-characters-long" {
		t.Errorf("JWTSecret = %q", cfg.JWTSecret)
	}
	if cfg.FrontendOrigin != "https://app.example.com" {
		t.Errorf("FrontendOrigin = %q", cfg.FrontendOrigin)
	}
}

func TestLoad_Dev_Defaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("FRONTEND_ORIGIN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() err = %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q; want 8080", cfg.Port)
	}
	if cfg.FrontendOrigin != "http://localhost:5173" {
		t.Errorf("FrontendOrigin = %q; want http://localhost:5173", cfg.FrontendOrigin)
	}
	if cfg.JWTSecret != "dev-secret-change-in-production" {
		t.Errorf("JWTSecret = %q; want dev default", cfg.JWTSecret)
	}
}
