package lessons

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
