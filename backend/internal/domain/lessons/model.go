package lessons

import (
	"time"

	"github.com/google/uuid"
)

type Module struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	SortOrder   int        `json:"sort_order"`
	Lessons     []Lesson   `json:"lessons,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Lesson struct {
	ID           uuid.UUID        `json:"id"`
	ModuleID     uuid.UUID        `json:"module_id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	LessonType   string           `json:"lesson_type"` // theory, practice, project
	SortOrder    int              `json:"sort_order"`
	Steps        []LessonStep     `json:"steps,omitempty"`
	Materials    []LessonMaterial `json:"materials,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
	Checklist    []ChecklistItem  `json:"checklist,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	LikesCount   int              `json:"likes_count,omitempty"`
	DislikesCount int             `json:"dislikes_count,omitempty"`
	UserReaction *string          `json:"user_reaction,omitempty"` // "like" or "dislike" if authenticated
}

type LessonStep struct {
	ID        uuid.UUID `json:"id"`
	LessonID  uuid.UUID `json:"lesson_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
}

type LessonMaterial struct {
	ID         uuid.UUID `json:"id"`
	LessonID   uuid.UUID `json:"lesson_id"`
	Kind       string    `json:"kind"` // link, file
	URLOrPath  string    `json:"url_or_path"`
	Title      string    `json:"title,omitempty"`
}

type ChecklistItem struct {
	ID        uuid.UUID `json:"id"`
	LessonID  uuid.UUID `json:"lesson_id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
}
