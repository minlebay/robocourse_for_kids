package lessons

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

// Length limits for validation (defence against huge payloads and DB bloat).
const (
	MaxTitleLength       = 500
	MaxDescriptionLength = 10 * 1024  // 10 KB
	MaxStepContentLength = 100 * 1024 // 100 KB per step
	MaxStepsPerLesson    = 200
	MaxTagLength         = 100
)

// Repository defines the data access interface for lessons and modules.
type Repository interface {
	ListModules(ctx context.Context, tag *string) ([]Module, error)
	GetModuleByID(ctx context.Context, id uuid.UUID) (*Module, error)
	CreateModule(ctx context.Context, title, description string, sortOrder int) (*Module, error)
	DeleteModule(ctx context.Context, id uuid.UUID) (bool, error)
	GetLessonByID(ctx context.Context, id uuid.UUID) (*Lesson, error)
	CreateLesson(ctx context.Context, moduleID uuid.UUID, title, description, lessonType string, sortOrder int, steps []LessonStep) (*Lesson, error)
	DeleteLesson(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateLesson(ctx context.Context, id uuid.UUID, title, description *string, steps []LessonStep) (*Lesson, error)
}

// LessonReactionProvider optionally provides reaction counts for a lesson (implemented by reactions domain or adapter).
type LessonReactionProvider interface {
	GetForLesson(ctx context.Context, lessonID, userID uuid.UUID) (likes, dislikes int, userReaction string, err error)
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// handleErr maps service HTTPError to an HTTP response. Returns true if the error was handled.
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

func (h *Handler) ListModules(c *gin.Context) {
	var tag *string
	if t := c.Query("tag"); t != "" {
		if len(t) > MaxTagLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tag must be at most %d characters", MaxTagLength)})
			return
		}
		tag = &t
	}
	list, err := h.svc.ListModules(c.Request.Context(), tag)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) GetModule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module id"})
		return
	}
	mod, err := h.svc.GetModuleByID(c.Request.Context(), id)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, mod)
}

type CreateModuleRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

func (h *Handler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	mod, err := h.svc.CreateModule(c.Request.Context(), req.Title, req.Description, req.SortOrder)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, mod)
}

func (h *Handler) DeleteModule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module id"})
		return
	}
	deleted, err := h.svc.DeleteModule(c.Request.Context(), id)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteLesson(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	deleted, err := h.svc.DeleteLesson(c.Request.Context(), id)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

type CreateLessonRequest struct {
	Title       string       `json:"title" binding:"required"`
	Description string       `json:"description"`
	LessonType  string       `json:"lesson_type"`
	SortOrder   int          `json:"sort_order"`
	Steps       []LessonStep `json:"steps"`
}

func (h *Handler) CreateLesson(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module id"})
		return
	}
	var req CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	lesson, err := h.svc.CreateLesson(c.Request.Context(), moduleID, req.Title, req.Description, req.LessonType, req.SortOrder, req.Steps)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, lesson)
}

func (h *Handler) GetLesson(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	userID := middleware.UserID(c)
	lesson, err := h.svc.GetLesson(c.Request.Context(), id, userID)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, lesson)
}

// UpdateLessonRequest body for PUT /lessons/:id (teacher only).
// Steps: if nil — do not touch steps; if empty array — delete all steps.
type UpdateLessonRequest struct {
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Steps       *[]LessonStep `json:"steps"`
}

func (h *Handler) UpdateLesson(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	var body UpdateLessonRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	var rawSteps []LessonStep
	if body.Steps != nil {
		rawSteps = *body.Steps
	} else {
		rawSteps = nil
	}
	lesson, err := h.svc.UpdateLesson(c.Request.Context(), id, body.Title, body.Description, rawSteps)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, lesson)
}
