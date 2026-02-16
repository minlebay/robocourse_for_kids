package progress

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

type mockRepo struct {
	GetProgressFn                func(ctx context.Context, userID uuid.UUID) (*UserProgress, error)
	SetLessonProgressFn          func(ctx context.Context, userID, lessonID uuid.UUID, status string) error
	SetChecklistItemCompletedFn  func(ctx context.Context, userID, checklistItemID uuid.UUID, completed bool) error
	ChecklistItemBelongsToLessonFn func(ctx context.Context, lessonID, itemID uuid.UUID) (bool, error)
}

func (m *mockRepo) GetProgress(ctx context.Context, userID uuid.UUID) (*UserProgress, error) {
	if m.GetProgressFn != nil {
		return m.GetProgressFn(ctx, userID)
	}
	return &UserProgress{Lessons: nil, Checklist: nil}, nil
}

func (m *mockRepo) SetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, status string) error {
	if m.SetLessonProgressFn != nil {
		return m.SetLessonProgressFn(ctx, userID, lessonID, status)
	}
	return nil
}

func (m *mockRepo) SetChecklistItemCompleted(ctx context.Context, userID, checklistItemID uuid.UUID, completed bool) error {
	if m.SetChecklistItemCompletedFn != nil {
		return m.SetChecklistItemCompletedFn(ctx, userID, checklistItemID, completed)
	}
	return nil
}

func (m *mockRepo) ChecklistItemBelongsToLesson(ctx context.Context, lessonID, itemID uuid.UUID) (bool, error) {
	if m.ChecklistItemBelongsToLessonFn != nil {
		return m.ChecklistItemBelongsToLessonFn(ctx, lessonID, itemID)
	}
	return true, nil
}

type mockUserChecker struct {
	UserExistsFn func(ctx context.Context, id uuid.UUID) (bool, error)
}

func (m *mockUserChecker) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.UserExistsFn != nil {
		return m.UserExistsFn(ctx, id)
	}
	return true, nil
}

func jsonRequest(method, target string, body interface{}, params map[string]string, userID uuid.UUID) (*httptest.ResponseRecorder, *gin.Context) {
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
	if userID != uuid.Nil {
		c.Set("user_id", userID)
	}
	return w, c
}

func TestGetProgress_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{}, nil)
	w, c := jsonRequest(http.MethodGet, "/progress", nil, nil, uuid.Nil)
	h.GetProgress(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", w.Code)
	}
}

func TestGetProgress_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepo{
		GetProgressFn: func(ctx context.Context, uid uuid.UUID) (*UserProgress, error) {
			return &UserProgress{Lessons: []LessonProgress{}, Checklist: []ChecklistProgressItem{}}, nil
		},
	}
	h := NewHandler(repo, nil)
	w, c := jsonRequest(http.MethodGet, "/progress", nil, nil, userID)
	h.GetProgress(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestGetProgress_RepoError(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepo{GetProgressFn: func(ctx context.Context, uid uuid.UUID) (*UserProgress, error) {
		return nil, errors.New("db error")
	}}
	h := NewHandler(repo, nil)
	w, c := jsonRequest(http.MethodGet, "/progress", nil, nil, userID)
	h.GetProgress(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d; want 500", w.Code)
	}
}

func TestSetLessonProgress_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{}, nil)
	lessonID := uuid.New()
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/progress", SetProgressRequest{Status: StatusInProgress}, map[string]string{"id": lessonID.String()}, uuid.Nil)
	h.SetLessonProgress(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", w.Code)
	}
}

func TestSetLessonProgress_InvalidLessonID(t *testing.T) {
	h := NewHandler(&mockRepo{}, nil)
	w, c := jsonRequest(http.MethodPut, "/lessons/bad/progress", SetProgressRequest{Status: StatusCompleted}, map[string]string{"id": "bad"}, uuid.New())
	h.SetLessonProgress(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestSetLessonProgress_InvalidStatus(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	h := NewHandler(&mockRepo{}, nil)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/progress", SetProgressRequest{Status: "invalid"}, map[string]string{"id": lessonID.String()}, userID)
	h.SetLessonProgress(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestSetLessonProgress_Success(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	repo := &mockRepo{}
	h := NewHandler(repo, nil)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/progress", SetProgressRequest{Status: StatusCompleted}, map[string]string{"id": lessonID.String()}, userID)
	h.SetLessonProgress(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestSetChecklistItem_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{}, nil)
	lessonID := uuid.New()
	itemID := uuid.New()
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/checklist/"+itemID.String(), map[string]bool{"completed": true}, map[string]string{"id": lessonID.String(), "itemId": itemID.String()}, uuid.Nil)
	h.SetChecklistItem(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", w.Code)
	}
}

func TestSetChecklistItem_InvalidLessonID(t *testing.T) {
	h := NewHandler(&mockRepo{}, nil)
	itemID := uuid.New()
	w, c := jsonRequest(http.MethodPut, "/lessons/bad/checklist/"+itemID.String(), map[string]bool{"completed": true}, map[string]string{"id": "bad", "itemId": itemID.String()}, uuid.New())
	h.SetChecklistItem(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestSetChecklistItem_InvalidItemID(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	h := NewHandler(&mockRepo{}, nil)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/checklist/bad", map[string]bool{"completed": true}, map[string]string{"id": lessonID.String(), "itemId": "bad"}, userID)
	h.SetChecklistItem(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestSetChecklistItem_ItemNotBelongsToLesson(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	itemID := uuid.New()
	repo := &mockRepo{ChecklistItemBelongsToLessonFn: func(ctx context.Context, lid, iid uuid.UUID) (bool, error) {
		return false, nil
	}}
	h := NewHandler(repo, nil)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/checklist/"+itemID.String(), map[string]bool{"completed": true}, map[string]string{"id": lessonID.String(), "itemId": itemID.String()}, userID)
	h.SetChecklistItem(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestSetChecklistItem_Success(t *testing.T) {
	userID := uuid.New()
	lessonID := uuid.New()
	itemID := uuid.New()
	repo := &mockRepo{}
	h := NewHandler(repo, nil)
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String()+"/checklist/"+itemID.String(), map[string]bool{"completed": true}, map[string]string{"id": lessonID.String(), "itemId": itemID.String()}, userID)
	h.SetChecklistItem(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestGetUserProgress_InvalidUserID(t *testing.T) {
	h := NewHandler(&mockRepo{}, nil)
	w, c := jsonRequest(http.MethodGet, "/users/bad/progress", nil, map[string]string{"id": "bad"}, uuid.New())
	h.GetUserProgress(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestGetUserProgress_UserNotFound(t *testing.T) {
	targetID := uuid.New()
	checker := &mockUserChecker{
		UserExistsFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(&mockRepo{}, checker)
	w, c := jsonRequest(http.MethodGet, "/users/"+targetID.String()+"/progress", nil, map[string]string{"id": targetID.String()}, uuid.New())
	h.GetUserProgress(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

func TestGetUserProgress_Success(t *testing.T) {
	targetID := uuid.New()
	repo := &mockRepo{
		GetProgressFn: func(ctx context.Context, uid uuid.UUID) (*UserProgress, error) {
			return &UserProgress{Lessons: []LessonProgress{}, Checklist: nil}, nil
		},
	}
	h := NewHandler(repo, nil)
	w, c := jsonRequest(http.MethodGet, "/users/"+targetID.String()+"/progress", nil, map[string]string{"id": targetID.String()}, uuid.New())
	h.GetUserProgress(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}
