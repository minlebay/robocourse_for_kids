package lessons

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) ListModules(ctx context.Context, platform, tag *string) ([]Module, error) {
	query := `
		SELECT id, title, description, sort_order, created_at
		FROM modules
		ORDER BY sort_order, created_at`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Module
	for rows.Next() {
		var m Module
		err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.SortOrder, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		// Filter by tag if needed (simplified: load all and filter in app, or join lesson_tags)
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if tag != nil || platform != nil {
		filtered := list[:0]
		for _, m := range list {
			lessons, _ := r.listLessonsByModuleID(ctx, m.ID)
			keep := true
			if tag != nil {
				hasTag := false
				for _, l := range lessons {
					for _, t := range l.Tags {
						if t == *tag {
							hasTag = true
							break
						}
					}
				}
				if !hasTag {
					keep = false
				}
			}
			if keep {
				filtered = append(filtered, m)
			}
		}
		list = filtered
	}

	return list, nil
}

func (r *Repo) CreateModule(ctx context.Context, title, description string, sortOrder int) (*Module, error) {
	if sortOrder == 0 {
		var maxOrder int
		_ = r.pool.QueryRow(ctx, "SELECT COALESCE(MAX(sort_order), 0) + 1 FROM modules").Scan(&maxOrder)
		sortOrder = maxOrder
	}
	var m Module
	var desc *string
	if description != "" {
		desc = &description
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO modules (title, description, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, sort_order, created_at
	`, title, desc, sortOrder).Scan(&m.ID, &m.Title, &m.Description, &m.SortOrder, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
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
	if sortOrder == 0 {
		var maxOrder int
		_ = r.pool.QueryRow(ctx, "SELECT COALESCE(MAX(sort_order), 0) + 1 FROM lessons WHERE module_id = $1", moduleID).Scan(&maxOrder)
		sortOrder = maxOrder
	}
	var l Lesson
	var desc *string
	if description != "" {
		desc = &description
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO lessons (module_id, title, description, lesson_type, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, module_id, title, description, lesson_type, sort_order, created_at
	`, moduleID, title, desc, lessonType, sortOrder).Scan(&l.ID, &l.ModuleID, &l.Title, &l.Description, &l.LessonType, &l.SortOrder, &l.CreatedAt)
	if err != nil {
		return nil, err
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
			_, err := r.pool.Exec(ctx,
				"INSERT INTO lesson_steps (lesson_id, title, content, sort_order) VALUES ($1, $2, $3, $4)",
				l.ID, s.Title, nullIfEmpty(s.Content), so)
			if err != nil {
				return nil, err
			}
		}
		l.Steps, _ = r.getStepsByLessonID(ctx, l.ID)
	}
	return &l, nil
}

func (r *Repo) GetModuleByID(ctx context.Context, id uuid.UUID) (*Module, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, description, sort_order, created_at
		FROM modules WHERE id = $1`, id)
	var m Module
	err := row.Scan(&m.ID, &m.Title, &m.Description, &m.SortOrder, &m.CreatedAt)
	if err != nil {
		return nil, err
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
	var list []Lesson
	for rows.Next() {
		var l Lesson
		err := rows.Scan(&l.ID, &l.ModuleID, &l.Title, &l.Description, &l.LessonType, &l.SortOrder, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

func (r *Repo) GetLessonByID(ctx context.Context, id uuid.UUID) (*Lesson, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, module_id, title, description, lesson_type, sort_order, created_at
		FROM lessons WHERE id = $1`, id)
	var l Lesson
	err := row.Scan(&l.ID, &l.ModuleID, &l.Title, &l.Description, &l.LessonType, &l.SortOrder, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	l.Steps, _ = r.getStepsByLessonID(ctx, id)
	l.Materials, _ = r.getMaterialsByLessonID(ctx, id)
	l.Tags, _ = r.getTagsByLessonID(ctx, id)
	l.Checklist, _ = r.getChecklistByLessonID(ctx, id)
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
	var list []LessonStep
	for rows.Next() {
		var s LessonStep
		if err := rows.Scan(&s.ID, &s.LessonID, &s.Title, &s.Content, &s.SortOrder); err != nil {
			return nil, err
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
	var list []LessonMaterial
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
	var tags []string
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
	var list []ChecklistItem
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
	if title != nil || description != nil {
		if title != nil && *title == "" {
			return nil, errors.New("title must not be empty")
		}
		// Build dynamic update
		updates := []string{}
		args := []interface{}{id}
		pos := 2
		if title != nil {
			updates = append(updates, "title = $"+strconv.Itoa(pos))
			args = append(args, *title)
			pos++
		}
		if description != nil {
			updates = append(updates, "description = $"+strconv.Itoa(pos))
			args = append(args, *description)
			pos++
		}
		if len(updates) > 0 {
			query := "UPDATE lessons SET " + strings.Join(updates, ", ") + " WHERE id = $1"
			_, err := r.pool.Exec(ctx, query, args...)
			if err != nil {
				return nil, err
			}
		}
	}
	if steps != nil {
		if _, err := r.pool.Exec(ctx, "DELETE FROM lesson_steps WHERE lesson_id = $1", id); err != nil {
			return nil, err
		}
		for i, s := range steps {
			sortOrder := i
			if s.SortOrder > 0 {
				sortOrder = s.SortOrder
			}
			stepID := uuid.New()
			_, err := r.pool.Exec(ctx,
				"INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order) VALUES ($1, $2, $3, $4, $5)",
				stepID, id, s.Title, nullIfEmpty(s.Content), sortOrder)
			if err != nil {
				return nil, err
			}
		}
	}
	return r.GetLessonByID(ctx, id)
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
