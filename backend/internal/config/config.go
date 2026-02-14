package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
	GeminiAPIKey   string
}

// Load loads config from environment. Returns error if JWT_SECRET is not set in production.
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/learn_kids?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if os.Getenv("APP_ENV") == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set when APP_ENV=production")
		}
		jwtSecret = "dev-secret-change-in-production"
	}
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	return &Config{
		Port:           port,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		FrontendOrigin: frontendOrigin,
		GeminiAPIKey:   geminiAPIKey,
	}, nil
}
