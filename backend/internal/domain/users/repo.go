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

func (r *Repo) Create(ctx context.Context, login, passwordHash, name, role, email string, mustChangePassword bool) (*User, error) {
	id := uuid.New()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, login, password_hash, name, role, theme, email, must_change_password)
		VALUES ($1, $2, $3, $4, $5, 'default', NULLIF($6, ''), $7)`,
		id, login, passwordHash, name, role, email, mustChangePassword)
	if err != nil {
		return nil, err
	}
	var emailPtr *string
	if email != "" {
		e := email
		emailPtr = &e
	}
	return &User{ID: id, Login: login, Name: name, Role: role, Theme: "default", Email: emailPtr, MustChangePassword: mustChangePassword}, nil
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

// GetByIDWithPassword returns the user including their password hash.
// Use only when password verification is required (e.g. ChangePassword).
func (r *Repo) GetByIDWithPassword(ctx context.Context, id uuid.UUID) (*UserWithPassword, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, login, password_hash, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users WHERE id = $1`, id)
	var u UserWithPassword
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Name, &u.Role, &u.Theme, &u.CreatedAt,
		&u.Email, &u.MustChangePassword, &u.IsBlocked)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// IsBlocked returns true if the user with the given ID has is_blocked = true.
// Returns false (not blocked) when the user does not exist.
func (r *Repo) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	var blocked bool
	err := r.pool.QueryRow(ctx, `SELECT is_blocked FROM users WHERE id = $1`, id).Scan(&blocked)
	if err != nil {
		return false, err
	}
	return blocked, nil
}

// List returns students only (role = 'student'), for use by teacher endpoints.
// limit/offset control pagination; pass limit=0 to use the default cap (500).
func (r *Repo) List(ctx context.Context, limit, offset int) ([]User, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, login, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users
		WHERE role = 'student'
		ORDER BY created_at
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

// ListAll returns all users regardless of role, for use by admin endpoints.
// limit/offset control pagination; pass limit=0 to use the default cap (500).
func (r *Repo) ListAll(ctx context.Context, limit, offset int) ([]User, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, login, name, role, COALESCE(theme, 'default'), created_at,
		       email, must_change_password, is_blocked
		FROM users
		ORDER BY created_at
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUsers(rows)
}

// scanUsers scans a pgx.Rows result into a slice of User.
func scanUsers(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]User, error) {
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

// UpdatePasswordAndMustChange atomically updates the password hash and must_change_password flag.
// Returns true if a row was found and updated (false means the user does not exist).
func (r *Repo) UpdatePasswordAndMustChange(ctx context.Context, id uuid.UUID, hash string, mustChange bool) (bool, error) {
	res, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, must_change_password = $2 WHERE id = $3`,
		hash, mustChange, id)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() > 0, nil
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
	return scanUsers(rows)
}
