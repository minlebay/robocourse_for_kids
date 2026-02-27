package chat

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// Chat input/output limits.
const (
	maxMessageText     = 1000    // max characters per single user message
	maxHistoryMessages = 50      // max messages loaded from DB for Gemini context
	maxResponseSize    = 2 << 20 // 2 MB — limit on Gemini response body
)

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
type Request struct {
	LessonID string `json:"lesson_id"`
	Message  string `json:"message"`
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
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// handleErr maps service HTTPError to an HTTP response. Returns true if the error was handled.
func handleErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var he *HTTPError
	if errors.As(err, &he) {
		c.JSON(he.Code, gin.H{"error": he.Message})
		return true
	}
	httplog.LogError(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	return true
}

func (h *Handler) Chat(c *gin.Context) {
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
	lessonID := uuid.Nil
	if req.LessonID != "" {
		if parsed, err := uuid.Parse(req.LessonID); err == nil {
			lessonID = parsed
		}
	}
	responseText, err := h.svc.Chat(c.Request.Context(), userID, lessonID, req.Message)
	if handleErr(c, err) {
		return
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
	messages, err := h.svc.GetHistory(c.Request.Context(), userID, lessonID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
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
	if _, err := h.svc.ClearHistory(c.Request.Context(), userID, lessonID); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}
