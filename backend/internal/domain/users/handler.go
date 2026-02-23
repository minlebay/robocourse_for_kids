package users

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/requestcontext"
)

// Repository defines the data access interface for users.
type Repository interface {
	Create(ctx context.Context, login, passwordHash, name, role, email string) (*User, error)
	GetByLogin(ctx context.Context, login string) (*UserWithPassword, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context) ([]User, error)
	ListAll(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateTheme(ctx context.Context, id uuid.UUID, theme string) error
	BlockUser(ctx context.Context, id uuid.UUID, block bool) error
	SetMustChangePassword(ctx context.Context, id uuid.UUID, v bool) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error
	GetStats(ctx context.Context) (usersCount, modulesCount, lessonsCount int, err error)
	GetActivity(ctx context.Context, limit int) ([]User, error)
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
	user, err := h.repo.Create(c.Request.Context(), req.Login, string(hash), req.Name, req.Role, "")
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
	token, err := h.generateToken(user.ID, user.Role, user.MustChangePassword)
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
	if u.IsBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is blocked"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}
	token, err := h.generateToken(u.ID, u.Role, u.MustChangePassword)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": &u.User, "token": token})
}

func (h *Handler) Me(c *gin.Context) {
	uid := requestcontext.GetUserID(c)
	if uid == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid)
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
	uid := requestcontext.GetUserID(c)
	if uid == uuid.Nil {
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
	if err := h.repo.UpdateTheme(c.Request.Context(), uid, req.Theme); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid)
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
	currentID := requestcontext.GetUserID(c)
	if currentID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentID == targetID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete yourself"})
		return
	}

	target, err := h.repo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if target.Role != RoleStudent {
		c.JSON(http.StatusForbidden, gin.H{"error": "teachers can only delete students"})
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
	httplog.LogAudit(c, "delete_user", "actor", currentID, "target", targetID)
	c.Status(http.StatusNoContent)
}

// ChangePasswordRequest is the body for POST /api/v1/auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// ChangePassword handles POST /api/v1/auth/change-password.
func (h *Handler) ChangePassword(c *gin.Context) {
	uid := requestcontext.GetUserID(c)
	if uid == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}
	if len(req.NewPassword) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at most 72 characters"})
		return
	}

	// Load the user to verify current password
	u, err := h.repo.GetByID(c.Request.Context(), uid)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Fetch user with password hash
	uwp, err := h.repo.GetByLogin(c.Request.Context(), u.Login)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(uwp.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.repo.UpdatePassword(c.Request.Context(), uid, string(newHash)); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.repo.SetMustChangePassword(c.Request.Context(), uid, false); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	token, err := h.generateToken(uid, u.Role, false)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// AdminListUsers handles GET /api/v1/admin/users.
func (h *Handler) AdminListUsers(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// AdminCreateUserRequest is the body for POST /api/v1/admin/users.
type AdminCreateUserRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

// AdminCreateUser handles POST /api/v1/admin/users.
func (h *Handler) AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Login) < 3 || len(req.Login) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login must be 3-50 characters"})
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}
	if len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at most 72 characters"})
		return
	}
	if len(req.Name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be at most 200 characters"})
		return
	}

	if req.Role == "" {
		req.Role = RoleStudent
	}
	if req.Role != RoleStudent && req.Role != RoleTeacher && req.Role != RoleAdministrator {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be student, teacher, or administrator"})
		return
	}

	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := h.repo.Create(c.Request.Context(), req.Login, string(hash), req.Name, req.Role, req.Email)
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

	if err := h.repo.SetMustChangePassword(c.Request.Context(), user.ID, true); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	user.MustChangePassword = true

	c.JSON(http.StatusCreated, gin.H{"user": user, "temp_password": req.Password})
}

// AdminDeleteUser handles DELETE /api/v1/admin/users/:id.
func (h *Handler) AdminDeleteUser(c *gin.Context) {
	currentID := requestcontext.GetUserID(c)
	if currentID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentID == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	deleted, err := h.repo.Delete(c.Request.Context(), targetID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	httplog.LogAudit(c, "admin_delete_user", "actor", currentID, "target", targetID)
	c.Status(http.StatusNoContent)
}

// AdminBlockUserRequest is the body for POST /api/v1/admin/users/:id/block.
type AdminBlockUserRequest struct {
	Block bool `json:"block"`
}

// AdminBlockUser handles POST /api/v1/admin/users/:id/block.
func (h *Handler) AdminBlockUser(c *gin.Context) {
	currentID := requestcontext.GetUserID(c)
	if currentID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentID == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
		return
	}

	var req AdminBlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.BlockUser(c.Request.Context(), targetID, req.Block); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	httplog.LogAudit(c, "admin_block_user", "actor", currentID, "target", targetID, "blocked", req.Block)
	c.JSON(http.StatusOK, gin.H{"blocked": req.Block})
}

// AdminResetPassword handles POST /api/v1/admin/users/:id/reset-password.
func (h *Handler) AdminResetPassword(c *gin.Context) {
	actorID := requestcontext.GetUserID(c)

	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	tempPassword := generateTempPassword(10)
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.repo.UpdatePassword(c.Request.Context(), targetID, string(hash)); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.repo.SetMustChangePassword(c.Request.Context(), targetID, true); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	httplog.LogAudit(c, "admin_reset_password", "actor", actorID, "target", targetID)
	c.JSON(http.StatusOK, gin.H{"temp_password": tempPassword})
}

// AdminGetStats handles GET /api/v1/admin/stats.
func (h *Handler) AdminGetStats(c *gin.Context) {
	usersCount, modulesCount, lessonsCount, err := h.repo.GetStats(c.Request.Context())
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"users":   usersCount,
		"modules": modulesCount,
		"lessons": lessonsCount,
	})
}

// AdminGetActivity handles GET /api/v1/admin/activity.
func (h *Handler) AdminGetActivity(c *gin.Context) {
	list, err := h.repo.GetActivity(c.Request.Context(), 20)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// generateTempPassword generates a cryptographically random password of given length.
func generateTempPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
