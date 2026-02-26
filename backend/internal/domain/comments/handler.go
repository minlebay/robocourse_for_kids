package comments

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
	"learn_kids/backend/internal/sanitize"
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
	repo             Repository
	reactionProvider CommentReactionProvider // optional
}

func NewHandler(repo Repository, reactionProvider CommentReactionProvider) *Handler {
	return &Handler{repo: repo, reactionProvider: reactionProvider}
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
	if h.reactionProvider != nil && len(list) > 0 {
		ids := make([]uuid.UUID, len(list))
		for i := range list {
			ids[i] = list[i].ID
		}
		userID := middleware.UserID(c)
		counts, userReactions, err := h.reactionProvider.GetForComments(c.Request.Context(), ids, userID)
		if err != nil {
			httplog.LogError(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		for i := range list {
			if c, ok := counts[list[i].ID]; ok {
				list[i].LikesCount = c.Likes
				list[i].DislikesCount = c.Dislikes
			}
			if r, ok := userReactions[list[i].ID]; ok && r != "" {
				list[i].UserReaction = &r
			}
		}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// Sanitize to prevent XSS when comment is rendered (plain text or HTML).
	text := sanitize.Description(req.Text)
	if len(text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text must be 1–2000 characters"})
		return
	}
	if len(text) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text must be at most 2000 characters"})
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
