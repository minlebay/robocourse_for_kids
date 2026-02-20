package users

import (
	"context"

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

func (r *Repo) Create(ctx context.Context, login, passwordHash, name, role, email string) (*User, error) {
	id := uuid.New()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, login, password_hash, name, role, theme, email)
		VALUES ($1, $2, $3, $4, $5, 'default', NULLIF($6, ''))`,
		id, login, passwordHash, name, role, email)
	if err != nil {
		return nil, err
	}
	var emailPtr *string
	if email != "" {
		e := email
		emailPtr = &e
	}
	return &User{ID: id, Login: login, Name: name, Role: role, Theme: "default", Email: emailPtr}, nil
}

func (r *Repo) GetByLogin(ctx context.Context, login string) (*UserWithPassword, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, login, password_hash, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users WHERE login = $1`, login)
	var u UserWithPassword
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Name, &u.Role, &u.Theme, &u.CreatedAt,
		&u.Email, &u.MustChangePassword, &u.IsBlocked)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, login, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users WHERE id = $1`, id)
	var u User
	err := row.Scan(&u.ID, &u.Login, &u.Name, &u.Role, &u.Theme, &u.CreatedAt,
		&u.Email, &u.MustChangePassword, &u.IsBlocked)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, login, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Login, &u.Name, &u.Role, &u.Theme, &u.CreatedAt,
			&u.Email, &u.MustChangePassword, &u.IsBlocked); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *Repo) ListAll(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, login, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Login, &u.Name, &u.Role, &u.Theme, &u.CreatedAt,
			&u.Email, &u.MustChangePassword, &u.IsBlocked); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
}

func (r *Repo) UpdateTheme(ctx context.Context, id uuid.UUID, theme string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET theme = $1 WHERE id = $2`, theme, id)
	return err
}

func (r *Repo) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET is_blocked = $1 WHERE id = $2`, block, id)
	return err
}

func (r *Repo) SetMustChangePassword(ctx context.Context, id uuid.UUID, v bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET must_change_password = $1 WHERE id = $2`, v, id)
	return err
}

func (r *Repo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func (r *Repo) GetStats(ctx context.Context) (usersCount, modulesCount, lessonsCount int, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM modules),
			(SELECT COUNT(*) FROM lessons)
	`)
	err = row.Scan(&usersCount, &modulesCount, &lessonsCount)
	return
}

func (r *Repo) GetActivity(ctx context.Context, limit int) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, login, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Login, &u.Name, &u.Role, &u.Theme, &u.CreatedAt,
			&u.Email, &u.MustChangePassword, &u.IsBlocked); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}
