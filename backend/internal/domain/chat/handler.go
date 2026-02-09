package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// ChatMessage — одно сообщение в диалоге (user или model).
type ChatMessage struct {
	Role string `json:"role"` // "user" | "model"
	Text string `json:"text"`
}

// Request — тело запроса к нашему API.
type Request struct {
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
}

func NewHandler(apiKey string) *Handler {
	return &Handler{apiKey: apiKey}
}

func (h *Handler) Chat(c *gin.Context) {
	if h.apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "чат с ИИ не настроен: отсутствует GEMINI_API_KEY"})
		return
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверное тело запроса: " + err.Error()})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нужно хотя бы одно сообщение"})
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "создание запроса к Gemini: " + err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "запрос к Gemini: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "чтение ответа Gemini"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("Gemini вернул %d: %s", resp.StatusCode, string(b))})
		return
	}

	var gemini geminiResponse
	if err := json.Unmarshal(b, &gemini); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "разбор ответа Gemini"})
		return
	}
	if len(gemini.Candidates) == 0 || len(gemini.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusOK, gin.H{"text": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"text": gemini.Candidates[0].Content.Parts[0].Text})
}
