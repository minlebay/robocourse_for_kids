package users

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"net/mail"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// HTTPError carries an HTTP status code and client-facing message from the service layer.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

func httpErr(code int, msg string) *HTTPError {
	return &HTTPError{Code: code, Message: msg}
}

// dummyHash is a pre-computed bcrypt hash used for timing attack protection.
// When a login attempt is made for a non-existent user, we still run bcrypt
// to ensure consistent response times and prevent user enumeration.
var dummyHash []byte

func init() {
	var err error
	dummyHash, err = bcrypt.GenerateFromPassword([]byte("timing-attack-protection"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("users: failed to pre-compute dummy bcrypt hash: %v", err)
	}
}

// Service holds the business logic for the users domain.
type Service struct {
	repo       Repository
	inviteCode string // required for teacher registration; empty = disabled
}

func NewService(repo Repository, inviteCode string) *Service {
	return &Service{repo: repo, inviteCode: inviteCode}
}

// validateCredentials checks login/password/name lengths and returns an HTTPError if invalid.
func validateCredentials(login, password, name string) *HTTPError {
	if len(login) < 3 || len(login) > 50 {
		return httpErr(http.StatusBadRequest, "login must be 3-50 characters")
	}
	if len(password) < 6 {
		return httpErr(http.StatusBadRequest, "password must be at least 6 characters")
	}
	// bcrypt silently truncates at 72 bytes
	if len(password) > 72 {
		return httpErr(http.StatusBadRequest, "password must be at most 72 characters")
	}
	if len(name) > 200 {
		return httpErr(http.StatusBadRequest, "name must be at most 200 characters")
	}
	return nil
}

// validatePasswordLength checks only password constraints (for password change flow).
func validatePasswordLength(password string) *HTTPError {
	if len(password) < 6 {
		return httpErr(http.StatusBadRequest, "password must be at least 6 characters")
	}
	if len(password) > 72 {
		return httpErr(http.StatusBadRequest, "password must be at most 72 characters")
	}
	return nil
}

// Register validates credentials and role, enforces invite code for teachers,
// hashes the password, and creates the user.
func (s *Service) Register(ctx context.Context, login, password, name, role, inviteCode, email string, mustChangePassword bool) (*User, error) {
	if he := validateCredentials(login, password, name); he != nil {
		return nil, he
	}
	if role == "" {
		role = RoleStudent
	}
	if role != RoleStudent && role != RoleTeacher {
		return nil, httpErr(http.StatusBadRequest, "role must be student or teacher")
	}
	if role == RoleTeacher {
		if s.inviteCode == "" {
			return nil, httpErr(http.StatusForbidden, "teacher registration is not available")
		}
		if inviteCode != s.inviteCode {
			slog.Warn("invalid teacher invite code attempt", "login", login)
			return nil, httpErr(http.StatusForbidden, "invalid invite code")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.Create(ctx, login, string(hash), name, role, email, mustChangePassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httpErr(http.StatusBadRequest, "login already exists")
		}
		return nil, err
	}
	return user, nil
}

// Authenticate verifies login credentials and block status using timing-safe comparison.
func (s *Service) Authenticate(ctx context.Context, login, password string) (*UserWithPassword, error) {
	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		// Timing attack protection: always run bcrypt even if user not found.
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpErr(http.StatusUnauthorized, "invalid login or password")
		}
		return nil, err
	}
	if u.IsBlocked {
		return nil, httpErr(http.StatusForbidden, "user is blocked")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, httpErr(http.StatusUnauthorized, "invalid login or password")
	}
	return u, nil
}

// ValidateDeleteUser ensures a teacher cannot delete themselves and can only delete students.
// Returns the target user on success.
func (s *Service) ValidateDeleteUser(ctx context.Context, currentID, targetID uuid.UUID) (*User, error) {
	if currentID == targetID {
		return nil, httpErr(http.StatusForbidden, "cannot delete yourself")
	}
	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpErr(http.StatusNotFound, "user not found")
		}
		return nil, err
	}
	hasStudent, err := s.repo.HasRole(ctx, targetID, RoleStudent)
	if err != nil {
		return nil, err
	}
	if !hasStudent {
		return nil, httpErr(http.StatusForbidden, "teachers can only delete students")
	}
	return target, nil
}

// ChangePassword verifies the current password and updates to the new one.
// Returns the user's role (needed to issue a fresh JWT after the change).
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) (string, error) {
	if he := validatePasswordLength(newPassword); he != nil {
		return "", he
	}
	uwp, err := s.repo.GetByIDWithPassword(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", httpErr(http.StatusNotFound, "user not found")
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(uwp.PasswordHash), []byte(currentPassword)); err != nil {
		return "", httpErr(http.StatusUnauthorized, "current password is incorrect")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if _, err := s.repo.UpdatePasswordAndMustChange(ctx, userID, string(newHash), false); err != nil {
		return "", err
	}
	return uwp.Role, nil
}

// AdminCreate validates and creates a user (admin endpoint — no invite code required).
func (s *Service) AdminCreate(ctx context.Context, login, password, name, role, email string) (*User, error) {
	if he := validateCredentials(login, password, name); he != nil {
		return nil, he
	}
	if role == "" {
		role = RoleStudent
	}
	if role != RoleStudent && role != RoleTeacher && role != RoleAdministrator {
		return nil, httpErr(http.StatusBadRequest, "role must be student, teacher, or administrator")
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return nil, httpErr(http.StatusBadRequest, "invalid email address")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.Create(ctx, login, string(hash), name, role, email, true)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httpErr(http.StatusBadRequest, "login already exists")
		}
		return nil, err
	}
	return user, nil
}

// ResetPassword generates a temporary password, hashes it, and updates the user record.
func (s *Service) ResetPassword(ctx context.Context, targetID uuid.UUID) (string, error) {
	tempPassword := generateTempPassword(10)
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	found, err := s.repo.UpdatePasswordAndMustChange(ctx, targetID, string(hash), true)
	if err != nil {
		return "", err
	}
	if !found {
		return "", httpErr(http.StatusNotFound, "user not found")
	}
	return tempPassword, nil
}

// IsBlocked reports whether the user is blocked. Used by the auth middleware.
func (s *Service) IsBlocked(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.IsBlocked(ctx, id)
}

// GetByID returns a user by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetRoles returns the list of roles for the user.
func (s *Service) GetRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.repo.GetRoles(ctx, userID)
}

// HasRole returns true if the user has the given role.
func (s *Service) HasRole(ctx context.Context, userID uuid.UUID, role string) (bool, error) {
	return s.repo.HasRole(ctx, userID, role)
}

// UpdateTheme updates the user's UI theme preference.
func (s *Service) UpdateTheme(ctx context.Context, id uuid.UUID, theme string) error {
	return s.repo.UpdateTheme(ctx, id, theme)
}

// List returns students (for teacher dashboard) with roles attached from user_roles.
func (s *Service) List(ctx context.Context, limit, offset int) ([]User, error) {
	list, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.attachRoles(ctx, list)
}

// ListAll returns all users (admin only) with roles attached from user_roles.
func (s *Service) ListAll(ctx context.Context, limit, offset int) ([]User, error) {
	list, err := s.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return s.attachRoles(ctx, list)
}

// attachRoles fills Roles and Role (primary) for each user from user_roles.
func (s *Service) attachRoles(ctx context.Context, list []User) ([]User, error) {
	if len(list) == 0 {
		return list, nil
	}
	ids := make([]uuid.UUID, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}
	rolesByID, err := s.repo.GetRolesByUserIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range list {
		roles := rolesByID[list[i].ID]
		list[i].Roles = roles
		list[i].Role = PrimaryRole(roles)
	}
	return list, nil
}

// Delete removes a user by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	return s.repo.Delete(ctx, id)
}

// BlockUser blocks or unblocks a user.
func (s *Service) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
	return s.repo.BlockUser(ctx, id, block)
}

// GetStats returns platform-wide statistics.
func (s *Service) GetStats(ctx context.Context) (usersCount, modulesCount, lessonsCount int, err error) {
	return s.repo.GetStats(ctx)
}

// GetActivity returns recently registered users with roles attached from user_roles.
func (s *Service) GetActivity(ctx context.Context, limit int) ([]User, error) {
	list, err := s.repo.GetActivity(ctx, limit)
	if err != nil {
		return nil, err
	}
	return s.attachRoles(ctx, list)
}

// generateTempPassword generates a cryptographically random alphanumeric password.
func generateTempPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
