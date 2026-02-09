package config

import (
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	FrontendOrigin string
	GeminiAPIKey   string
}

func Load() *Config {
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
	}
}
