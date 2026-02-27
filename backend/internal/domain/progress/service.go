package progress

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// HTTPError carries an HTTP status code and client-facing message from the service layer.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

func httpErr(code int, msg string) *HTTPError {
	return &HTTPError{Code: code, Message: msg}
}

// Service holds the business logic for the progress domain.
type Service struct {
	repo        Repository
	userChecker UserChecker // optional
}

func NewService(repo Repository, userChecker UserChecker) *Service {
	return &Service{repo: repo, userChecker: userChecker}
}

// GetProgress returns the current user's progress.
func (s *Service) GetProgress(ctx context.Context, userID uuid.UUID) (*UserProgress, error) {
	return s.repo.GetProgress(ctx, userID)
}

// SetLessonProgress validates the status and records lesson progress for a user.
func (s *Service) SetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, status string) error {
	if status != StatusNotStarted && status != StatusInProgress && status != StatusCompleted {
		return httpErr(http.StatusBadRequest, "status must be not_started, in_progress, or completed")
	}
	if err := s.repo.SetLessonProgress(ctx, userID, lessonID, status); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return httpErr(http.StatusNotFound, "lesson not found")
		}
		return err
	}
	return nil
}

// SetChecklistItem verifies the item belongs to the lesson and records completion.
func (s *Service) SetChecklistItem(ctx context.Context, userID, lessonID, itemID uuid.UUID, completed bool) error {
	ok, err := s.repo.ChecklistItemBelongsToLesson(ctx, lessonID, itemID)
	if err != nil {
		return err
	}
	if !ok {
		return httpErr(http.StatusBadRequest, "checklist item does not belong to this lesson")
	}
	return s.repo.SetChecklistItemCompleted(ctx, userID, itemID, completed)
}

// GetUserProgress returns progress for another user. Returns 404 if userChecker is set and the user does not exist.
func (s *Service) GetUserProgress(ctx context.Context, targetID uuid.UUID) (*UserProgress, error) {
	if s.userChecker != nil {
		exists, err := s.userChecker.UserExists(ctx, targetID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, httpErr(http.StatusNotFound, "user not found")
		}
	}
	return s.repo.GetProgress(ctx, targetID)
}
