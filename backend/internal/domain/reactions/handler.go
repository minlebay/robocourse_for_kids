package reactions

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

// CommentLessonChecker returns the lesson ID for a comment. Used to verify comment belongs to lesson in path.
// Optional: if nil, the check is skipped (backward compatibility).
type CommentLessonChecker interface {
	GetCommentLessonID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
}

type Handler struct {
	repo            Repository
	commentChecker  CommentLessonChecker // optional: when set, comment reactions verify lesson from path
}

func NewHandler(repo Repository, commentChecker CommentLessonChecker) *Handler {
	return &Handler{repo: repo, commentChecker: commentChecker}
}

type SetReactionRequest struct {
	Reaction string `json:"reaction"`
}

// SetLessonReaction sets or updates the current user's reaction for a lesson. JWT required.
func (h *Handler) SetLessonReaction(c *gin.Context) {
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
	var req SetReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if !validReaction[req.Reaction] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reaction must be 'like' or 'dislike'"})
		return
	}
	if err := h.repo.SetLessonReaction(c.Request.Context(), lessonID, userID, req.Reaction); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteLessonReaction removes the current user's reaction for a lesson. JWT required.
func (h *Handler) DeleteLessonReaction(c *gin.Context) {
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
	deleted, err := h.repo.DeleteLessonReaction(c.Request.Context(), lessonID, userID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "reaction not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// SetCommentReaction sets or updates the current user's reaction for a comment. JWT required.
func (h *Handler) SetCommentReaction(c *gin.Context) {
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
	if h.commentChecker != nil {
		commentLessonID, err := h.commentChecker.GetCommentLessonID(c.Request.Context(), commentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
				return
			}
			httplog.LogError(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if commentLessonID != lessonID {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
	}
	var req SetReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if !validReaction[req.Reaction] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reaction must be 'like' or 'dislike'"})
		return
	}
	if err := h.repo.SetCommentReaction(c.Request.Context(), commentID, userID, req.Reaction); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteCommentReaction removes the current user's reaction for a comment. JWT required.
func (h *Handler) DeleteCommentReaction(c *gin.Context) {
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
	if h.commentChecker != nil {
		commentLessonID, err := h.commentChecker.GetCommentLessonID(c.Request.Context(), commentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
				return
			}
			httplog.LogError(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if commentLessonID != lessonID {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
	}
	deleted, err := h.repo.DeleteCommentReaction(c.Request.Context(), commentID, userID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "reaction not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
