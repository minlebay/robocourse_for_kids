package users

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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func init() { gin.SetMode(gin.TestMode) }

// --- mock ---

type mockRepo struct {
	CreateFn      func(ctx context.Context, login, passwordHash, name, role string) (*User, error)
	GetByLoginFn  func(ctx context.Context, login string) (*UserWithPassword, error)
	GetByIDFn     func(ctx context.Context, id uuid.UUID) (*User, error)
	ListFn        func(ctx context.Context) ([]User, error)
	DeleteFn      func(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateThemeFn func(ctx context.Context, id uuid.UUID, theme string) error
}

func (m *mockRepo) Create(ctx context.Context, login, passwordHash, name, role string) (*User, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, login, passwordHash, name, role)
	}
	return &User{ID: uuid.New(), Login: login, Name: name, Role: role, Theme: "default", CreatedAt: time.Now()}, nil
}

func (m *mockRepo) GetByLogin(ctx context.Context, login string) (*UserWithPassword, error) {
	if m.GetByLoginFn != nil {
		return m.GetByLoginFn(ctx, login)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) List(ctx context.Context) ([]User, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return false, nil
}

func (m *mockRepo) UpdateTheme(ctx context.Context, id uuid.UUID, theme string) error {
	if m.UpdateThemeFn != nil {
		return m.UpdateThemeFn(ctx, id, theme)
	}
	return nil
}

// --- helpers ---

func jsonRequest(method, target string, body interface{}) (*httptest.ResponseRecorder, *gin.Context) {
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

func parseJSON(w *httptest.ResponseRecorder) map[string]interface{} {
	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)
	return m
}

// ==================== RegisterUser ====================

func TestRegisterUser_Success(t *testing.T) {
	h := NewHandler(&mockRepo{}, "test-secret", "inv")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "testuser", Password: "password123", Name: "Test User",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["token"] == nil || resp["token"] == "" {
		t.Fatal("expected token in response")
	}
	user := resp["user"].(map[string]interface{})
	if user["role"] != RoleStudent {
		t.Fatalf("role = %v; want %s", user["role"], RoleStudent)
	}
}

func TestRegisterUser_TeacherSuccess(t *testing.T) {
	h := NewHandler(&mockRepo{}, "test-secret", "invite123")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "teacher1", Password: "password123", Name: "Teacher", Role: RoleTeacher, InviteCode: "invite123",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestRegisterUser_LoginTooShort(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "ab", Password: "password123", Name: "Test",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegisterUser_PasswordTooShort(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "testuser", Password: "12345", Name: "Test",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegisterUser_PasswordTooLong(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	long := make([]byte, 73)
	for i := range long {
		long[i] = 'a'
	}
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "testuser", Password: string(long), Name: "Test",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegisterUser_InvalidRole(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "testuser", Password: "password123", Name: "Test", Role: "admin",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegisterUser_TeacherDisabled(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "") // empty invite = disabled
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "teacher1", Password: "password123", Name: "Teacher", Role: RoleTeacher,
	})

	h.RegisterUser(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusForbidden)
	}
}

func TestRegisterUser_TeacherWrongInvite(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "correct")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "teacher1", Password: "password123", Name: "Teacher", Role: RoleTeacher, InviteCode: "wrong",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusForbidden)
	}
}

func TestRegisterUser_DuplicateLogin(t *testing.T) {
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, login, passwordHash, name, role string) (*User, error) {
			return nil, &pgconn.PgError{Code: "23505"}
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/register", RegisterRequest{
		Login: "existing", Password: "password123", Name: "Test",
	})

	h.RegisterUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["error"] != "login already exists" {
		t.Fatalf("error = %v; want 'login already exists'", resp["error"])
	}
}

// ==================== Login ====================

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	uid := uuid.New()
	repo := &mockRepo{
		GetByLoginFn: func(ctx context.Context, login string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User:         User{ID: uid, Login: login, Name: "Test", Role: RoleStudent, Theme: "default", CreatedAt: time.Now()},
				PasswordHash: string(hash),
			}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/login", LoginRequest{Login: "testuser", Password: "password123"})

	h.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["token"] == nil || resp["token"] == "" {
		t.Fatal("expected token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	repo := &mockRepo{
		GetByLoginFn: func(ctx context.Context, login string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User:         User{ID: uuid.New(), Login: login, Name: "Test", Role: RoleStudent, CreatedAt: time.Now()},
				PasswordHash: string(hash),
			}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/login", LoginRequest{Login: "user", Password: "wrong"})

	h.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	h := NewHandler(&mockRepo{}, "test-secret", "") // default returns pgx.ErrNoRows
	w, c := jsonRequest(http.MethodPost, "/login", LoginRequest{Login: "noone", Password: "password123"})

	h.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

// ==================== Me ====================

func TestMe_Success(t *testing.T) {
	uid := uuid.New()
	repo := &mockRepo{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return &User{ID: id, Login: "me", Name: "Me", Role: RoleStudent, Theme: "default", CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(repo, "s", "")
	w, c := jsonRequest(http.MethodGet, "/me", nil)
	c.Set("user_id", uid)

	h.Me(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestMe_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	w, c := jsonRequest(http.MethodGet, "/me", nil)

	h.Me(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

// ==================== ListUsers ====================

func TestListUsers_Success(t *testing.T) {
	repo := &mockRepo{
		ListFn: func(ctx context.Context) ([]User, error) {
			return []User{
				{ID: uuid.New(), Login: "a", Name: "A", Role: RoleStudent},
				{ID: uuid.New(), Login: "b", Name: "B", Role: RoleTeacher},
			}, nil
		},
	}
	h := NewHandler(repo, "s", "")
	w, c := jsonRequest(http.MethodGet, "/users", nil)

	h.ListUsers(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusOK)
	}
	var list []User
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("len = %d; want 2", len(list))
	}
}

// ==================== DeleteUser ====================

func TestDeleteUser_Success(t *testing.T) {
	repo := &mockRepo{
		DeleteFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	h := NewHandler(repo, "s", "")
	myID := uuid.New()
	targetID := uuid.New()

	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) { c.Set("user_id", myID); c.Next() }, h.DeleteUser)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusNoContent, w.Body.String())
	}
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	myID := uuid.New()
	w, c := jsonRequest(http.MethodDelete, "/users/"+myID.String(), nil)
	c.Set("user_id", myID)
	c.Params = gin.Params{{Key: "id", Value: myID.String()}}

	h.DeleteUser(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusForbidden)
	}
}

// ==================== UpdateMe ====================

func TestUpdateMe_Success(t *testing.T) {
	uid := uuid.New()
	repo := &mockRepo{
		UpdateThemeFn: func(ctx context.Context, id uuid.UUID, theme string) error { return nil },
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return &User{ID: id, Login: "me", Name: "Me", Role: RoleStudent, Theme: "cyberpunk", CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(repo, "s", "")
	w, c := jsonRequest(http.MethodPut, "/me", UpdateMeRequest{Theme: "cyberpunk"})
	c.Set("user_id", uid)

	h.UpdateMe(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUpdateMe_InvalidTheme(t *testing.T) {
	h := NewHandler(&mockRepo{}, "s", "")
	w, c := jsonRequest(http.MethodPut, "/me", UpdateMeRequest{Theme: "neon"})
	c.Set("user_id", uuid.New())

	h.UpdateMe(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}
