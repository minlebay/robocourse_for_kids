package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
	"learn_kids/backend/internal/sanitize"
)

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// Chat input/output limits.
const (
	maxMessageText     = 1000 // max characters per single user message
	maxHistoryMessages = 50   // max messages loaded from DB for Gemini context
	maxResponseSize    = 2 << 20 // 2 MB — limit on Gemini response body
)

var defaultHTTPClient = &http.Client{Timeout: 60 * time.Second}

// LessonContextFunc builds a system prompt for the AI based on the lesson ID.
// Returns a safe server-generated prompt to prevent prompt injection.
type LessonContextFunc func(ctx context.Context, lessonID uuid.UUID) string

// Repository defines the data access interface for chat history.
type Repository interface {
	ListByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) ([]StoredMessage, error)
	Save(ctx context.Context, userID, lessonID uuid.UUID, role, text string) error
	DeleteByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) (int64, error)
}

// ChatMessage — одно сообщение в диалоге (user или model).
type ChatMessage struct {
	Role string `json:"role"` // "user" | "model"
	Text string `json:"text"`
}

// Request — тело запроса к нашему API.
// Клиент отправляет только последнее сообщение; историю сервер загружает из БД.
type Request struct {
	LessonID string `json:"lesson_id"` // id урока для контекста и истории
	Message  string `json:"message"`   // новое сообщение пользователя
}

// Gemini API response types.
type geminiPart struct {
	Text string `json:"text"`
}
type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}
type geminiCandidate struct {
	Content geminiContent `json:"content"`
}
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type Handler struct {
	apiKey        string
	repo          Repository
	lessonContext LessonContextFunc
	// Overridable for testing.
	APIBaseURL string       // defaults to geminiAPIURL
	HTTPClient *http.Client // defaults to defaultHTTPClient
}

func NewHandler(apiKey string, repo Repository, lcf LessonContextFunc) *Handler {
	return &Handler{
		apiKey:        apiKey,
		repo:          repo,
		lessonContext: lcf,
		APIBaseURL:    geminiAPIURL,
		HTTPClient:    defaultHTTPClient,
	}
}

func (h *Handler) Chat(c *gin.Context) {
	if h.apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "чат с ИИ не настроен: отсутствует GEMINI_API_KEY"})
		return
	}

	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(req.Message) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "сообщение не может быть пустым"})
		return
	}
	if len(req.Message) > maxMessageText {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("сообщение слишком длинное (максимум %d символов)", maxMessageText)})
		return
	}
	req.Message = sanitize.ChatMessage(req.Message)
	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "сообщение не может быть пустым"})
		return
	}

	ctx := c.Request.Context()

	lessonID := uuid.Nil
	if req.LessonID != "" {
		if parsed, err := uuid.Parse(req.LessonID); err == nil {
			lessonID = parsed
		}
	}

	// 1. Save the new user message to DB first.
	if lessonID != uuid.Nil && h.repo != nil {
		if err := h.repo.Save(ctx, userID, lessonID, "user", req.Message); err != nil {
			httplog.LogWarn(c, fmt.Sprintf("save user chat message: %v", err))
		}
	}

	// 2. Load conversation history from DB (server-authoritative — prevents prompt injection).
	//    The history includes the just-saved user message.
	var chatMessages []ChatMessage
	if lessonID != uuid.Nil && h.repo != nil {
		stored, err := h.repo.ListByUserAndLesson(ctx, userID, lessonID)
		if err != nil {
			httplog.LogWarn(c, fmt.Sprintf("load chat history: %v", err))
		} else {
			// Limit to last N messages to stay within reasonable context size.
			if len(stored) > maxHistoryMessages {
				stored = stored[len(stored)-maxHistoryMessages:]
			}
			for _, m := range stored {
				chatMessages = append(chatMessages, ChatMessage{Role: m.Role, Text: m.Text})
			}
		}
	}

	// Fallback: if no history loaded (no lesson_id or DB error), use just the new message.
	if len(chatMessages) == 0 {
		chatMessages = []ChatMessage{{Role: "user", Text: req.Message}}
	}

	// 3. Build Gemini request.
	// system_instruction: formed on the server by lesson_id (prompt injection protection).
	systemPrompt := h.lessonContext(ctx, lessonID)
	systemParts := []map[string]interface{}{{"text": systemPrompt}}

	// contents: conversation history in Gemini format (role + parts).
	contents := make([]map[string]interface{}, 0, len(chatMessages))
	for _, m := range chatMessages {
		role := "user"
		if m.Role == "model" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": []map[string]interface{}{{"text": m.Text}},
		})
	}

	body := map[string]interface{}{
		"system_instruction": map[string]interface{}{"parts": systemParts},
		"contents":           contents,
	}
	raw, _ := json.Marshal(body)

	apiURL := h.APIBaseURL
	client := h.HTTPClient

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(raw))
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", h.apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI service unavailable"})
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if len(b) > maxResponseSize {
		rid, _ := c.Get("request_id")
		slog.Warn("gemini response too large", "request_id", rid, "size", len(b))
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI response too large"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		rid, _ := c.Get("request_id")
		bodyPreview := string(b)
		if len(bodyPreview) > 500 {
			bodyPreview = bodyPreview[:500]
		}
		slog.Error("gemini request failed", "request_id", rid, "status", resp.StatusCode, "body", bodyPreview)
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI service error"})
		return
	}

	var gemini geminiResponse
	if err := json.Unmarshal(b, &gemini); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	responseText := ""
	if len(gemini.Candidates) > 0 && len(gemini.Candidates[0].Content.Parts) > 0 {
		responseText = gemini.Candidates[0].Content.Parts[0].Text
		// Save model response to history.
		if lessonID != uuid.Nil && h.repo != nil {
			if err := h.repo.Save(ctx, userID, lessonID, "model", responseText); err != nil {
				httplog.LogWarn(c, fmt.Sprintf("save model chat message: %v", err))
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"text": responseText})
}

// GetHistory возвращает историю чата пользователя по уроку.
func (h *Handler) GetHistory(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	list, err := h.repo.ListByUserAndLesson(c.Request.Context(), userID, lessonID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if list == nil {
		list = []StoredMessage{}
	}
	messages := make([]ChatMessage, len(list))
	for i, m := range list {
		messages[i] = ChatMessage{Role: m.Role, Text: m.Text}
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// ClearHistory очищает историю чата пользователя по уроку.
func (h *Handler) ClearHistory(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	_, err = h.repo.DeleteByUserAndLesson(c.Request.Context(), userID, lessonID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
