package reactions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
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
	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}
	var req SetReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
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
	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
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
