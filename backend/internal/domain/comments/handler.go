package comments

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

// Repository defines the data access interface for comments.
type Repository interface {
	ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]Comment, error)
	Create(ctx context.Context, lessonID, userID uuid.UUID, text string) (*Comment, error)
	DeleteByIDAndUser(ctx context.Context, commentID, lessonID, userID uuid.UUID) (bool, error)
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(c *gin.Context) {
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	list, err := h.repo.ListByLesson(c.Request.Context(), lessonID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if list == nil {
		list = []Comment{}
	}
	c.JSON(http.StatusOK, list)
}

type CreateCommentRequest struct {
	Text string `json:"text"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	text := req.Text
	if len(text) == 0 || len(text) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text must be 1–2000 characters"})
		return
	}
	comment, err := h.repo.Create(c.Request.Context(), lessonID, userID, text)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}
	deleted, err := h.repo.DeleteByIDAndUser(c.Request.Context(), commentID, lessonID, userID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found or access denied"})
		return
	}
	c.Status(http.StatusNoContent)
}
