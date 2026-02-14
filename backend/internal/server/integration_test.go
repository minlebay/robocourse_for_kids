package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"learn_kids/backend/internal/db"
	"learn_kids/backend/internal/domain/chat"
	"learn_kids/backend/internal/domain/comments"
	"learn_kids/backend/internal/domain/lessons"
	"learn_kids/backend/internal/domain/progress"
	"learn_kids/backend/internal/domain/users"

	_ "learn_kids/backend/internal/middleware" // ensure middleware compiles
)

var testPool *testDeps

type testDeps struct {
	srv *Server
}

func TestMain(m *testing.M) {
	// Загружаем .test.env из корня проекта или из backend (при запуске из backend или из корня)
	_ = godotenv.Load("../.test.env")
	_ = godotenv.Load(".test.env")

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/learn_kids?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		os.Exit(m.Run()) // no DB: skip integration tests (exit 0)
		return
	}
	defer pool.Close()
	if err := db.RunMigrations(url); err != nil {
		os.Exit(m.Run())
		return
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-jwt-secret"
	}
	geminiKey := os.Getenv("GEMINI_API_KEY")
	inviteCode := os.Getenv("TEACHER_INVITE_CODE")

	// Dummy lesson context func for tests
	lessonCtxFn := func(_ context.Context, _ uuid.UUID) string {
		return "Test system prompt"
	}

	testPool = &testDeps{
		srv: New(Deps{
			Pool:           pool,
			Lessons:        lessons.NewHandler(lessons.NewRepo(pool)),
			Users:          users.NewHandler(users.NewRepo(pool), jwtSecret, inviteCode),
			Progress:       progress.NewHandler(progress.NewRepo(pool)),
			Chat:           chat.NewHandler(geminiKey, chat.NewRepo(pool), lessonCtxFn),
			Comments:       comments.NewHandler(comments.NewRepo(pool)),
			JWTSecret:      jwtSecret,
			FrontendOrigin: "http://localhost:5173",
		}),
	}
	os.Exit(m.Run())
}

func TestHealth(t *testing.T) {
	if testPool == nil {
		t.Skip("no database")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	testPool.srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/v1/health: status = %d, want 200", rec.Code)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	if testPool == nil {
		t.Skip("no database")
	}
	login := "testuser_" + uuid.New().String()[:8]
	body := []byte(`{"login":"` + login + `","password":"pass123","name":"Test User","role":"student"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testPool.srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("POST register: status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"login":"`+login+`","password":"pass123"}`)))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	testPool.srv.Handler().ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("POST login: status = %d, want 200; body = %s", rec2.Code, rec2.Body.String())
	}
}

func TestListModules(t *testing.T) {
	if testPool == nil {
		t.Skip("no database")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/modules", nil)
	rec := httptest.NewRecorder()
	testPool.srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/v1/modules: status = %d, want 200", rec.Code)
	}
}
