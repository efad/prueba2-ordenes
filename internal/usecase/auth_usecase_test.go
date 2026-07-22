package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/efad/prueba2-ordenes/internal/domain"
	jwtservice "github.com/efad/prueba2-ordenes/internal/service/jwt"
	"github.com/efad/prueba2-ordenes/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	usersByEmail map[string]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(_ context.Context, user *domain.User) error {
	normalized := domain.NormalizeEmail(user.Email)
	if _, exists := m.usersByEmail[normalized]; exists {
		return domain.ErrEmailAlreadyExists
	}

	copyUser := *user
	if copyUser.ID == "" {
		copyUser.ID = "user-" + normalized
	}
	if copyUser.CreatedAt.IsZero() {
		copyUser.CreatedAt = time.Now().UTC()
	}

	m.usersByEmail[normalized] = &copyUser
	*user = copyUser
	return nil
}

func (m *mockUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := m.usersByEmail[domain.NormalizeEmail(email)]
	if !ok {
		return nil, nil
	}

	copyUser := *user
	return &copyUser, nil
}

func (m *mockUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	for _, user := range m.usersByEmail {
		if user.ID == id {
			copyUser := *user
			return &copyUser, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) FindByIDs(_ context.Context, ids []string) ([]domain.User, error) {
	return []domain.User{}, nil
}

func setupAuthUseCase(t *testing.T) (*usecase.AuthUseCase, *mockUserRepository) {
	t.Helper()

	tokenService, err := jwtservice.NewService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	users := newMockUserRepository()
	return usecase.NewAuthUseCase(users, tokenService), users
}

func TestAuthUseCaseRegisterSuccess(t *testing.T) {
	t.Parallel()

	authUC, _ := setupAuthUseCase(t)
	ctx := context.Background()

	result, err := authUC.Register(ctx, "user@example.com", "password1")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token on register")
	}
	if result.User.Email != "user@example.com" {
		t.Fatalf("email = %q, want user@example.com", result.User.Email)
	}
}

func TestAuthUseCaseRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	authUC, _ := setupAuthUseCase(t)
	ctx := context.Background()

	if _, err := authUC.Register(ctx, "dup@example.com", "password1"); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	_, err := authUC.Register(ctx, "dup@example.com", "password2")
	if err != domain.ErrEmailAlreadyExists {
		t.Fatalf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestAuthUseCaseLoginInvalidCredentials(t *testing.T) {
	t.Parallel()

	authUC, users := setupAuthUseCase(t)
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	users.usersByEmail["known@example.com"] = &domain.User{
		ID:           "user-1",
		Email:        "known@example.com",
		PasswordHash: string(hash),
	}

	_, err = authUC.Login(ctx, "unknown@example.com", "password1")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("Login unknown user error = %v, want ErrInvalidCredentials", err)
	}

	_, err = authUC.Login(ctx, "known@example.com", "wrong-password")
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("Login wrong password error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthUseCaseLoginSuccess(t *testing.T) {
	t.Parallel()

	authUC, _ := setupAuthUseCase(t)
	ctx := context.Background()

	if _, err := authUC.Register(ctx, "login@example.com", "password1"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	result, err := authUC.Login(ctx, "login@example.com", "password1")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token on login")
	}
}
