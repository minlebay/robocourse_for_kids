package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/requestcontext"
	"learn_kids/backend/internal/testutil"
)

func init() { gin.SetMode(gin.TestMode) }

// helper: builds a router with the given setup middleware (for seeding context values)
// followed by the middleware under test, and a final handler that always returns 200.
func newRouter(seed gin.HandlerFunc, mw gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	if seed != nil {
		r.Use(seed)
	}
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/api/v1/auth/change-password", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func get(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

// ==================== RequireAuth ====================

func TestRequireAuth_AllowsAuthenticatedUser(t *testing.T) {
	uid := uuid.New()
	r := newRouter(
		func(c *gin.Context) { c.Set("user_id", uid); c.Next() },
		RequireAuth(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestRequireAuth_RejectsAnonymous(t *testing.T) {
	r := newRouter(nil, RequireAuth())
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusUnauthorized)
	resp := testutil.ParseJSON(t, w)
	if resp["error"] != "unauthorized" {
		t.Errorf("error = %v; want 'unauthorized'", resp["error"])
	}
}

// ==================== RequireTeacher ====================

func TestRequireTeacher_AllowsTeacher(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("user_roles", []string{"teacher"}); c.Next() },
		RequireTeacher(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestRequireTeacher_AllowsAdministrator(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("user_roles", []string{"administrator"}); c.Next() },
		RequireTeacher(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestRequireTeacher_RejectsStudent(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("user_roles", []string{"student"}); c.Next() },
		RequireTeacher(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusForbidden)
	resp := testutil.ParseJSON(t, w)
	if resp["error"] != "insufficient role" {
		t.Errorf("error = %v; want 'insufficient role'", resp["error"])
	}
}

func TestRequireTeacher_RejectsNoRole(t *testing.T) {
	r := newRouter(nil, RequireTeacher())
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusUnauthorized)
}

// ==================== RequireAdmin ====================

func TestRequireAdmin_AllowsAdministrator(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("user_roles", []string{"administrator"}); c.Next() },
		RequireAdmin(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestRequireAdmin_RejectsTeacher(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("user_roles", []string{"teacher"}); c.Next() },
		RequireAdmin(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusForbidden)
	resp := testutil.ParseJSON(t, w)
	if resp["error"] != "admin access required" {
		t.Errorf("error = %v; want 'admin access required'", resp["error"])
	}
}

func TestRequireAdmin_RejectsStudent(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("user_roles", []string{"student"}); c.Next() },
		RequireAdmin(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusForbidden)
}

func TestRequireAdmin_RejectsNoRole(t *testing.T) {
	r := newRouter(nil, RequireAdmin())
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusUnauthorized)
}

// ==================== RequireFreshPassword ====================

func TestRequireFreshPassword_BlocksWhenRequired(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("must_change_password", true); c.Next() },
		RequireFreshPassword(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusForbidden)
	resp := testutil.ParseJSON(t, w)
	if resp["error"] != "password_change_required" {
		t.Errorf("error = %v; want 'password_change_required'", resp["error"])
	}
}

func TestRequireFreshPassword_AllowsChangePasswordPath(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("must_change_password", true); c.Next() },
		RequireFreshPassword(),
	)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", nil)
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestRequireFreshPassword_PassesWhenFlagFalse(t *testing.T) {
	r := newRouter(
		func(c *gin.Context) { c.Set("must_change_password", false); c.Next() },
		RequireFreshPassword(),
	)
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestRequireFreshPassword_PassesWhenFlagAbsent(t *testing.T) {
	r := newRouter(nil, RequireFreshPassword())
	w := get(r, "/test")
	testutil.AssertStatus(t, w, http.StatusOK)
}

// ==================== RateLimiter ====================

func TestRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute, nil)
	r := gin.New()
	r.Use(rl.Handler())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d; want %d", i+1, w.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_Blocks429WhenExceeded(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute, nil)
	r := gin.New()
	r.Use(rl.Handler())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "1.2.3.5:5678"
		r.ServeHTTP(w, req)
	}

	// Third request must be blocked.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "1.2.3.5:5678"
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusTooManyRequests)
}

func TestRateLimiter_TracksIPsSeparately(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, nil)
	r := gin.New()
	r.Use(rl.Handler())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Exhaust limit for IP A.
	for i := 0; i < 1; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1111"
		r.ServeHTTP(w, req)
	}

	// IP B must still be allowed.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.2:2222"
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusOK)
}

// ==================== Auth: IsUserBlocked error ====================

type mockAuthProvider struct {
	parseTokenFn    func(string) (uuid.UUID, []string, bool, time.Time, error)
	isUserBlockedFn func(context.Context, uuid.UUID) (bool, error)
	newTokenFn      func(uuid.UUID, []string, bool) (string, error)
}

func (m *mockAuthProvider) ParseToken(s string) (uuid.UUID, []string, bool, time.Time, error) {
	if m.parseTokenFn != nil {
		return m.parseTokenFn(s)
	}
	return uuid.Nil, nil, false, time.Time{}, errors.New("invalid")
}

func (m *mockAuthProvider) IsUserBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.isUserBlockedFn != nil {
		return m.isUserBlockedFn(ctx, id)
	}
	return false, nil
}

func (m *mockAuthProvider) NewToken(userID uuid.UUID, roles []string, mustChangePassword bool) (string, error) {
	if m.newTokenFn != nil {
		return m.newTokenFn(userID, roles, mustChangePassword)
	}
	return "", nil
}

func TestAuth_Returns401WhenIsUserBlockedReturnsError(t *testing.T) {
	// Valid token parse, but IsUserBlocked returns error (e.g. user deleted) → 401
	mock := &mockAuthProvider{
		parseTokenFn: func(string) (uuid.UUID, []string, bool, time.Time, error) {
			return uuid.New(), []string{"student"}, false, time.Now().Add(time.Hour), nil
		},
		isUserBlockedFn: func(context.Context, uuid.UUID) (bool, error) {
			return false, errors.New("user not found")
		},
	}
	r := gin.New()
	r.Use(Auth(mock))
	r.GET("/test", func(c *gin.Context) {
		if _, ok := c.Get(requestcontext.UserIDKey); ok {
			t.Error("user_id must not be set when IsUserBlocked returns error")
		}
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer any-token")
	r.ServeHTTP(w, req)
	testutil.AssertStatus(t, w, http.StatusUnauthorized)
	resp := testutil.ParseJSON(t, w)
	if resp["error"] != "invalid or expired token" {
		t.Errorf("error = %v; want 'invalid or expired token'", resp["error"])
	}
}
