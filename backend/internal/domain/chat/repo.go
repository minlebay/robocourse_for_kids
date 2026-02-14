package chat

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

type StoredMessage struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *Repo) ListByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) ([]StoredMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT role, text, created_at
		FROM chat_messages
		WHERE user_id = $1 AND lesson_id = $2
		ORDER BY created_at ASC
	`, userID, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []StoredMessage
	for rows.Next() {
		var m StoredMessage
		if err := rows.Scan(&m.Role, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (r *Repo) Save(ctx context.Context, userID, lessonID uuid.UUID, role, text string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_messages (user_id, lesson_id, role, text)
		VALUES ($1, $2, $3, $4)
	`, userID, lessonID, role, text)
	return err
}

func (r *Repo) DeleteByUserAndLesson(ctx context.Context, userID, lessonID uuid.UUID) (int64, error) {
	cmd, err := r.pool.Exec(ctx, `
		DELETE FROM chat_messages WHERE user_id = $1 AND lesson_id = $2
	`, userID, lessonID)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}
