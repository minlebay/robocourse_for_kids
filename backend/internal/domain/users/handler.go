package users

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"learn_kids/backend/internal/httplog"
)

// Repository defines the data access interface for users.
type Repository interface {
	Create(ctx context.Context, login, passwordHash, name, role string) (*User, error)
	GetByLogin(ctx context.Context, login string) (*UserWithPassword, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateTheme(ctx context.Context, id uuid.UUID, theme string) error
}

// dummyHash is a pre-computed bcrypt hash used for timing attack protection.
// When a login attempt is made for a non-existent user, we still run bcrypt
// to ensure consistent response times and prevent user enumeration.
var dummyHash []byte

func init() {
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-attack-protection"), bcrypt.DefaultCost)
}

type Handler struct {
	repo       Repository
	jwtKey     []byte
	inviteCode string // required for teacher registration; empty = disabled
}

func NewHandler(repo Repository, jwtSecret, inviteCode string) *Handler {
	return &Handler{repo: repo, jwtKey: []byte(jwtSecret), inviteCode: inviteCode}
}

var validThemes = map[string]bool{
	"default": true, "light": true, "cyberpunk": true,
	"contrast-light": true, "contrast-dark": true,
	"cream": true, "snow": true, "midnight": true, "forest": true,
}

type RegisterRequest struct {
	Login      string `json:"login" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Role       string `json:"role"`
	InviteCode string `json:"invite_code"`
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// --- Validate login length ---
	if len(req.Login) < 3 || len(req.Login) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login must be 3-50 characters"})
		return
	}

	// --- Validate password length (bcrypt silently truncates at 72 bytes) ---
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}
	if len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at most 72 characters"})
		return
	}

	// --- Validate & authorize role ---
	if req.Role == "" {
		req.Role = RoleStudent
	}
	if req.Role != RoleStudent && req.Role != RoleTeacher {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be student or teacher"})
		return
	}
	if req.Role == RoleTeacher {
		if h.inviteCode == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "teacher registration is not available"})
			return
		}
		if req.InviteCode != h.inviteCode {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid invite code"})
			return
		}
	}

	// --- Validate name length ---
	if len(req.Name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be at most 200 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	user, err := h.repo.Create(c.Request.Context(), req.Login, string(hash), req.Name, req.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "login already exists"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	token, err := h.generateToken(user.ID, user.Role)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "token": token})
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.repo.GetByLogin(c.Request.Context(), req.Login)
	if err != nil {
		// Timing attack protection: always run bcrypt even if user not found
		bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}
	user := &User{ID: u.ID, Login: u.Login, Name: u.Name, Role: u.Role, Theme: u.Theme, CreatedAt: u.CreatedAt}
	token, err := h.generateToken(user.ID, user.Role)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (h *Handler) Me(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists || uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid.(uuid.UUID))
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

type UpdateMeRequest struct {
	Theme string `json:"theme"`
}

func (h *Handler) UpdateMe(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists || uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Theme == "" || !validThemes[req.Theme] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "theme must be one of: default, light, cyberpunk, contrast-light, contrast-dark, cream, snow, midnight, forest"})
		return
	}
	if err := h.repo.UpdateTheme(c.Request.Context(), uid.(uuid.UUID), req.Theme); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid.(uuid.UUID))
	if err != nil || u == nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) ListUsers(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	uid, exists := c.Get("user_id")
	if !exists || uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	currentID := uid.(uuid.UUID)

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentID == targetID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete yourself"})
		return
	}

	deleted, err := h.repo.Delete(ctx, targetID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
