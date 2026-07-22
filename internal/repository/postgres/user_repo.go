package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	_, err := r.db.querier(ctx).Exec(ctx, `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, user.ID, domain.NormalizeEmail(user.Email), user.PasswordHash, user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.querier(ctx).QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, domain.NormalizeEmail(email))

	user, err := scanUser(row)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.querier(ctx).QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id)

	user, err := scanUser(row)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrUnauthorized
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByIDs(ctx context.Context, ids []string) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}

	rows, err := r.db.querier(ctx).Query(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("find users by ids: %w", err)
	}
	defer rows.Close()

	usersByID := make(map[string]domain.User, len(ids))
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		usersByID[user.ID] = *user
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(ids))
	for _, id := range ids {
		if user, ok := usersByID[id]; ok {
			users = append(users, user)
		}
	}

	return users, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(row scannable) (*domain.User, error) {
	var user domain.User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
