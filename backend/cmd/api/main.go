package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"learn_kids/backend/internal/config"
	"learn_kids/backend/internal/db"
	"learn_kids/backend/internal/domain/chat"
	"learn_kids/backend/internal/domain/comments"
	"learn_kids/backend/internal/domain/lessons"
	"learn_kids/backend/internal/domain/progress"
	"learn_kids/backend/internal/domain/users"
	"learn_kids/backend/internal/server"
)

func main() {
	cfg := config.Load()

	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	srv := server.New(server.Deps{
		Pool:      pool,
		Lessons:   lessons.NewHandler(lessons.NewRepo(pool)),
		Users:     users.NewHandler(users.NewRepo(pool), cfg.JWTSecret),
		Progress:  progress.NewHandler(progress.NewRepo(pool)),
		Chat:      chat.NewHandler(cfg.GeminiAPIKey, chat.NewRepo(pool)),
		Comments:  comments.NewHandler(comments.NewRepo(pool)),
		JWTSecret: cfg.JWTSecret,
	})

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)

	httpSrv := &http.Server{Addr: addr, Handler: srv.Handler()}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	fmt.Println("bye")
}
