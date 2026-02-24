package progress

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	repo        Repository
	userChecker UserChecker // optional: when set, GetUserProgress returns 404 if user does not exist
}

func NewHandler(repo Repository, userChecker UserChecker) *Handler {
	return &Handler{repo: repo, userChecker: userChecker}
}

func (h *Handler) GetProgress(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	progress, err := h.repo.GetProgress(c.Request.Context(), userID)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status != StatusNotStarted && req.Status != StatusInProgress && req.Status != StatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be not_started, in_progress, or completed"})
		return
	}
	if err := h.repo.SetLessonProgress(c.Request.Context(), userID, lessonID, req.Status); err != nil {
		var pgErr *pgconn.PgError
		// FK violation (23503): lesson does not exist → return 404 instead of 500.
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			return
		}
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	ok, err := h.repo.ChecklistItemBelongsToLesson(c.Request.Context(), lessonID, itemID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checklist item does not belong to this lesson"})
		return
	}
	// Request body: { "completed": true/false }
	var req struct {
		Completed bool `json:"completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Completed = true
	}
	if err := h.repo.SetChecklistItemCompleted(c.Request.Context(), userID, itemID, req.Completed); err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetUserProgress returns progress for user :id (teacher only - middleware.RequireTeacher must be applied by router).
// Returns 404 if userChecker is set and the user does not exist.
func (h *Handler) GetUserProgress(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if h.userChecker != nil {
		exists, err := h.userChecker.UserExists(c.Request.Context(), targetID)
		if err != nil {
			httplog.LogError(c, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
	}
	progress, err := h.repo.GetProgress(c.Request.Context(), targetID)
	if err != nil {
		httplog.LogError(c, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, progress)
}
