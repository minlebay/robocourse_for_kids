package reactions

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// Service holds the business logic for the reactions domain.
type Service struct {
	repo           Repository
	commentChecker CommentLessonChecker // optional
}

func NewService(repo Repository, cc CommentLessonChecker) *Service {
	return &Service{repo: repo, commentChecker: cc}
}

// validateReaction returns an error if the reaction string is not "like" or "dislike".
func (s *Service) validateReaction(reaction string) error {
	if !validReaction[reaction] {
		return httpErr(http.StatusBadRequest, "reaction must be 'like' or 'dislike'")
	}
	return nil
}

// validateCommentLesson verifies the comment belongs to the given lesson.
// Returns HTTPError 404 if the comment is not found or belongs to a different lesson.
func (s *Service) validateCommentLesson(ctx context.Context, commentID, lessonID uuid.UUID) error {
	if s.commentChecker == nil {
		return nil
	}
	commentLessonID, err := s.commentChecker.GetCommentLessonID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpErr(http.StatusNotFound, "comment not found")
		}
		return err
	}
	if commentLessonID != lessonID {
		return httpErr(http.StatusNotFound, "comment not found")
	}
	return nil
}

// SetLessonReaction sets or updates a user's reaction for a lesson.
func (s *Service) SetLessonReaction(ctx context.Context, lessonID, userID uuid.UUID, reaction string) error {
	if err := s.validateReaction(reaction); err != nil {
		return err
	}
	return s.repo.SetLessonReaction(ctx, lessonID, userID, reaction)
}

// DeleteLessonReaction removes a user's reaction for a lesson.
func (s *Service) DeleteLessonReaction(ctx context.Context, lessonID, userID uuid.UUID) (bool, error) {
	return s.repo.DeleteLessonReaction(ctx, lessonID, userID)
}

// SetCommentReaction validates the comment belongs to the lesson, then sets the reaction.
func (s *Service) SetCommentReaction(ctx context.Context, commentID, lessonID, userID uuid.UUID, reaction string) error {
	if err := s.validateCommentLesson(ctx, commentID, lessonID); err != nil {
		return err
	}
	if err := s.validateReaction(reaction); err != nil {
		return err
	}
	return s.repo.SetCommentReaction(ctx, commentID, userID, reaction)
}

// DeleteCommentReaction validates the comment belongs to the lesson, then removes the reaction.
func (s *Service) DeleteCommentReaction(ctx context.Context, commentID, lessonID, userID uuid.UUID) (bool, error) {
	if err := s.validateCommentLesson(ctx, commentID, lessonID); err != nil {
		return false, err
	}
	return s.repo.DeleteCommentReaction(ctx, commentID, userID)
}
