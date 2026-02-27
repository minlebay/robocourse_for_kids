package progress

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"learn_kids/backend/internal/httplog"
	"learn_kids/backend/internal/middleware"
)

// Repository defines the data access interface for progress tracking.
type Repository interface {
	GetProgress(ctx context.Context, userID uuid.UUID) (*UserProgress, error)
	SetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, status string) error
	SetChecklistItemCompleted(ctx context.Context, userID, checklistItemID uuid.UUID, completed bool) error
	ChecklistItemBelongsToLesson(ctx context.Context, lessonID, itemID uuid.UUID) (bool, error)
}

// UserChecker optionally checks if a user exists. Used by GetUserProgress to return 404 for non-existent users.
// If nil, the check is skipped (legacy behaviour: 200 with empty progress).
type UserChecker interface {
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
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

func (h *Handler) GetProgress(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	progress, err := h.svc.GetProgress(c.Request.Context(), userID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *Handler) SetLessonProgress(c *gin.Context) {
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
	var req SetProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if handleErr(c, h.svc.SetLessonProgress(c.Request.Context(), userID, lessonID, req.Status)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) SetChecklistItem(c *gin.Context) {
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
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid checklist item id"})
		return
	}
	var req struct {
		Completed bool `json:"completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Completed = true
	}
	if handleErr(c, h.svc.SetChecklistItem(c.Request.Context(), userID, lessonID, itemID, req.Completed)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetUserProgress returns progress for user :id (teacher only - middleware.RequireTeacher must be applied by router).
func (h *Handler) GetUserProgress(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	progress, err := h.svc.GetUserProgress(c.Request.Context(), targetID)
	if handleErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, progress)
}
