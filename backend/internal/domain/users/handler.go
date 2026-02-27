package users

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/requestcontext"
)

// Repository defines the data access interface for users.
type Repository interface {
	Create(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error)
	GetByLogin(ctx context.Context, login string) (*UserWithPassword, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*UserWithPassword, error)
	IsBlocked(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	ListAll(ctx context.Context, limit, offset int) ([]User, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateTheme(ctx context.Context, id uuid.UUID, theme string) error
	BlockUser(ctx context.Context, id uuid.UUID, block bool) error
	SetMustChangePassword(ctx context.Context, id uuid.UUID, v bool) error
	UpdatePasswordAndMustChange(ctx context.Context, id uuid.UUID, hash string, mustChange bool) (bool, error)
	GetStats(ctx context.Context) (usersCount, modulesCount, lessonsCount int, err error)
	GetActivity(ctx context.Context, limit int) ([]User, error)
}

type Handler struct {
	svc    *Service
	jwtKey []byte
}

func NewHandler(svc *Service, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtKey: []byte(jwtSecret)}
}

// IsUserBlocked reports whether the user with the given ID is blocked.
// Used by the Auth middleware to reject blocked users on every request.
func (h *Handler) IsUserBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	return h.svc.IsBlocked(ctx, id)
}

// handleErr maps service errors to HTTP responses. Returns true if the error was handled.
func handleErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var he *HTTPError
	if errors.As(err, &he) {
		c.JSON(he.Code, gin.H{"error": he.Message})
		return true
	}
	httplog.LogError(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	return true
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, err := h.svc.Register(c.Request.Context(), req.Login, req.Password, req.Name, req.Role, req.InviteCode, "", false)
	if handleErr(c, err) {
		return
	}
	token, err := h.generateToken(user.ID, user.Role, user.MustChangePassword)
	if handleErr(c, err) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	u, err := h.svc.Authenticate(c.Request.Context(), req.Login, req.Password)
	if handleErr(c, err) {
		return
	}
	token, err := h.generateToken(u.ID, u.Role, u.MustChangePassword)
	if handleErr(c, err) {
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
	u, err := h.svc.GetByID(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Theme == "" || !validThemes[req.Theme] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "theme must be one of: default, light, cyberpunk, contrast-light, contrast-dark, cream, snow, midnight, forest"})
		return
	}
	if err := h.svc.UpdateTheme(c.Request.Context(), uid, req.Theme); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), uid)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) ListUsers(c *gin.Context) {
	limit, offset := parsePagination(c, 500)
	list, err := h.svc.List(c.Request.Context(), limit, offset)
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
	if _, err := h.svc.ValidateDeleteUser(ctx, currentID, targetID); handleErr(c, err) {
		return
	}
	deleted, err := h.svc.Delete(ctx, targetID)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	role, err := h.svc.ChangePassword(c.Request.Context(), uid, req.CurrentPassword, req.NewPassword)
	if handleErr(c, err) {
		return
	}
	httplog.LogAudit(c, "change_password", "user", uid)
	token, err := h.generateToken(uid, role, false)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// AdminListUsers handles GET /api/v1/admin/users.
func (h *Handler) AdminListUsers(c *gin.Context) {
	limit, offset := parsePagination(c, 500)
	list, err := h.svc.ListAll(c.Request.Context(), limit, offset)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, err := h.svc.AdminCreate(c.Request.Context(), req.Login, req.Password, req.Name, req.Role, req.Email)
	if handleErr(c, err) {
		return
	}
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
	deleted, err := h.svc.Delete(c.Request.Context(), targetID)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.svc.BlockUser(c.Request.Context(), targetID, req.Block); err != nil {
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
	tempPassword, err := h.svc.ResetPassword(c.Request.Context(), targetID)
	if handleErr(c, err) {
		return
	}
	httplog.LogAudit(c, "admin_reset_password", "actor", actorID, "target", targetID)
	c.JSON(http.StatusOK, gin.H{"temp_password": tempPassword})
}

// AdminGetStats handles GET /api/v1/admin/stats.
func (h *Handler) AdminGetStats(c *gin.Context) {
	usersCount, modulesCount, lessonsCount, err := h.svc.GetStats(c.Request.Context())
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
	list, err := h.svc.GetActivity(c.Request.Context(), 20)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// parsePagination reads ?limit= and ?offset= query params with a hard cap on limit.
func parsePagination(c *gin.Context, maxLimit int) (limit, offset int) {
	limit = maxLimit
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		if v < maxLimit {
			limit = v
		}
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		offset = v
	}
	return
}
