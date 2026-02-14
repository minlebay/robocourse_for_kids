package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

type mockRepo struct {
	ListByLessonFn      func(ctx context.Context, lessonID uuid.UUID) ([]Comment, error)
	CreateFn            func(ctx context.Context, lessonID, userID uuid.UUID, text string) (*Comment, error)
	DeleteByIDAndUserFn func(ctx context.Context, commentID, lessonID, userID uuid.UUID) (bool, error)
}

func (m *mockRepo) ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]Comment, error) {
	if m.ListByLessonFn != nil {
		return m.ListByLessonFn(ctx, lessonID)
	}
	return nil, nil
}

func (m *mockRepo) Create(ctx context.Context, lessonID, userID uuid.UUID, text string) (*Comment, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, lessonID, userID, text)
	}
	return &Comment{ID: uuid.New(), LessonID: lessonID, UserID: userID, Text: text, CreatedAt: time.Now()}, nil
}

func (m *mockRepo) DeleteByIDAndUser(ctx context.Context, commentID, lessonID, userID uuid.UUID) (bool, error) {
	if m.DeleteByIDAndUserFn != nil {
		return m.DeleteByIDAndUserFn(ctx, commentID, lessonID, userID)
	}
	return true, nil
}

func jsonRequest(method, target string, body interface{}, params map[string]string) (*httptest.ResponseRecorder, *gin.Context) {
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
	if params != nil {
		for k, v := range params {
			c.Params = append(c.Params, gin.Param{Key: k, Value: v})
		}
	}
	return w, c
}

func TestList_InvalidLessonID(t *testing.T) {
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodGet, "/lessons/bad/comments", nil, map[string]string{"id": "bad"})
	h.List(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestList_Success(t *testing.T) {
	lessonID := uuid.New()
	repo := &mockRepo{
		ListByLessonFn: func(ctx context.Context, lid uuid.UUID) ([]Comment, error) {
			return []Comment{{ID: uuid.New(), LessonID: lid, Text: "hello"}}, nil
		},
	}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodGet, "/lessons/"+lessonID.String()+"/comments", nil, map[string]string{"id": lessonID.String()})
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestList_RepoError(t *testing.T) {
	lessonID := uuid.New()
	repo := &mockRepo{ListByLessonFn: func(ctx context.Context, lid uuid.UUID) ([]Comment, error) {
		return nil, errors.New("db error")
	}}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodGet, "/lessons/"+lessonID.String()+"/comments", nil, map[string]string{"id": lessonID.String()})
	h.List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d; want 500", w.Code)
	}
}

func TestCreate_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{})
	lessonID := uuid.New()
	w, c := jsonRequest(http.MethodPost, "/lessons/"+lessonID.String()+"/comments", CreateCommentRequest{Text: "hi"}, map[string]string{"id": lessonID.String()})
	// no user_id in context
	h.Create(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", w.Code)
	}
}

func TestCreate_InvalidLessonID(t *testing.T) {
	userID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodPost, "/lessons/not-uuid/comments", CreateCommentRequest{Text: "hi"}, map[string]string{"id": "not-uuid"})
	c.Set("user_id", userID)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestCreate_EmptyTextAfterSanitize(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	h := NewHandler(&mockRepo{})
	// only HTML tags -> sanitize strips all -> empty
	w, c := jsonRequest(http.MethodPost, "/lessons/"+lessonID.String()+"/comments", CreateCommentRequest{Text: "<script></script>"}, map[string]string{"id": lessonID.String()})
	c.Set("user_id", userID)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body = %s", w.Code, w.Body.String())
	}
}

func TestCreate_TextTooLong(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	h := NewHandler(&mockRepo{})
	long := ""
	for i := 0; i < 2001; i++ {
		long += "a"
	}
	w, c := jsonRequest(http.MethodPost, "/lessons/"+lessonID.String()+"/comments", CreateCommentRequest{Text: long}, map[string]string{"id": lessonID.String()})
	c.Set("user_id", userID)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestCreate_Success(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	var capturedText string
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, lid, uid uuid.UUID, text string) (*Comment, error) {
			capturedText = text
			return &Comment{ID: uuid.New(), LessonID: lid, UserID: uid, Text: text, CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodPost, "/lessons/"+lessonID.String()+"/comments", CreateCommentRequest{Text: "Hello world"}, map[string]string{"id": lessonID.String()})
	c.Set("user_id", userID)
	h.Create(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201; body = %s", w.Code, w.Body.String())
	}
	if capturedText != "Hello world" {
		t.Errorf("captured text = %q; want Hello world", capturedText)
	}
}

func TestDelete_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{})
	lessonID := uuid.New()
	commentID := uuid.New()
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+lessonID.String()+"/comments/"+commentID.String(), nil, map[string]string{"id": lessonID.String(), "commentId": commentID.String()})
	h.Delete(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", w.Code)
	}
}

func TestDelete_InvalidCommentID(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	h := NewHandler(&mockRepo{})
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+lessonID.String()+"/comments/bad", nil, map[string]string{"id": lessonID.String(), "commentId": "bad"})
	c.Set("user_id", userID)
	h.Delete(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	commentID := uuid.New()
	repo := &mockRepo{DeleteByIDAndUserFn: func(ctx context.Context, cid, lid, uid uuid.UUID) (bool, error) {
		return false, nil
	}}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+lessonID.String()+"/comments/"+commentID.String(), nil, map[string]string{"id": lessonID.String(), "commentId": commentID.String()})
	c.Set("user_id", userID)
	h.Delete(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

func TestDelete_Success(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	commentID := uuid.New()
	repo := &mockRepo{}
	h := NewHandler(repo)
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+lessonID.String()+"/comments/"+commentID.String(), nil, map[string]string{"id": lessonID.String(), "commentId": commentID.String()})
	c.Set("user_id", userID)
	h.Delete(c)
	// Handler returns 204 No Content; in some test contexts Gin may record 200
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 204 or 200", w.Code)
	}
}
