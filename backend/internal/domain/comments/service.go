package comments

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"learn_kids/backend/internal/sanitize"
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

// Service holds the business logic for the comments domain.
type Service struct {
	repo             Repository
	reactionProvider CommentReactionProvider // optional
}

func NewService(repo Repository, rp CommentReactionProvider) *Service {
	return &Service{repo: repo, reactionProvider: rp}
}

// ListComments returns comments for a lesson, with reaction counts injected when available.
func (s *Service) ListComments(ctx context.Context, lessonID, userID uuid.UUID) ([]Comment, error) {
	list, err := s.repo.ListByLesson(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []Comment{}
	}
	if s.reactionProvider != nil && len(list) > 0 {
		ids := make([]uuid.UUID, len(list))
		for i := range list {
			ids[i] = list[i].ID
		}
		counts, userReactions, err := s.reactionProvider.GetForComments(ctx, ids, userID)
		if err != nil {
			return nil, err
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
	return list, nil
}

// CreateComment sanitizes text, validates length, and creates the comment.
func (s *Service) CreateComment(ctx context.Context, lessonID, userID uuid.UUID, rawText string) (*Comment, error) {
	text := sanitize.Description(rawText)
	if len(text) == 0 {
		return nil, httpErr(http.StatusBadRequest, "text must be 1–2000 characters")
	}
	if len(text) > 2000 {
		return nil, httpErr(http.StatusBadRequest, "text must be at most 2000 characters")
	}
	return s.repo.Create(ctx, lessonID, userID, text)
}

// DeleteComment removes a comment by ID (only if it belongs to the user and lesson).
func (s *Service) DeleteComment(ctx context.Context, commentID, lessonID, userID uuid.UUID) (bool, error) {
	return s.repo.DeleteByIDAndUser(ctx, commentID, lessonID, userID)
}
