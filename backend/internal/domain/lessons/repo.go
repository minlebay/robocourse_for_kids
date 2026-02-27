package lessons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time check: *Repo implements Repository.
var _ Repository = (*Repo)(nil)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) ListModules(ctx context.Context, tag *string, ownerID *uuid.UUID) ([]Module, error) {
	var rows pgx.Rows
	var err error

	if ownerID != nil {
		// Only modules owned by this user (?mine=true)
		if tag != nil {
			rows, err = r.pool.Query(ctx, `
				SELECT DISTINCT m.id, m.title, m.description, m.sort_order, m.created_at, m.owner_id
				FROM modules m
				JOIN lessons l ON l.module_id = m.id
				JOIN lesson_tags lt ON lt.lesson_id = l.id
				WHERE m.owner_id = $1 AND lt.tag = $2
				ORDER BY m.sort_order, m.created_at`, *ownerID, *tag)
		} else {
			rows, err = r.pool.Query(ctx, `
				SELECT id, title, description, sort_order, created_at, owner_id
				FROM modules
				WHERE owner_id = $1
				ORDER BY sort_order, created_at`, *ownerID)
		}
	} else if tag != nil {
		rows, err = r.pool.Query(ctx, `
			SELECT DISTINCT m.id, m.title, m.description, m.sort_order, m.created_at, m.owner_id
			FROM modules m
			JOIN lessons l ON l.module_id = m.id
			JOIN lesson_tags lt ON lt.lesson_id = l.id
			WHERE lt.tag = $1
			ORDER BY m.sort_order, m.created_at`, *tag)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, title, description, sort_order, created_at, owner_id
			FROM modules
			ORDER BY sort_order, created_at`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Module
	for rows.Next() {
		var m Module
		var desc *string
		if err := rows.Scan(&m.ID, &m.Title, &desc, &m.SortOrder, &m.CreatedAt, &m.OwnerID); err != nil {
			return nil, err
		}
		if desc != nil {
			m.Description = *desc
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if list == nil {
		list = []Module{}
	}
	return list, nil
}

func (r *Repo) CreateModule(ctx context.Context, title, description string, sortOrder int, ownerID *uuid.UUID) (*Module, error) {
	var m Module
	var desc *string
	if description != "" {
		desc = &description
	}
	var outDesc *string
	var outOwnerID *uuid.UUID
	// When sortOrder is 0, auto-assign the next value using a CTE — no transaction needed.
	err := r.pool.QueryRow(ctx, `
		WITH next_order AS (
			SELECT CASE WHEN $3 = 0
				THEN COALESCE((SELECT MAX(sort_order) FROM modules), 0) + 1
				ELSE $3
			END AS val
		)
		INSERT INTO modules (title, description, sort_order, owner_id)
		SELECT $1, $2, val, $4 FROM next_order
		RETURNING id, title, description, sort_order, created_at, owner_id
	`, title, desc, sortOrder, ownerID).Scan(&m.ID, &m.Title, &outDesc, &m.SortOrder, &m.CreatedAt, &outOwnerID)
	if err != nil {
		return nil, err
	}
	if outDesc != nil {
		m.Description = *outDesc
	}
	m.OwnerID = outOwnerID
	return &m, nil
}

func (r *Repo) DeleteModule(ctx context.Context, id uuid.UUID) (deleted bool, err error) {
	cmd, err := r.pool.Exec(ctx, "DELETE FROM modules WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *Repo) DeleteLesson(ctx context.Context, id uuid.UUID) (deleted bool, err error) {
	cmd, err := r.pool.Exec(ctx, "DELETE FROM lessons WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *Repo) CreateLesson(ctx context.Context, moduleID uuid.UUID, title, description, lessonType string, sortOrder int, steps []LessonStep) (*Lesson, error) {
	if title == "" {
		return nil, errors.New("title must not be empty")
	}
	if lessonType == "" {
		lessonType = "theory"
	}
	if lessonType != "theory" && lessonType != "practice" && lessonType != "project" {
		return nil, errors.New("lesson_type must be theory, practice, or project")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if sortOrder == 0 {
		var maxOrder int
		_ = tx.QueryRow(ctx, "SELECT COALESCE(MAX(sort_order), 0) + 1 FROM lessons WHERE module_id = $1", moduleID).Scan(&maxOrder)
		sortOrder = maxOrder
	}
	var l Lesson
	var desc *string
	if description != "" {
		desc = &description
	}
	var outDesc *string
	err = tx.QueryRow(ctx, `
		INSERT INTO lessons (module_id, title, description, lesson_type, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, module_id, title, description, lesson_type, sort_order, created_at
	`, moduleID, title, desc, lessonType, sortOrder).Scan(&l.ID, &l.ModuleID, &l.Title, &outDesc, &l.LessonType, &l.SortOrder, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	if outDesc != nil {
		l.Description = *outDesc
	}
	if steps != nil {
		for i, s := range steps {
			if s.Title == "" {
				continue
			}
			so := i
			if s.SortOrder > 0 {
				so = s.SortOrder
			}
			_, err = tx.Exec(ctx,
				"INSERT INTO lesson_steps (lesson_id, title, content, sort_order) VALUES ($1, $2, $3, $4)",
				l.ID, s.Title, nullIfEmpty(s.Content), so)
			if err != nil {
				return nil, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	if steps != nil {
		l.Steps, _ = r.getStepsByLessonID(ctx, l.ID)
	}
	return &l, nil
}

func (r *Repo) GetModuleByID(ctx context.Context, id uuid.UUID) (*Module, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, description, sort_order, created_at, owner_id
		FROM modules WHERE id = $1`, id)
	var m Module
	var modDesc *string
	err := row.Scan(&m.ID, &m.Title, &modDesc, &m.SortOrder, &m.CreatedAt, &m.OwnerID)
	if err != nil {
		return nil, err
	}
	if modDesc != nil {
		m.Description = *modDesc
	}
	lessons, err := r.listLessonsByModuleID(ctx, id)
	if err != nil {
		return nil, err
	}
	m.Lessons = lessons
	return &m, nil
}

func (r *Repo) listLessonsByModuleID(ctx context.Context, moduleID uuid.UUID) ([]Lesson, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, module_id, title, description, lesson_type, sort_order, created_at
		FROM lessons WHERE module_id = $1 ORDER BY sort_order, created_at`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Lesson{}
	for rows.Next() {
		var l Lesson
		var desc *string
		err := rows.Scan(&l.ID, &l.ModuleID, &l.Title, &desc, &l.LessonType, &l.SortOrder, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		if desc != nil {
			l.Description = *desc
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

// GetLessonByID loads a lesson together with its steps, materials, tags and
// checklist in a single database round-trip using correlated subqueries.
func (r *Repo) GetLessonByID(ctx context.Context, id uuid.UUID) (*Lesson, error) {
	const q = `
		SELECT
			l.id, l.module_id, l.title, l.description, l.lesson_type, l.sort_order, l.created_at,
			(SELECT COALESCE(json_agg(s ORDER BY s.sort_order), '[]')
			 FROM (SELECT id, lesson_id, title, content, sort_order
			       FROM lesson_steps WHERE lesson_id = l.id) s),
			(SELECT COALESCE(json_agg(m), '[]')
			 FROM (SELECT id, lesson_id, kind, url_or_path, title
			       FROM lesson_materials WHERE lesson_id = l.id) m),
			(SELECT COALESCE(json_agg(t.tag ORDER BY t.tag), '[]')
			 FROM lesson_tags t WHERE t.lesson_id = l.id),
			(SELECT COALESCE(json_agg(c ORDER BY c.sort_order), '[]')
			 FROM (SELECT id, lesson_id, title, sort_order
			       FROM checklist_items WHERE lesson_id = l.id) c)
		FROM lessons l
		WHERE l.id = $1`

	var l Lesson
	var desc *string
	var stepsJSON, materialsJSON, tagsJSON, checklistJSON []byte

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&l.ID, &l.ModuleID, &l.Title, &desc,
		&l.LessonType, &l.SortOrder, &l.CreatedAt,
		&stepsJSON, &materialsJSON, &tagsJSON, &checklistJSON,
	)
	if err != nil {
		return nil, err
	}
	if desc != nil {
		l.Description = *desc
	}

	l.Steps = []LessonStep{}
	l.Materials = []LessonMaterial{}
	l.Tags = []string{}
	l.Checklist = []ChecklistItem{}

	if err := json.Unmarshal(stepsJSON, &l.Steps); err != nil {
		return nil, fmt.Errorf("parse lesson steps: %w", err)
	}
	if err := json.Unmarshal(materialsJSON, &l.Materials); err != nil {
		return nil, fmt.Errorf("parse lesson materials: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &l.Tags); err != nil {
		return nil, fmt.Errorf("parse lesson tags: %w", err)
	}
	if err := json.Unmarshal(checklistJSON, &l.Checklist); err != nil {
		return nil, fmt.Errorf("parse lesson checklist: %w", err)
	}

	return &l, nil
}

func (r *Repo) getStepsByLessonID(ctx context.Context, lessonID uuid.UUID) ([]LessonStep, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, title, content, sort_order
		FROM lesson_steps WHERE lesson_id = $1 ORDER BY sort_order`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []LessonStep{}
	for rows.Next() {
		var s LessonStep
		var content *string
		if err := rows.Scan(&s.ID, &s.LessonID, &s.Title, &content, &s.SortOrder); err != nil {
			return nil, err
		}
		if content != nil {
			s.Content = *content
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *Repo) getMaterialsByLessonID(ctx context.Context, lessonID uuid.UUID) ([]LessonMaterial, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, kind, url_or_path, title
		FROM lesson_materials WHERE lesson_id = $1`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []LessonMaterial{}
	for rows.Next() {
		var m LessonMaterial
		var title *string
		if err := rows.Scan(&m.ID, &m.LessonID, &m.Kind, &m.URLOrPath, &title); err != nil {
			return nil, err
		}
		if title != nil {
			m.Title = *title
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (r *Repo) getTagsByLessonID(ctx context.Context, lessonID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT tag FROM lesson_tags WHERE lesson_id = $1`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []string{}
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *Repo) getChecklistByLessonID(ctx context.Context, lessonID uuid.UUID) ([]ChecklistItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, title, sort_order
		FROM checklist_items WHERE lesson_id = $1 ORDER BY sort_order`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []ChecklistItem{}
	for rows.Next() {
		var c ChecklistItem
		if err := rows.Scan(&c.ID, &c.LessonID, &c.Title, &c.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// UpdateLesson updates lesson title and description. If steps is not nil, replaces all steps for the lesson.
func (r *Repo) UpdateLesson(ctx context.Context, id uuid.UUID, title, description *string, steps []LessonStep) (*Lesson, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if title != nil && *title == "" {
		return nil, errors.New("title must not be empty")
	}
	if title != nil || description != nil {
		var titleVal, descVal interface{}
		if title != nil {
			titleVal = *title
		}
		if description != nil {
			descVal = *description
		}
		_, err = tx.Exec(ctx,
			`UPDATE lessons SET
				title       = COALESCE($2, title),
				description = COALESCE($3, description)
			WHERE id = $1`,
			id, titleVal, descVal)
		if err != nil {
			return nil, err
		}
	}
	if steps != nil {
		if _, err = tx.Exec(ctx, "DELETE FROM lesson_steps WHERE lesson_id = $1", id); err != nil {
			return nil, err
		}
		for i, s := range steps {
			sortOrder := i
			if s.SortOrder > 0 {
				sortOrder = s.SortOrder
			}
			stepID := uuid.New()
			_, err = tx.Exec(ctx,
				"INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order) VALUES ($1, $2, $3, $4, $5)",
				stepID, id, s.Title, nullIfEmpty(s.Content), sortOrder)
			if err != nil {
				return nil, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetLessonByID(ctx, id)
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
