package lessons

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
	"learn_kids/backend/internal/sanitize"
)

// Length limits for validation (defence against huge payloads and DB bloat).
const (
	MaxTitleLength        = 500
	MaxDescriptionLength  = 10 * 1024   // 10 KB
	MaxStepContentLength  = 100 * 1024  // 100 KB per step
	MaxStepsPerLesson     = 200
)

var validLessonTypes = map[string]bool{"theory": true, "practice": true, "project": true}

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
	repo             Repository
	reactionProvider LessonReactionProvider // optional; nil = do not attach reaction counts
}

func NewHandler(repo Repository, reactionProvider LessonReactionProvider) *Handler {
	return &Handler{repo: repo, reactionProvider: reactionProvider}
}

func (h *Handler) ListModules(c *gin.Context) {
	var tag *string
	if t := c.Query("tag"); t != "" {
		tag = &t
	}
	list, err := h.repo.ListModules(c.Request.Context(), tag)
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
	mod, err := h.repo.GetModuleByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if mod == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		return
	}
	c.JSON(http.StatusOK, mod)
}

type CreateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

func (h *Handler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}
	if len(req.Title) > MaxTitleLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("title must be at most %d characters", MaxTitleLength)})
		return
	}
	if len(req.Description) > MaxDescriptionLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength)})
		return
	}
	title := req.Title
	description := sanitize.Description(req.Description)
	mod, err := h.repo.CreateModule(c.Request.Context(), title, description, req.SortOrder)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	deleted, err := h.repo.DeleteModule(c.Request.Context(), id)
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
	deleted, err := h.repo.DeleteLesson(c.Request.Context(), id)
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
	Title       string       `json:"title"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}
	if len(req.Title) > MaxTitleLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("title must be at most %d characters", MaxTitleLength)})
		return
	}
	if len(req.Description) > MaxDescriptionLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength)})
		return
	}
	if req.LessonType == "" {
		req.LessonType = "theory"
	}
	if !validLessonTypes[req.LessonType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lesson_type must be theory, practice, or project"})
		return
	}
	steps := make([]LessonStep, 0, len(req.Steps))
	for i, s := range req.Steps {
		if s.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "step " + strconv.Itoa(i+1) + " title must not be empty"})
			return
		}
		steps = append(steps, s)
	}
	if len(steps) > MaxStepsPerLesson {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("at most %d steps per lesson", MaxStepsPerLesson)})
		return
	}
	for i := range steps {
		if len(steps[i].Title) > MaxTitleLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("step %d title must be at most %d characters", i+1, MaxTitleLength)})
			return
		}
		if len(steps[i].Content) > MaxStepContentLength {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("step %d content must be at most %d characters", i+1, MaxStepContentLength)})
			return
		}
		steps[i].Title = sanitize.Description(steps[i].Title)
		steps[i].Content = sanitize.LessonContent(steps[i].Content)
	}
	title := req.Title
	description := sanitize.Description(req.Description)
	lesson, err := h.repo.CreateLesson(c.Request.Context(), moduleID, title, description, req.LessonType, req.SortOrder, steps)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	lesson, err := h.repo.GetLessonByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if lesson == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if h.reactionProvider != nil {
		userID := middleware.UserID(c)
		likes, dislikes, userReaction, err := h.reactionProvider.GetForLesson(c.Request.Context(), id, userID)
		if err != nil {
			httplog.LogError(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		lesson.LikesCount = likes
		lesson.DislikesCount = dislikes
		if userReaction != "" {
			lesson.UserReaction = &userReaction
		}
	}
	c.JSON(http.StatusOK, lesson)
}

// UpdateLessonRequest body for PUT /lessons/:id (teacher only).
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	if body.Title != nil && *body.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must not be empty"})
		return
	}
	if body.Title != nil && len(*body.Title) > MaxTitleLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("title must be at most %d characters", MaxTitleLength)})
		return
	}
	if body.Description != nil && len(*body.Description) > MaxDescriptionLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength)})
		return
	}
	var steps []LessonStep
	if body.Steps != nil {
		steps = *body.Steps
		if len(steps) > MaxStepsPerLesson {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("at most %d steps per lesson", MaxStepsPerLesson)})
			return
		}
		for i, s := range steps {
			if s.Title == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "step " + strconv.Itoa(i+1) + " title must not be empty"})
				return
			}
			if len(s.Title) > MaxTitleLength {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("step %d title must be at most %d characters", i+1, MaxTitleLength)})
				return
			}
			if len(s.Content) > MaxStepContentLength {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("step %d content must be at most %d characters", i+1, MaxStepContentLength)})
				return
			}
		}
		// Sanitize before passing to repo.
		for i := range steps {
			steps[i].Title = sanitize.Description(steps[i].Title)
			steps[i].Content = sanitize.LessonContent(steps[i].Content)
		}
	}
	var title, description *string
	if body.Title != nil {
		t := sanitize.Description(*body.Title)
		title = &t
	}
	if body.Description != nil {
		d := sanitize.Description(*body.Description)
		description = &d
	}
	lesson, err := h.repo.UpdateLesson(c.Request.Context(), id, title, description, steps)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, lesson)
}
