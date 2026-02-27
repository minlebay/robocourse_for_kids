package comments

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

// CommentReactionProvider optionally provides reaction counts for comments (implemented by reactions domain or adapter).
type CommentReactionProvider interface {
	GetForComments(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) (
		counts map[uuid.UUID]struct{ Likes, Dislikes int },
		userReactions map[uuid.UUID]string,
		err error,
	)
}

// Repository defines the data access interface for comments.
type Repository interface {
	ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]Comment, error)
	Create(ctx context.Context, lessonID, userID uuid.UUID, text string) (*Comment, error)
	DeleteByIDAndUser(ctx context.Context, commentID, lessonID, userID uuid.UUID) (bool, error)
	GetCommentLessonID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
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

type CreateCommentRequest struct {
	Text string `json:"text"`
}

func (h *Handler) List(c *gin.Context) {
	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}
	userID := middleware.UserID(c)
	list, err := h.svc.ListComments(c.Request.Context(), lessonID, userID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, list)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	comment, err := h.svc.CreateComment(c.Request.Context(), lessonID, userID, req.Text)
	if handleErr(c, err) {
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
	deleted, err := h.svc.DeleteComment(c.Request.Context(), commentID, lessonID, userID)
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
