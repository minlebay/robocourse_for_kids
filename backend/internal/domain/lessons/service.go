package lessons

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// FreeLessonsCount is the number of lessons (by sort_order) accessible without login.
const FreeLessonsCount = 3

var validLessonTypes = map[string]bool{"theory": true, "practice": true, "project": true}

// Service holds the business logic for the lessons domain.
type Service struct {
	repo             Repository
	reactionProvider LessonReactionProvider // optional
}

func NewService(repo Repository, reactionProvider LessonReactionProvider) *Service {
	return &Service{repo: repo, reactionProvider: reactionProvider}
}

// ValidateAndSanitizeSteps validates and sanitizes a slice of lesson steps.
func (s *Service) ValidateAndSanitizeSteps(steps []LessonStep) ([]LessonStep, error) {
	if len(steps) > MaxStepsPerLesson {
		return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("at most %d steps per lesson", MaxStepsPerLesson))
	}
	for i, step := range steps {
		if step.Title == "" {
			return nil, httpErr(http.StatusBadRequest, "step "+strconv.Itoa(i+1)+" title must not be empty")
		}
		if len(step.Title) > MaxTitleLength {
			return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("step %d title must be at most %d characters", i+1, MaxTitleLength))
		}
		if len(step.Content) > MaxStepContentLength {
			return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("step %d content must be at most %d characters", i+1, MaxStepContentLength))
		}
		steps[i].Title = sanitize.Description(step.Title)
		steps[i].Content = sanitize.LessonContent(step.Content)
	}
	return steps, nil
}

// GetLesson loads a lesson, enforces freemium access, and optionally injects reaction counts.
func (s *Service) GetLesson(ctx context.Context, id, userID uuid.UUID) (*Lesson, error) {
	lesson, err := s.repo.GetLessonByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, httpErr(http.StatusNotFound, "lesson not found")
		}
		return nil, err
	}
	// Free preview: only the first FreeLessonsCount lessons are accessible without auth.
	if userID == uuid.Nil && lesson.SortOrder >= FreeLessonsCount {
		return nil, httpErr(http.StatusForbidden, "auth_required")
	}
	if s.reactionProvider != nil {
		likes, dislikes, userReaction, err := s.reactionProvider.GetForLesson(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		lesson.LikesCount = likes
		lesson.DislikesCount = dislikes
		if userReaction != "" {
			lesson.UserReaction = &userReaction
		}
	}
	return lesson, nil
}

// CreateLesson validates and creates a lesson with steps.
func (s *Service) CreateLesson(ctx context.Context, moduleID uuid.UUID, title, description, lessonType string, sortOrder int, rawSteps []LessonStep) (*Lesson, error) {
	if len(title) > MaxTitleLength {
		return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("title must be at most %d characters", MaxTitleLength))
	}
	if len(description) > MaxDescriptionLength {
		return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength))
	}
	if lessonType == "" {
		lessonType = "theory"
	}
	if !validLessonTypes[lessonType] {
		return nil, httpErr(http.StatusBadRequest, "lesson_type must be theory, practice, or project")
	}
	steps, err := s.ValidateAndSanitizeSteps(rawSteps)
	if err != nil {
		return nil, err
	}
	title = sanitize.Description(title)
	description = sanitize.Description(description)
	lesson, err := s.repo.CreateLesson(ctx, moduleID, title, description, lessonType, sortOrder, steps)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, httpErr(http.StatusNotFound, "module not found")
		}
		return nil, err
	}
	return lesson, nil
}

// UpdateLesson validates and updates lesson fields. nil steps means "do not update steps".
func (s *Service) UpdateLesson(ctx context.Context, id uuid.UUID, title, description *string, rawSteps []LessonStep) (*Lesson, error) {
	if title != nil {
		if *title == "" {
			return nil, httpErr(http.StatusBadRequest, "title must not be empty")
		}
		if len(*title) > MaxTitleLength {
			return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("title must be at most %d characters", MaxTitleLength))
		}
		t := sanitize.Description(*title)
		title = &t
	}
	if description != nil {
		if len(*description) > MaxDescriptionLength {
			return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength))
		}
		d := sanitize.Description(*description)
		description = &d
	}
	var steps []LessonStep // nil = do not update
	if rawSteps != nil {
		var err error
		steps, err = s.ValidateAndSanitizeSteps(rawSteps)
		if err != nil {
			return nil, err
		}
	}
	lesson, err := s.repo.UpdateLesson(ctx, id, title, description, steps)
	if err != nil {
		if isNotFound(err) {
			return nil, httpErr(http.StatusNotFound, "lesson not found")
		}
		return nil, err
	}
	return lesson, nil
}

// CreateModule validates, sanitizes, and creates a module.
func (s *Service) CreateModule(ctx context.Context, title, description string, sortOrder int) (*Module, error) {
	if len(title) > MaxTitleLength {
		return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("title must be at most %d characters", MaxTitleLength))
	}
	if len(description) > MaxDescriptionLength {
		return nil, httpErr(http.StatusBadRequest, fmt.Sprintf("description must be at most %d characters", MaxDescriptionLength))
	}
	return s.repo.CreateModule(ctx, sanitize.Description(title), sanitize.Description(description), sortOrder)
}

// ListModules lists modules, with optional tag filter.
func (s *Service) ListModules(ctx context.Context, tag *string) ([]Module, error) {
	return s.repo.ListModules(ctx, tag)
}

// GetModuleByID returns a module with its lessons.
func (s *Service) GetModuleByID(ctx context.Context, id uuid.UUID) (*Module, error) {
	mod, err := s.repo.GetModuleByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, httpErr(http.StatusNotFound, "module not found")
		}
		return nil, err
	}
	return mod, nil
}

// DeleteModule removes a module by ID.
func (s *Service) DeleteModule(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.DeleteModule(ctx, id)
}

// DeleteLesson removes a lesson by ID.
func (s *Service) DeleteLesson(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.DeleteLesson(ctx, id)
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
