package reactions

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

// CommentLessonChecker returns the lesson ID for a comment. Used to verify comment belongs to lesson in path.
// Optional: if nil, the check is skipped (backward compatibility).
type CommentLessonChecker interface {
	GetCommentLessonID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type SetReactionRequest struct {
	Reaction string `json:"reaction"`
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
	if handleErr(c, h.svc.SetLessonReaction(c.Request.Context(), lessonID, userID, req.Reaction)) {
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
	deleted, err := h.svc.DeleteLessonReaction(c.Request.Context(), lessonID, userID)
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
	var req SetReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if handleErr(c, h.svc.SetCommentReaction(c.Request.Context(), commentID, lessonID, userID, req.Reaction)) {
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
	deleted, err := h.svc.DeleteCommentReaction(c.Request.Context(), commentID, lessonID, userID)
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
