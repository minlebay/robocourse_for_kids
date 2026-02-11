package users

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo   *Repo
	jwtKey []byte
}

func NewHandler(repo *Repo, jwtSecret string) *Handler {
	return &Handler{repo: repo, jwtKey: []byte(jwtSecret)}
}

var validThemes = map[string]bool{"default": true, "light": true, "cyberpunk": true}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("/auth/register", h.RegisterUser)
	r.POST("/auth/login", h.Login)
	r.GET("/auth/me", h.Me)
	r.PATCH("/auth/me", h.UpdateMe)
	r.GET("/users", h.RequireTeacher, h.ListUsers)
}

func (h *Handler) RequireTeacher(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists || uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid.(uuid.UUID))
	if err != nil || u == nil || u.Role != RoleTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: teacher role required"})
		c.Abort()
		return
	}
	c.Next()
}

type RegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = RoleStudent
	}
	if req.Role != RoleStudent && req.Role != RoleTeacher {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be student or teacher"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := h.generateToken(user.ID, user.Role)
	if err != nil {
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
	if err != nil || u == nil {
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}
	user := &User{ID: u.ID, Login: u.Login, Name: u.Name, Role: u.Role, Theme: u.Theme, CreatedAt: u.CreatedAt}
	token, err := h.generateToken(user.ID, user.Role)
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "theme must be one of: default, light, cyberpunk"})
		return
	}
	if err := h.repo.UpdateTheme(c.Request.Context(), uid.(uuid.UUID), req.Theme); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u, err := h.repo.GetByID(c.Request.Context(), uid.(uuid.UUID))
	if err != nil || u == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found after update"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) ListUsers(c *gin.Context) {
	list, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

