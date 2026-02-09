package users

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

func (r *Repo) Create(ctx context.Context, login, passwordHash, name, role string) (*User, error) {
	id := uuid.New()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, login, password_hash, name, role)
		VALUES ($1, $2, $3, $4, $5)`,
		id, login, passwordHash, name, role)
	if err != nil {
		return nil, err
	}
	return &User{ID: id, Login: login, Name: name, Role: role}, nil
}

func (r *Repo) GetByLogin(ctx context.Context, login string) (*UserWithPassword, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, login, password_hash, name, role, created_at
		FROM users WHERE login = $1`, login)
	var u UserWithPassword
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, login, name, role, created_at
		FROM users WHERE id = $1`, id)
	var u User
	err := row.Scan(&u.ID, &u.Login, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, login, name, role, created_at
		FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Login, &u.Name, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}
