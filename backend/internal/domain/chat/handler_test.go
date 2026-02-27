package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

// --- mock ---

type mockChatRepo struct {
	ListFn   func(ctx context.Context, userID, lessonID uuid.UUID) ([]StoredMessage, error)
	SaveFn   func(ctx context.Context, userID, lessonID uuid.UUID, role, text string) error
	DeleteFn func(ctx context.Context, userID, lessonID uuid.UUID) (int64, error)
}

func (m *mockChatRepo) ListByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) ([]StoredMessage, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userID, lessonID)
	}
	return nil, nil
}

func (m *mockChatRepo) Save(ctx context.Context, userID, lessonID uuid.UUID, role, text string) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, userID, lessonID, role, text)
	}
	return nil
}

func (m *mockChatRepo) DeleteByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) (int64, error) {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, lessonID)
	}
	return 0, nil
}

// --- helpers ---

func chatRequest(method, target string, body interface{}) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, target, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	c.Request = req
	return w, c
}

func defaultLessonCtx(_ context.Context, _ uuid.UUID) string {
	return "test system prompt"
}

func geminiOK(text string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(geminiResponse{
			Candidates: []geminiCandidate{{
				Content: geminiContent{
					Role:  "model",
					Parts: []geminiPart{{Text: text}},
				},
			}},
		})
	}
}

// ==================== Chat validation ====================

func TestChat_NoAPIKey(t *testing.T) {
	svc := NewService("", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	w, c := chatRequest(http.MethodPost, "/chat", Request{Message: "hello"})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestChat_Unauthorized(t *testing.T) {
	svc := NewService("key", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	w, c := chatRequest(http.MethodPost, "/chat", Request{Message: "hello"})
	// no user_id in context

	h.Chat(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestChat_EmptyMessage(t *testing.T) {
	svc := NewService("key", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	w, c := chatRequest(http.MethodPost, "/chat", Request{Message: ""})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestChat_MessageTooLong(t *testing.T) {
	svc := NewService("key", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	long := make([]byte, maxMessageText+1)
	for i := range long {
		long[i] = 'a'
	}
	w, c := chatRequest(http.MethodPost, "/chat", Request{Message: string(long)})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

// ==================== Chat with mock Gemini ====================

func TestChat_SuccessWithHistory(t *testing.T) {
	lessonID := uuid.New()

	geminiMock := httptest.NewServer(geminiOK("Привет! Чем помочь?"))
	defer geminiMock.Close()

	dbMessages := []StoredMessage{
		{Role: "user", Text: "old message", CreatedAt: time.Now().Add(-time.Minute)},
		{Role: "model", Text: "old response", CreatedAt: time.Now().Add(-30 * time.Second)},
	}

	repo := &mockChatRepo{
		SaveFn: func(_ context.Context, _, _ uuid.UUID, role, text string) error {
			dbMessages = append(dbMessages, StoredMessage{Role: role, Text: text, CreatedAt: time.Now()})
			return nil
		},
		ListFn: func(_ context.Context, _, _ uuid.UUID) ([]StoredMessage, error) {
			return dbMessages, nil
		},
	}

	svc := NewService("test-api-key", repo, defaultLessonCtx)
	h := NewHandler(svc)
	svc.APIBaseURL = geminiMock.URL
	svc.HTTPClient = &http.Client{Timeout: 5 * time.Second}

	w, c := chatRequest(http.MethodPost, "/chat", Request{
		LessonID: lessonID.String(),
		Message:  "Привет!",
	})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["text"] != "Привет! Чем помочь?" {
		t.Fatalf("text = %v; want 'Привет! Чем помочь?'", resp["text"])
	}
}

// TestChat_HistoryFromDB verifies the server uses DB history, NOT client-supplied messages.
func TestChat_HistoryFromDB(t *testing.T) {
	lessonID := uuid.New()

	var sentContents []interface{}
	geminiMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		sentContents = body["contents"].([]interface{})
		json.NewEncoder(w).Encode(geminiResponse{
			Candidates: []geminiCandidate{{
				Content: geminiContent{Role: "model", Parts: []geminiPart{{Text: "ok"}}},
			}},
		})
	}))
	defer geminiMock.Close()

	dbMessages := []StoredMessage{
		{Role: "user", Text: "first question"},
		{Role: "model", Text: "first answer"},
	}

	repo := &mockChatRepo{
		SaveFn: func(_ context.Context, _, _ uuid.UUID, role, text string) error {
			dbMessages = append(dbMessages, StoredMessage{Role: role, Text: text})
			return nil
		},
		ListFn: func(_ context.Context, _, _ uuid.UUID) ([]StoredMessage, error) {
			return dbMessages, nil
		},
	}

	svc := NewService("key", repo, defaultLessonCtx)
	h := NewHandler(svc)
	svc.APIBaseURL = geminiMock.URL
	svc.HTTPClient = &http.Client{Timeout: 5 * time.Second}

	w, c := chatRequest(http.MethodPost, "/chat", Request{
		LessonID: lessonID.String(),
		Message:  "second question",
	})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	// History: first question, first answer, second question = 3 messages
	if len(sentContents) != 3 {
		t.Fatalf("Gemini got %d messages; want 3 (from DB history)", len(sentContents))
	}
}

func TestChat_NoLessonFallback(t *testing.T) {
	// Without lesson_id, should use only the new message.
	geminiMock := httptest.NewServer(geminiOK("ответ"))
	defer geminiMock.Close()

	svc := NewService("key", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	svc.APIBaseURL = geminiMock.URL
	svc.HTTPClient = &http.Client{Timeout: 5 * time.Second}

	w, c := chatRequest(http.MethodPost, "/chat", Request{
		Message: "hello without lesson",
	})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestChat_GeminiError(t *testing.T) {
	geminiMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal"}`))
	}))
	defer geminiMock.Close()

	svc := NewService("key", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	svc.APIBaseURL = geminiMock.URL
	svc.HTTPClient = &http.Client{Timeout: 5 * time.Second}

	w, c := chatRequest(http.MethodPost, "/chat", Request{Message: "hello"})
	c.Set("user_id", uuid.New())

	h.Chat(c)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadGateway)
	}
}

// ==================== GetHistory ====================

func TestGetHistory_Success(t *testing.T) {
	lessonID := uuid.New()
	repo := &mockChatRepo{
		ListFn: func(_ context.Context, _, _ uuid.UUID) ([]StoredMessage, error) {
			return []StoredMessage{
				{Role: "user", Text: "hi", CreatedAt: time.Now()},
				{Role: "model", Text: "hello!", CreatedAt: time.Now()},
			}, nil
		},
	}
	svc := NewService("key", repo, defaultLessonCtx)
	h := NewHandler(svc)
	w, c := chatRequest(http.MethodGet, "/chat/history/"+lessonID.String(), nil)
	c.Set("user_id", uuid.New())
	c.Params = gin.Params{{Key: "lessonId", Value: lessonID.String()}}

	h.GetHistory(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	msgs := resp["messages"].([]interface{})
	if len(msgs) != 2 {
		t.Fatalf("messages count = %d; want 2", len(msgs))
	}
}

func TestGetHistory_Unauthorized(t *testing.T) {
	svc := NewService("key", &mockChatRepo{}, defaultLessonCtx)
	h := NewHandler(svc)
	w, c := chatRequest(http.MethodGet, "/chat/history/"+uuid.New().String(), nil)
	c.Params = gin.Params{{Key: "lessonId", Value: uuid.New().String()}}

	h.GetHistory(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

// ==================== ClearHistory ====================

func TestClearHistory_Success(t *testing.T) {
	lessonID := uuid.New()
	deleted := false
	repo := &mockChatRepo{
		DeleteFn: func(_ context.Context, _, _ uuid.UUID) (int64, error) {
			deleted = true
			return 5, nil
		},
	}
	svc := NewService("key", repo, defaultLessonCtx)
	h := NewHandler(svc)

	r := gin.New()
	r.DELETE("/chat/history/:lessonId", func(c *gin.Context) { c.Set("user_id", uuid.New()); c.Next() }, h.ClearHistory)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/chat/history/"+lessonID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusNoContent)
	}
	if !deleted {
		t.Fatal("expected DeleteByUserAndLesson to be called")
	}
}
