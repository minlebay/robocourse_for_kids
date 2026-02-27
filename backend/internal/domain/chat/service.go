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

	"github.com/google/uuid"
	"learn_kids/backend/internal/sanitize"
)

// HTTPError carries an HTTP status code and client-facing message from the service layer.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

func httpErr(code int, msg string) *HTTPError {
	return &HTTPError{Code: code, Message: msg}
}

// Service holds the business logic and Gemini orchestration for the chat domain.
type Service struct {
	apiKey        string
	repo          Repository
	lessonContext LessonContextFunc
	// Overridable for testing.
	APIBaseURL string
	HTTPClient *http.Client
}

func NewService(apiKey string, repo Repository, lcf LessonContextFunc) *Service {
	return &Service{
		apiKey:        apiKey,
		repo:          repo,
		lessonContext: lcf,
		APIBaseURL:    geminiAPIURL,
		HTTPClient:    defaultHTTPClient,
	}
}

// Chat validates the message, orchestrates a Gemini call with DB history, saves the reply,
// and returns the model's response text.
func (s *Service) Chat(ctx context.Context, userID, lessonID uuid.UUID, rawMessage string) (string, error) {
	if s.apiKey == "" {
		return "", httpErr(http.StatusServiceUnavailable, "чат с ИИ не настроен: отсутствует GEMINI_API_KEY")
	}
	if len(rawMessage) == 0 {
		return "", httpErr(http.StatusBadRequest, "сообщение не может быть пустым")
	}
	if len(rawMessage) > maxMessageText {
		return "", httpErr(http.StatusBadRequest, fmt.Sprintf("сообщение слишком длинное (максимум %d символов)", maxMessageText))
	}
	message := sanitize.ChatMessage(rawMessage)
	if message == "" {
		return "", httpErr(http.StatusBadRequest, "сообщение не может быть пустым")
	}

	// 1. Save the new user message to DB first.
	if lessonID != uuid.Nil && s.repo != nil {
		if err := s.repo.Save(ctx, userID, lessonID, "user", message); err != nil {
			slog.Warn("save user chat message", "err", err)
		}
	}

	// 2. Load conversation history from DB (server-authoritative — prevents prompt injection).
	var chatMessages []ChatMessage
	if lessonID != uuid.Nil && s.repo != nil {
		stored, err := s.repo.ListByUserAndLesson(ctx, userID, lessonID)
		if err != nil {
			slog.Warn("load chat history", "err", err)
		} else {
			if len(stored) > maxHistoryMessages {
				stored = stored[len(stored)-maxHistoryMessages:]
			}
			for _, m := range stored {
				chatMessages = append(chatMessages, ChatMessage{Role: m.Role, Text: m.Text})
			}
		}
	}
	if len(chatMessages) == 0 {
		chatMessages = []ChatMessage{{Role: "user", Text: message}}
	}

	// 3. Build Gemini request.
	systemPrompt := s.lessonContext(ctx, lessonID)
	systemParts := []map[string]interface{}{{"text": systemPrompt}}
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.APIBaseURL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", s.apiKey)

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return "", httpErr(http.StatusBadGateway, "AI service unavailable")
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return "", err
	}
	if len(b) > maxResponseSize {
		slog.Warn("gemini response too large", "size", len(b))
		return "", httpErr(http.StatusBadGateway, "AI response too large")
	}
	if resp.StatusCode != http.StatusOK {
		bodyPreview := string(b)
		if len(bodyPreview) > 500 {
			bodyPreview = bodyPreview[:500]
		}
		slog.Error("gemini request failed", "status", resp.StatusCode, "body", bodyPreview)
		return "", httpErr(http.StatusBadGateway, "AI service error")
	}

	var gemini geminiResponse
	if err := json.Unmarshal(b, &gemini); err != nil {
		return "", err
	}
	responseText := ""
	if len(gemini.Candidates) > 0 && len(gemini.Candidates[0].Content.Parts) > 0 {
		responseText = gemini.Candidates[0].Content.Parts[0].Text
		if lessonID != uuid.Nil && s.repo != nil {
			if err := s.repo.Save(ctx, userID, lessonID, "model", responseText); err != nil {
				slog.Warn("save model chat message", "err", err)
			}
		}
	}
	return responseText, nil
}

// GetHistory returns stored chat messages for a user and lesson.
func (s *Service) GetHistory(ctx context.Context, userID, lessonID uuid.UUID) ([]ChatMessage, error) {
	list, err := s.repo.ListByUserAndLesson(ctx, userID, lessonID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []ChatMessage{}, nil
	}
	messages := make([]ChatMessage, len(list))
	for i, m := range list {
		messages[i] = ChatMessage{Role: m.Role, Text: m.Text}
	}
	return messages, nil
}

// ClearHistory deletes all chat messages for a user and lesson.
func (s *Service) ClearHistory(ctx context.Context, userID, lessonID uuid.UUID) (int64, error) {
	return s.repo.DeleteByUserAndLesson(ctx, userID, lessonID)
}

// defaultHTTPClient is used unless overridden via Service.HTTPClient.
var defaultHTTPClient = &http.Client{Timeout: 60 * time.Second}
