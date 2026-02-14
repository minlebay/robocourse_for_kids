package progress

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChecklistItemBelongsToLesson returns true if the checklist item belongs to the given lesson.
func (r *Repo) ChecklistItemBelongsToLesson(ctx context.Context, lessonID, itemID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM checklist_items WHERE id = $1 AND lesson_id = $2)`,
		itemID, lessonID).Scan(&exists)
	return exists, err
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) GetProgress(ctx context.Context, userID uuid.UUID) (*UserProgress, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ulp.lesson_id, l.title, l.module_id, ulp.status, ulp.updated_at
		FROM user_lesson_progress ulp
		JOIN lessons l ON l.id = ulp.lesson_id
		WHERE ulp.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lessons []LessonProgress
	for rows.Next() {
		var lp LessonProgress
		if err := rows.Scan(&lp.LessonID, &lp.LessonTitle, &lp.ModuleID, &lp.Status, &lp.UpdatedAt); err != nil {
			return nil, err
		}
		lessons = append(lessons, lp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows2, err := r.pool.Query(ctx, `
		SELECT checklist_item_id, completed_at
		FROM user_checklist_progress
		WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	var checklist []ChecklistProgressItem
	for rows2.Next() {
		var c ChecklistProgressItem
		if err := rows2.Scan(&c.ChecklistItemID, &c.CompletedAt); err != nil {
			return nil, err
		}
		checklist = append(checklist, c)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	if lessons == nil {
		lessons = []LessonProgress{}
	}
	if checklist == nil {
		checklist = []ChecklistProgressItem{}
	}
	return &UserProgress{Lessons: lessons, Checklist: checklist}, nil
}

func (r *Repo) SetLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_lesson_progress (user_id, lesson_id, status, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, lesson_id) DO UPDATE SET status = $3, updated_at = now()`,
		userID, lessonID, status)
	return err
}

func (r *Repo) SetChecklistItemCompleted(ctx context.Context, userID, checklistItemID uuid.UUID, completed bool) error {
	if completed {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO user_checklist_progress (user_id, checklist_item_id, completed_at)
			VALUES ($1, $2, now())
			ON CONFLICT (user_id, checklist_item_id) DO NOTHING`,
			userID, checklistItemID)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_checklist_progress
		WHERE user_id = $1 AND checklist_item_id = $2`,
		userID, checklistItemID)
	return err
}
