package lessons

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"learn_kids/backend/internal/requestcontext"
	"learn_kids/backend/internal/domain/users"
)

func init() { gin.SetMode(gin.TestMode) }

// mockRepo captures CreateLesson and UpdateLesson arguments for assertions.
type mockRepo struct {
	ListModulesFn    func(ctx context.Context, tag *string, ownerID *uuid.UUID) ([]Module, error)
	GetModuleByIDFn  func(ctx context.Context, id uuid.UUID) (*Module, error)
	CreateModuleFn   func(ctx context.Context, title, description string, sortOrder int, ownerID *uuid.UUID) (*Module, error)
	DeleteModuleFn   func(ctx context.Context, id uuid.UUID) (bool, error)
	GetLessonByIDFn  func(ctx context.Context, id uuid.UUID) (*Lesson, error)
	CreateLessonFn   func(ctx context.Context, moduleID uuid.UUID, title, description, lessonType string, sortOrder int, steps []LessonStep) (*Lesson, error)
	DeleteLessonFn   func(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateLessonFn   func(ctx context.Context, id uuid.UUID, title, description *string, steps []LessonStep) (*Lesson, error)
}

func (m *mockRepo) ListModules(ctx context.Context, tag *string, ownerID *uuid.UUID) ([]Module, error) {
	if m.ListModulesFn != nil {
		return m.ListModulesFn(ctx, tag, ownerID)
	}
	return nil, nil
}
func (m *mockRepo) GetModuleByID(ctx context.Context, id uuid.UUID) (*Module, error) {
	if m.GetModuleByIDFn != nil {
		return m.GetModuleByIDFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}
func (m *mockRepo) CreateModule(ctx context.Context, title, description string, sortOrder int, ownerID *uuid.UUID) (*Module, error) {
	if m.CreateModuleFn != nil {
		return m.CreateModuleFn(ctx, title, description, sortOrder, ownerID)
	}
	return &Module{ID: uuid.New(), Title: title, Description: description, SortOrder: sortOrder}, nil
}
func (m *mockRepo) DeleteModule(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.DeleteModuleFn != nil {
		return m.DeleteModuleFn(ctx, id)
	}
	return true, nil
}
func (m *mockRepo) GetLessonByID(ctx context.Context, id uuid.UUID) (*Lesson, error) {
	if m.GetLessonByIDFn != nil {
		return m.GetLessonByIDFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}
func (m *mockRepo) CreateLesson(ctx context.Context, moduleID uuid.UUID, title, description, lessonType string, sortOrder int, steps []LessonStep) (*Lesson, error) {
	if m.CreateLessonFn != nil {
		return m.CreateLessonFn(ctx, moduleID, title, description, lessonType, sortOrder, steps)
	}
	return &Lesson{ID: uuid.New(), ModuleID: moduleID, Title: title, Description: description, LessonType: lessonType, SortOrder: sortOrder, Steps: steps}, nil
}
func (m *mockRepo) DeleteLesson(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.DeleteLessonFn != nil {
		return m.DeleteLessonFn(ctx, id)
	}
	return true, nil
}
func (m *mockRepo) UpdateLesson(ctx context.Context, id uuid.UUID, title, description *string, steps []LessonStep) (*Lesson, error) {
	if m.UpdateLessonFn != nil {
		return m.UpdateLessonFn(ctx, id, title, description, steps)
	}
	return &Lesson{ID: id}, nil
}

func jsonRequest(method, target string, body interface{}, paramID string) (*httptest.ResponseRecorder, *gin.Context) {
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
	if paramID != "" {
		c.Params = gin.Params{{Key: "id", Value: paramID}}
	}
	return w, c
}

// setEditorContext sets user_id and user_roles so that canEditModule passes (e.g. administrator).
func setEditorContext(c *gin.Context) {
	uid := uuid.New()
	c.Set(requestcontext.UserIDKey, uid)
	c.Set("user_roles", []string{users.RoleAdministrator})
}

// --- CreateLesson: sanitization ---

func TestCreateLesson_SanitizesStepContent(t *testing.T) {
	moduleID := uuid.New()
	var capturedSteps []LessonStep
	repo := &mockRepo{
		CreateLessonFn: func(ctx context.Context, modID uuid.UUID, title, description, lessonType string, sortOrder int, steps []LessonStep) (*Lesson, error) {
			capturedSteps = steps
			return &Lesson{ID: uuid.New(), ModuleID: modID, Title: title, Steps: steps}, nil
		},
	}
	h := NewHandler(NewService(repo, nil))

	body := CreateLessonRequest{
		Title:      "Lesson",
		LessonType: "theory",
		Steps: []LessonStep{
			{Title: "Step 1", Content: `<script>alert("xss")</script> and <img src=x onerror="evil()">`},
		},
	}
	w, c := jsonRequest(http.MethodPost, "/modules/"+moduleID.String()+"/lessons", body, moduleID.String())
	c.Params = gin.Params{{Key: "id", Value: moduleID.String()}}
	setEditorContext(c)
	repo.GetModuleByIDFn = func(ctx context.Context, id uuid.UUID) (*Module, error) {
		return &Module{ID: id, OwnerID: nil}, nil
	}

	h.CreateLesson(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201; body = %s", w.Code, w.Body.String())
	}
	if len(capturedSteps) != 1 {
		t.Fatalf("captured steps len = %d; want 1", len(capturedSteps))
	}
	content := capturedSteps[0].Content
	if content != " and " {
		t.Errorf("step content (should be sanitized) = %q; should not contain script or img tags", content)
	}
}

func TestCreateLesson_RejectsTooLongTitle(t *testing.T) {
	moduleID := uuid.New()
	repo := &mockRepo{}
	h := NewHandler(NewService(repo, nil))

	longTitle := ""
	for i := 0; i <= MaxTitleLength; i++ {
		longTitle += "a"
	}

	body := CreateLessonRequest{Title: longTitle, LessonType: "theory"}
	w, c := jsonRequest(http.MethodPost, "/modules/"+moduleID.String()+"/lessons", body, moduleID.String())
	c.Params = gin.Params{{Key: "id", Value: moduleID.String()}}
	setEditorContext(c)
	repo.GetModuleByIDFn = func(ctx context.Context, id uuid.UUID) (*Module, error) {
		return &Module{ID: id}, nil
	}

	h.CreateLesson(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body = %s", w.Code, w.Body.String())
	}
}

func TestCreateLesson_RejectsTooLongStepContent(t *testing.T) {
	moduleID := uuid.New()
	repo := &mockRepo{}
	h := NewHandler(NewService(repo, nil))

	longContent := ""
	for i := 0; i <= MaxStepContentLength; i++ {
		longContent += "x"
	}

	body := CreateLessonRequest{
		Title:      "Lesson",
		LessonType: "theory",
		Steps:      []LessonStep{{Title: "Step", Content: longContent}},
	}
	w, c := jsonRequest(http.MethodPost, "/modules/"+moduleID.String()+"/lessons", body, moduleID.String())
	c.Params = gin.Params{{Key: "id", Value: moduleID.String()}}
	setEditorContext(c)
	repo.GetModuleByIDFn = func(ctx context.Context, id uuid.UUID) (*Module, error) {
		return &Module{ID: id}, nil
	}

	h.CreateLesson(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body = %s", w.Code, w.Body.String())
	}
}

// --- UpdateLesson: sanitization and length ---

func TestUpdateLesson_SanitizesStepContent(t *testing.T) {
	lessonID := uuid.New()
	var capturedSteps []LessonStep
	repo := &mockRepo{
		GetLessonByIDFn: func(ctx context.Context, id uuid.UUID) (*Lesson, error) {
			return &Lesson{ID: id, ModuleID: uuid.New()}, nil
		},
		GetModuleByIDFn: func(ctx context.Context, id uuid.UUID) (*Module, error) {
			return &Module{ID: id}, nil
		},
		UpdateLessonFn: func(ctx context.Context, id uuid.UUID, title, description *string, steps []LessonStep) (*Lesson, error) {
			capturedSteps = steps
			return &Lesson{ID: id}, nil
		},
	}
	h := NewHandler(NewService(repo, nil))

	steps := []LessonStep{
		{Title: "Step", Content: `<a href="javascript:alert(1)">click</a>`},
	}
	body := UpdateLessonRequest{Steps: &steps}
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String(), body, lessonID.String())
	c.Params = gin.Params{{Key: "id", Value: lessonID.String()}}
	setEditorContext(c)

	h.UpdateLesson(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200; body = %s", w.Code, w.Body.String())
	}
	if len(capturedSteps) != 1 {
		t.Fatalf("captured steps len = %d; want 1", len(capturedSteps))
	}
	// bluemonday strips the tag but keeps link text
	if capturedSteps[0].Content != "click" {
		t.Errorf("step content = %q; want sanitized 'click' (no javascript:)", capturedSteps[0].Content)
	}
}

func TestCreateModule_RejectsTooLongTitle(t *testing.T) {
	repo := &mockRepo{}
	h := NewHandler(NewService(repo, nil))
	longTitle := ""
	for i := 0; i <= MaxTitleLength; i++ {
		longTitle += "a"
	}
	body := CreateModuleRequest{Title: longTitle}
	w, c := jsonRequest(http.MethodPost, "/modules", body, "")
	setEditorContext(c)
	h.CreateModule(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body = %s", w.Code, w.Body.String())
	}
}

func TestCreateModule_SanitizesDescription(t *testing.T) {
	var capturedDesc string
	repo := &mockRepo{
		CreateModuleFn: func(ctx context.Context, title, description string, sortOrder int, ownerID *uuid.UUID) (*Module, error) {
			capturedDesc = description
			return &Module{ID: uuid.New(), Title: title, Description: description}, nil
		},
	}
	h := NewHandler(NewService(repo, nil))
	body := CreateModuleRequest{Title: "Mod", Description: `<script>bad</script> ok`}
	w, c := jsonRequest(http.MethodPost, "/modules", body, "")
	setEditorContext(c)
	h.CreateModule(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201", w.Code)
	}
	if capturedDesc != " ok" {
		t.Errorf("description = %q; want sanitized ' ok'", capturedDesc)
	}
}

func TestUpdateLesson_RejectsTooLongStepContent(t *testing.T) {
	lessonID := uuid.New()
	moduleID := uuid.New()
	longContent := ""
	for i := 0; i <= MaxStepContentLength; i++ {
		longContent += "x"
	}
	repo := &mockRepo{
		GetLessonByIDFn: func(ctx context.Context, id uuid.UUID) (*Lesson, error) {
			return &Lesson{ID: id, ModuleID: moduleID}, nil
		},
		GetModuleByIDFn: func(ctx context.Context, id uuid.UUID) (*Module, error) {
			return &Module{ID: id}, nil
		},
	}
	h := NewHandler(NewService(repo, nil))
	steps := []LessonStep{{Title: "Step", Content: longContent}}
	body := UpdateLessonRequest{Steps: &steps}
	w, c := jsonRequest(http.MethodPut, "/lessons/"+lessonID.String(), body, lessonID.String())
	c.Params = gin.Params{{Key: "id", Value: lessonID.String()}}
	setEditorContext(c)

	h.UpdateLesson(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body = %s", w.Code, w.Body.String())
	}
}

func TestListModules_Success(t *testing.T) {
	repo := &mockRepo{
		ListModulesFn: func(ctx context.Context, tag *string, ownerID *uuid.UUID) ([]Module, error) {
			return []Module{{ID: uuid.New(), Title: "Mod1", SortOrder: 1}}, nil
		},
	}
	h := NewHandler(NewService(repo, nil))
	w, c := jsonRequest(http.MethodGet, "/modules", nil, "")
	h.ListModules(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
}

func TestGetModule_InvalidID(t *testing.T) {
	h := NewHandler(NewService(&mockRepo{}, nil))
	w, c := jsonRequest(http.MethodGet, "/modules/not-a-uuid", nil, "not-a-uuid")
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	h.GetModule(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestGetModule_NotFound(t *testing.T) {
	repo := &mockRepo{GetModuleByIDFn: func(ctx context.Context, id uuid.UUID) (*Module, error) {
		return nil, pgx.ErrNoRows
	}}
	h := NewHandler(NewService(repo, nil))
	id := uuid.New()
	w, c := jsonRequest(http.MethodGet, "/modules/"+id.String(), nil, id.String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.GetModule(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

func TestGetLesson_InvalidID(t *testing.T) {
	h := NewHandler(NewService(&mockRepo{}, nil))
	w, c := jsonRequest(http.MethodGet, "/lessons/bad", nil, "bad")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.GetLesson(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestGetLesson_NotFound(t *testing.T) {
	repo := &mockRepo{GetLessonByIDFn: func(ctx context.Context, id uuid.UUID) (*Lesson, error) {
		return nil, pgx.ErrNoRows
	}}
	h := NewHandler(NewService(repo, nil))
	id := uuid.New()
	w, c := jsonRequest(http.MethodGet, "/lessons/"+id.String(), nil, id.String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.GetLesson(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

func TestCreateLesson_InvalidLessonType(t *testing.T) {
	moduleID := uuid.New()
	repo := &mockRepo{
		GetModuleByIDFn: func(ctx context.Context, id uuid.UUID) (*Module, error) {
			return &Module{ID: id}, nil
		},
	}
	h := NewHandler(NewService(repo, nil))
	body := CreateLessonRequest{Title: "Lesson", LessonType: "invalid_type"}
	w, c := jsonRequest(http.MethodPost, "/modules/"+moduleID.String()+"/lessons", body, moduleID.String())
	c.Params = gin.Params{{Key: "id", Value: moduleID.String()}}
	setEditorContext(c)
	h.CreateLesson(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body = %s", w.Code, w.Body.String())
	}
}

func TestCreateModule_EmptyTitle(t *testing.T) {
	h := NewHandler(NewService(&mockRepo{}, nil))
	body := CreateModuleRequest{Title: ""}
	w, c := jsonRequest(http.MethodPost, "/modules", body, "")
	setEditorContext(c)
	h.CreateModule(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", w.Code)
	}
}

func TestDeleteModule_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{
		GetModuleByIDFn: func(ctx context.Context, mid uuid.UUID) (*Module, error) {
			return &Module{ID: mid}, nil
		},
		DeleteModuleFn: func(ctx context.Context, mid uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(NewService(repo, nil))
	w, c := jsonRequest(http.MethodDelete, "/modules/"+id.String(), nil, id.String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	setEditorContext(c)
	h.DeleteModule(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

func TestDeleteLesson_NotFound(t *testing.T) {
	id := uuid.New()
	moduleID := uuid.New()
	repo := &mockRepo{
		GetLessonByIDFn: func(ctx context.Context, lid uuid.UUID) (*Lesson, error) {
			return &Lesson{ID: lid, ModuleID: moduleID}, nil
		},
		GetModuleByIDFn: func(ctx context.Context, mid uuid.UUID) (*Module, error) {
			return &Module{ID: mid}, nil
		},
		DeleteLessonFn: func(ctx context.Context, lid uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(NewService(repo, nil))
	w, c := jsonRequest(http.MethodDelete, "/lessons/"+id.String(), nil, id.String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	setEditorContext(c)
	h.DeleteLesson(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404 when not deleted", w.Code)
	}
}

// --- GetLesson: free preview access control ---

// lessonRepoWithSortOrder returns a mockRepo whose GetLessonByIDFn returns a lesson
// with the given sort_order.
func lessonRepoWithSortOrder(sortOrder int) *mockRepo {
	return &mockRepo{
		GetLessonByIDFn: func(ctx context.Context, id uuid.UUID) (*Lesson, error) {
			return &Lesson{ID: id, SortOrder: sortOrder}, nil
		},
	}
}

func TestGetLesson_AnonymousFreeLessonAllowed(t *testing.T) {
	for sortOrder := 0; sortOrder < FreeLessonsCount; sortOrder++ {
		repo := lessonRepoWithSortOrder(sortOrder)
		h := NewHandler(NewService(repo, nil))
		id := uuid.New()
		w, c := jsonRequest(http.MethodGet, "/lessons/"+id.String(), nil, id.String())
		c.Params = gin.Params{{Key: "id", Value: id.String()}}
		// No user_id set → anonymous request
		h.GetLesson(c)
		if w.Code != http.StatusOK {
			t.Errorf("sort_order=%d: status = %d; want 200 for anonymous free lesson", sortOrder, w.Code)
		}
	}
}

func TestGetLesson_AnonymousLockedLessonForbidden(t *testing.T) {
	for _, sortOrder := range []int{FreeLessonsCount, FreeLessonsCount + 1, 10} {
		repo := lessonRepoWithSortOrder(sortOrder)
		h := NewHandler(NewService(repo, nil))
		id := uuid.New()
		w, c := jsonRequest(http.MethodGet, "/lessons/"+id.String(), nil, id.String())
		c.Params = gin.Params{{Key: "id", Value: id.String()}}
		// No user_id set → anonymous request
		h.GetLesson(c)
		if w.Code != http.StatusForbidden {
			t.Errorf("sort_order=%d: status = %d; want 403 for anonymous locked lesson", sortOrder, w.Code)
		}
	}
}

func TestGetLesson_AuthenticatedCanAccessLockedLesson(t *testing.T) {
	sortOrder := FreeLessonsCount + 5
	repo := lessonRepoWithSortOrder(sortOrder)
	h := NewHandler(NewService(repo, nil))
	id := uuid.New()
	w, c := jsonRequest(http.MethodGet, "/lessons/"+id.String(), nil, id.String())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	// Set user_id to simulate authenticated request
	c.Set(requestcontext.UserIDKey, uuid.New())
	h.GetLesson(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200 for authenticated user accessing locked lesson; body = %s", w.Code, w.Body.String())
	}
}
