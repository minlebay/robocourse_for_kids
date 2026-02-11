package progress

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
)

type LessonProgress struct {
	LessonID   uuid.UUID `json:"lesson_id"`
	LessonTitle string   `json:"lesson_title,omitempty"`
	ModuleID   uuid.UUID `json:"module_id,omitempty"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ChecklistProgressItem struct {
	ChecklistItemID uuid.UUID  `json:"checklist_item_id"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type UserProgress struct {
	Lessons    []LessonProgress       `json:"lessons"`
	Checklist  []ChecklistProgressItem `json:"checklist,omitempty"`
}

type SetProgressRequest struct {
	Status string `json:"status"` // not_started | in_progress | completed
}
