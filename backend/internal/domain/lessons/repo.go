package lessons

import (
	"context"

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
