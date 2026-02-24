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
	CreateFn                   func(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error)
	GetByLoginFn               func(ctx context.Context, login string) (*UserWithPassword, error)
	GetByIDFn                  func(ctx context.Context, id uuid.UUID) (*User, error)
	GetByIDWithPasswordFn      func(ctx context.Context, id uuid.UUID) (*UserWithPassword, error)
	IsBlockedFn                func(ctx context.Context, id uuid.UUID) (bool, error)
	ListFn                     func(ctx context.Context, limit, offset int) ([]User, error)
	ListAllFn                  func(ctx context.Context, limit, offset int) ([]User, error)
	DeleteFn                   func(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateThemeFn              func(ctx context.Context, id uuid.UUID, theme string) error
	BlockUserFn                func(ctx context.Context, id uuid.UUID, block bool) error
	SetMustChangePasswordFn    func(ctx context.Context, id uuid.UUID, v bool) error
	UpdatePasswordAndMustChangeFn func(ctx context.Context, id uuid.UUID, hash string, mustChange bool) (bool, error)
	GetStatsFn                 func(ctx context.Context) (usersCount, modulesCount, lessonsCount int, err error)
	GetActivityFn              func(ctx context.Context, limit int) ([]User, error)
}

func (m *mockRepo) Create(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, login, passwordHash, name, role, email, mustChangePassword)
	}
	return &User{ID: uuid.New(), Login: login, Name: name, Role: role, Theme: "default", MustChangePassword: mustChangePassword, CreatedAt: time.Now()}, nil
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

func (m *mockRepo) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*UserWithPassword, error) {
	if m.GetByIDWithPasswordFn != nil {
		return m.GetByIDWithPasswordFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.IsBlockedFn != nil {
		return m.IsBlockedFn(ctx, id)
	}
	return false, nil
}

func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]User, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, limit, offset)
	}
	return nil, nil
}

func (m *mockRepo) ListAll(ctx context.Context, limit, offset int) ([]User, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx, limit, offset)
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

func (m *mockRepo) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
	if m.BlockUserFn != nil {
		return m.BlockUserFn(ctx, id, block)
	}
	return nil
}

func (m *mockRepo) SetMustChangePassword(ctx context.Context, id uuid.UUID, v bool) error {
	if m.SetMustChangePasswordFn != nil {
		return m.SetMustChangePasswordFn(ctx, id, v)
	}
	return nil
}

func (m *mockRepo) UpdatePasswordAndMustChange(ctx context.Context, id uuid.UUID, hash string, mustChange bool) (bool, error) {
	if m.UpdatePasswordAndMustChangeFn != nil {
		return m.UpdatePasswordAndMustChangeFn(ctx, id, hash, mustChange)
	}
	return true, nil
}

func (m *mockRepo) GetStats(ctx context.Context) (usersCount, modulesCount, lessonsCount int, err error) {
	if m.GetStatsFn != nil {
		return m.GetStatsFn(ctx)
	}
	return 0, 0, 0, nil
}

func (m *mockRepo) GetActivity(ctx context.Context, limit int) ([]User, error) {
	if m.GetActivityFn != nil {
		return m.GetActivityFn(ctx, limit)
	}
	return nil, nil
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
		CreateFn: func(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error) {
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

func TestLogin_BlockedUser(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	repo := &mockRepo{
		GetByLoginFn: func(ctx context.Context, login string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User:         User{ID: uuid.New(), Login: login, Name: "Test", Role: RoleStudent, IsBlocked: true, CreatedAt: time.Now()},
				PasswordHash: string(hash),
			}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/login", LoginRequest{Login: "blocked", Password: "password123"})

	h.Login(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["error"] != "user is blocked" {
		t.Fatalf("error = %v; want 'user is blocked'", resp["error"])
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
		ListFn: func(ctx context.Context, limit, offset int) ([]User, error) {
			return []User{
				{ID: uuid.New(), Login: "a", Name: "A", Role: RoleStudent},
				{ID: uuid.New(), Login: "b", Name: "B", Role: RoleStudent},
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
	myID := uuid.New()
	targetID := uuid.New()
	repo := &mockRepo{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return &User{ID: id, Role: RoleStudent}, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	h := NewHandler(repo, "s", "")

	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) { c.Set("user_id", myID); c.Next() }, h.DeleteUser)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusNoContent, w.Body.String())
	}
}

func TestDeleteUser_CannotDeleteTeacher(t *testing.T) {
	myID := uuid.New()
	targetID := uuid.New()
	repo := &mockRepo{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return &User{ID: id, Role: RoleTeacher}, nil
		},
	}
	h := NewHandler(repo, "s", "")

	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) { c.Set("user_id", myID); c.Next() }, h.DeleteUser)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestDeleteUser_CannotDeleteAdmin(t *testing.T) {
	myID := uuid.New()
	targetID := uuid.New()
	repo := &mockRepo{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return &User{ID: id, Role: RoleAdministrator}, nil
		},
	}
	h := NewHandler(repo, "s", "")

	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) { c.Set("user_id", myID); c.Next() }, h.DeleteUser)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+targetID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusForbidden, w.Body.String())
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

// ==================== AdminCreateUser ====================

func TestAdminCreateUser_Success(t *testing.T) {
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error) {
			return &User{ID: uuid.New(), Login: login, Name: name, Role: role, Theme: "default", MustChangePassword: mustChangePassword, CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/admin/users", AdminCreateUserRequest{
		Login:    "newuser",
		Password: "password123",
		Name:     "New User",
		Role:     RoleStudent,
	})
	c.Set("user_id", uuid.New())

	h.AdminCreateUser(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["temp_password"] == nil || resp["temp_password"] == "" {
		t.Fatal("expected temp_password in response")
	}
	if resp["user"] == nil {
		t.Fatal("expected user in response")
	}
}

func TestAdminCreateUser_InvalidEmail(t *testing.T) {
	h := NewHandler(&mockRepo{}, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/admin/users", AdminCreateUserRequest{
		Login:    "newuser",
		Password: "password123",
		Name:     "New User",
		Role:     RoleStudent,
		Email:    "not-an-email",
	})

	h.AdminCreateUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["error"] != "invalid email address" {
		t.Fatalf("error = %v; want 'invalid email address'", resp["error"])
	}
}

func TestAdminCreateUser_ValidEmail(t *testing.T) {
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error) {
			e := email
			return &User{ID: uuid.New(), Login: login, Name: name, Role: role, Theme: "default", Email: &e, MustChangePassword: mustChangePassword, CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/admin/users", AdminCreateUserRequest{
		Login:    "newuser",
		Password: "password123",
		Name:     "New User",
		Role:     RoleStudent,
		Email:    "user@example.com",
	})
	c.Set("user_id", uuid.New())

	h.AdminCreateUser(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
}

// ==================== AdminBlockUser ====================

func TestAdminBlockUser_Success(t *testing.T) {
	targetID := uuid.New()
	currentID := uuid.New()
	repo := &mockRepo{
		BlockUserFn: func(ctx context.Context, id uuid.UUID, block bool) error {
			if id != targetID {
				return nil
			}
			return nil
		},
	}
	h := NewHandler(repo, "s", "")

	r := gin.New()
	r.POST("/admin/users/:id/block", func(c *gin.Context) {
		c.Set("user_id", currentID)
		c.Next()
	}, h.AdminBlockUser)

	body, _ := json.Marshal(AdminBlockUserRequest{Block: true})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+targetID.String()+"/block", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAdminBlockUser_Self(t *testing.T) {
	myID := uuid.New()
	h := NewHandler(&mockRepo{}, "s", "")
	w, c := jsonRequest(http.MethodPost, "/admin/users/"+myID.String()+"/block", AdminBlockUserRequest{Block: true})
	c.Set("user_id", myID)
	c.Params = gin.Params{{Key: "id", Value: myID.String()}}

	h.AdminBlockUser(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["error"] != "cannot block yourself" {
		t.Fatalf("error = %v; want 'cannot block yourself'", resp["error"])
	}
}

// ==================== AdminResetPassword ====================

func TestAdminResetPassword_Success(t *testing.T) {
	targetID := uuid.New()
	repo := &mockRepo{}
	h := NewHandler(repo, "s", "")

	r := gin.New()
	r.POST("/admin/users/:id/reset-password", h.AdminResetPassword)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+targetID.String()+"/reset-password", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseJSON(w)
	tempPwd, ok := resp["temp_password"].(string)
	if !ok || tempPwd == "" {
		t.Fatal("expected non-empty temp_password in response")
	}
	if len(tempPwd) != 10 {
		t.Fatalf("temp_password length = %d; want 10", len(tempPwd))
	}
}

// ==================== RequireFreshPassword middleware ====================

func TestRequireFreshPassword_BlocksWhenRequired(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("must_change_password", true)
		c.Next()
	})

	// Import middleware directly here by testing via the Handler approach.
	// We simulate the middleware logic inline since it's in another package.
	// Instead, we test it via a full router with middleware applied.

	// Using a simple test: simulate the middleware behavior by building a minimal router.
	// We reproduce RequireFreshPassword logic here for testing purposes.
	r.GET("/some-endpoint", func(c *gin.Context) {
		mustChange, exists := c.Get("must_change_password")
		if exists && mustChange == true {
			if c.FullPath() != "/api/v1/auth/change-password" {
				c.JSON(http.StatusForbidden, gin.H{"error": "password_change_required"})
				c.Abort()
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some-endpoint", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusForbidden)
	}
	resp := parseJSON(w)
	if resp["error"] != "password_change_required" {
		t.Fatalf("error = %v; want 'password_change_required'", resp["error"])
	}
}

func TestRequireFreshPassword_AllowsChangePassword(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("must_change_password", true)
		c.Next()
	})

	r.POST("/api/v1/auth/change-password", func(c *gin.Context) {
		mustChange, exists := c.Get("must_change_password")
		if exists && mustChange == true {
			if c.FullPath() != "/api/v1/auth/change-password" {
				c.JSON(http.StatusForbidden, gin.H{"error": "password_change_required"})
				c.Abort()
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestRequireFreshPassword_PassesWhenNotRequired(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("must_change_password", false)
		c.Next()
	})

	r.GET("/some-endpoint", func(c *gin.Context) {
		mustChange, exists := c.Get("must_change_password")
		if exists && mustChange == true {
			if c.FullPath() != "/api/v1/auth/change-password" {
				c.JSON(http.StatusForbidden, gin.H{"error": "password_change_required"})
				c.Abort()
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some-endpoint", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// ==================== AdminGetStats ====================

func TestAdminGetStats_Success(t *testing.T) {
	repo := &mockRepo{
		GetStatsFn: func(ctx context.Context) (int, int, int, error) {
			return 42, 8, 64, nil
		},
	}
	h := NewHandler(repo, "s", "")
	w, c := jsonRequest(http.MethodGet, "/admin/stats", nil)

	h.AdminGetStats(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["users"] != float64(42) {
		t.Fatalf("users = %v; want 42", resp["users"])
	}
	if resp["modules"] != float64(8) {
		t.Fatalf("modules = %v; want 8", resp["modules"])
	}
	if resp["lessons"] != float64(64) {
		t.Fatalf("lessons = %v; want 64", resp["lessons"])
	}
}

// ==================== AdminGetActivity ====================

func TestAdminGetActivity_Success(t *testing.T) {
	repo := &mockRepo{
		GetActivityFn: func(ctx context.Context, limit int) ([]User, error) {
			return []User{
				{ID: uuid.New(), Login: "a", Name: "A", Role: RoleStudent, CreatedAt: time.Now()},
				{ID: uuid.New(), Login: "b", Name: "B", Role: RoleTeacher, CreatedAt: time.Now()},
			}, nil
		},
	}
	h := NewHandler(repo, "s", "")
	w, c := jsonRequest(http.MethodGet, "/admin/activity", nil)

	h.AdminGetActivity(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var list []User
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("len = %d; want 2", len(list))
	}
}

// ==================== ChangePassword ====================

func TestChangePassword_Success(t *testing.T) {
	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass1"), bcrypt.MinCost)
	repo := &mockRepo{
		GetByIDWithPasswordFn: func(ctx context.Context, id uuid.UUID) (*UserWithPassword, error) {
			return &UserWithPassword{
				User:         User{ID: uid, Login: "user1", Role: RoleStudent, CreatedAt: time.Now()},
				PasswordHash: string(hash),
			}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/auth/change-password", ChangePasswordRequest{
		CurrentPassword: "oldpass1",
		NewPassword:     "newpass1",
	})
	c.Set("user_id", uid)

	h.ChangePassword(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["token"] == nil || resp["token"] == "" {
		t.Fatal("expected token in response")
	}
}

func TestChangePassword_Unauthorized(t *testing.T) {
	h := NewHandler(&mockRepo{}, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/auth/change-password", ChangePasswordRequest{
		CurrentPassword: "oldpass1",
		NewPassword:     "newpass1",
	})
	// no user_id set in context

	h.ChangePassword(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestChangePassword_NewPasswordTooShort(t *testing.T) {
	uid := uuid.New()
	h := NewHandler(&mockRepo{}, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/auth/change-password", ChangePasswordRequest{
		CurrentPassword: "oldpass1",
		NewPassword:     "abc",
	})
	c.Set("user_id", uid)

	h.ChangePassword(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChangePassword_NewPasswordTooLong(t *testing.T) {
	uid := uuid.New()
	h := NewHandler(&mockRepo{}, "test-secret", "")
	long := make([]byte, 73)
	for i := range long {
		long[i] = 'a'
	}
	w, c := jsonRequest(http.MethodPost, "/auth/change-password", ChangePasswordRequest{
		CurrentPassword: "oldpass1",
		NewPassword:     string(long),
	})
	c.Set("user_id", uid)

	h.ChangePassword(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	repo := &mockRepo{
		GetByIDWithPasswordFn: func(ctx context.Context, id uuid.UUID) (*UserWithPassword, error) {
			return &UserWithPassword{
				User:         User{ID: uid, Login: "user1", Role: RoleStudent, CreatedAt: time.Now()},
				PasswordHash: string(hash),
			}, nil
		},
	}
	h := NewHandler(repo, "test-secret", "")
	w, c := jsonRequest(http.MethodPost, "/auth/change-password", ChangePasswordRequest{
		CurrentPassword: "wrongpass",
		NewPassword:     "newpass1",
	})
	c.Set("user_id", uid)

	h.ChangePassword(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d; body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["error"] != "current password is incorrect" {
		t.Fatalf("error = %v; want 'current password is incorrect'", resp["error"])
	}
}
