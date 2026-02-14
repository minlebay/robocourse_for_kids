package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

var geminiClient = &http.Client{Timeout: 60 * time.Second}

// ChatMessage — одно сообщение в диалоге (user или model).
type ChatMessage struct {
	Role string `json:"role"` // "user" | "model"
	Text string `json:"text"`
}

// Request — тело запроса к нашему API.
type Request struct {
	LessonID      string        `json:"lesson_id"`       // id урока для сохранения истории
	LessonContext string        `json:"lesson_context"` // промпт с ролью и текстом урока
	Messages      []ChatMessage `json:"messages"`
}

// Response — ответ Gemini (извлекаем текст из первого кандидата).
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
	apiKey string
	repo   *Repo
}

func NewHandler(apiKey string, repo *Repo) *Handler {
	return &Handler{apiKey: apiKey, repo: repo}
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
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужно хотя бы одно сообщение"})
		return
	}

	lessonID := uuid.Nil
	if req.LessonID != "" {
		if parsed, err := uuid.Parse(req.LessonID); err == nil {
			lessonID = parsed
		}
	}

	// сохраняем последнее сообщение пользователя для истории
	if lessonID != uuid.Nil && h.repo != nil {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == "user" {
			_ = h.repo.Save(c.Request.Context(), userID, lessonID, "user", lastMsg.Text)
		}
	}

	// system_instruction: роль + контекст урока
	systemParts := []map[string]interface{}{{"text": req.LessonContext}}
	// contents: история в формате Gemini (role + parts)
	contents := make([]map[string]interface{}, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := "user"
		if m.Role == "model" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{{"text": m.Text}},
		})
	}

	body := map[string]interface{}{
		"system_instruction": map[string]interface{}{"parts": systemParts},
		"contents":           contents,
	}
	raw, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, geminiAPIURL+"?key="+h.apiKey, bytes.NewReader(raw))
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := geminiClient.Do(httpReq)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI service unavailable"})
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		rid, _ := c.Get("request_id")
		log.Printf("[%v] Gemini status %d: %s", rid, resp.StatusCode, string(b))
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
		// сохраняем ответ модели для истории
		if lessonID != uuid.Nil && h.repo != nil {
			_ = h.repo.Save(c.Request.Context(), userID, lessonID, "model", responseText)
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
