package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"learn_kids/backend/internal/config"
	"learn_kids/backend/internal/db"
	"learn_kids/backend/internal/domain/chat"
	"learn_kids/backend/internal/domain/comments"
	"learn_kids/backend/internal/domain/lessons"
	"learn_kids/backend/internal/domain/progress"
	"learn_kids/backend/internal/domain/users"
	"learn_kids/backend/internal/server"
)

// defaultSystemPrompt is used when no lesson context is available.
const defaultSystemPrompt = "Ты — дружелюбный ИИ-помощник для детского образовательного проекта. Помогай ученику разобраться в теме, объясняй простым языком, подходящим для детей."

func main() {
	// Structured JSON logging for production observability.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Опционально загружаем .env — три пути покрывают разные CWD:
	//   .env       — CWD = корень проекта
	//   ../.env    — CWD = backend/
	//   ../../.env — CWD = backend/cmd/api/
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	lessonsRepo := lessons.NewRepo(pool)

	// Build system prompt for AI chat on the server side (prevents prompt injection).
	lessonContextFn := func(ctx context.Context, lessonID uuid.UUID) string {
		if lessonID == uuid.Nil {
			return defaultSystemPrompt
		}
		lesson, err := lessonsRepo.GetLessonByID(ctx, lessonID)
		if err != nil || lesson == nil {
			return defaultSystemPrompt
		}
		return fmt.Sprintf(
			"Ты — дружелюбный ИИ-помощник для детского образовательного проекта. "+
				"Сейчас ученик изучает урок «%s». %s "+
				"Помогай разобраться в теме урока, объясняй простым языком, подходящим для детей. "+
				"Не отвечай на вопросы, не связанные с темой урока.",
			lesson.Title, lesson.Description,
		)
	}

	srv := server.New(server.Deps{
		Pool:           pool,
		Lessons:        lessons.NewHandler(lessonsRepo),
		Users:          users.NewHandler(users.NewRepo(pool), cfg.JWTSecret, cfg.TeacherInviteCode),
		Progress:       progress.NewHandler(progress.NewRepo(pool)),
		Chat:           chat.NewHandler(cfg.GeminiAPIKey, chat.NewRepo(pool), lessonContextFn),
		Comments:       comments.NewHandler(comments.NewRepo(pool)),
		JWTSecret:      cfg.JWTSecret,
		FrontendOrigin: cfg.FrontendOrigin,
	})

	addr := ":" + cfg.Port
	slog.Info("server started", "addr", addr)

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 90 * time.Second, // increased for Gemini AI proxy (up to 60s)
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
