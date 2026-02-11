package lessons

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/modules", h.ListModules)
	r.GET("/modules/:id", h.GetModule)
	r.GET("/lessons/:id", h.GetLesson)
}

func (h *Handler) ListModules(c *gin.Context) {
	var platform, tag *string
	if p := c.Query("platform"); p != "" {
		platform = &p
	}
	if t := c.Query("tag"); t != "" {
		tag = &t
	}
	list, err := h.repo.ListModules(c.Request.Context(), platform, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	mod, err := h.repo.CreateModule(c.Request.Context(), req.Title, req.Description, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

type CreateLessonRequest struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	LessonType  string        `json:"lesson_type"`
	SortOrder   int           `json:"sort_order"`
	Steps       []LessonStep  `json:"steps"`
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
	if req.LessonType == "" {
		req.LessonType = "theory"
	}
	var steps []LessonStep
	for _, s := range req.Steps {
		if s.Title != "" {
			steps = append(steps, s)
		}
	}
	lesson, err := h.repo.CreateLesson(c.Request.Context(), moduleID, req.Title, req.Description, req.LessonType, req.SortOrder, steps)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if lesson == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
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
	var steps []LessonStep
	if body.Steps != nil {
		steps = *body.Steps
		for i, s := range steps {
			if s.Title == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "step " + strconv.Itoa(i+1) + " title must not be empty"})
				return
			}
		}
	}
	lesson, err := h.repo.UpdateLesson(c.Request.Context(), id, body.Title, body.Description, steps)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lesson)
}
