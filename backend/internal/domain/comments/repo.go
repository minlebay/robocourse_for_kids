package comments

import (
	"context"
	"time"

	"github.com/google/uuid"
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

type Comment struct {
	ID        uuid.UUID `json:"id"`
	LessonID  uuid.UUID `json:"lesson_id"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *Repo) ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]Comment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.lesson_id, c.user_id, u.name, c.text, c.created_at
		FROM lesson_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.lesson_id = $1
		ORDER BY c.created_at ASC
	`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.LessonID, &c.UserID, &c.UserName, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *Repo) Create(ctx context.Context, lessonID, userID uuid.UUID, text string) (*Comment, error) {
	var c Comment
	err := r.pool.QueryRow(ctx, `
		INSERT INTO lesson_comments (lesson_id, user_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, lesson_id, user_id, text, created_at,
			(SELECT u.name FROM users u WHERE u.id = lesson_comments.user_id) AS user_name
	`, lessonID, userID, text).Scan(&c.ID, &c.LessonID, &c.UserID, &c.Text, &c.CreatedAt, &c.UserName)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) DeleteByIDAndUser(ctx context.Context, commentID, lessonID, userID uuid.UUID) (deleted bool, err error) {
	cmd, err := r.pool.Exec(ctx, `
		DELETE FROM lesson_comments WHERE id = $1 AND lesson_id = $2 AND user_id = $3
	`, commentID, lessonID, userID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}
