package usecase

import (
	"context"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	users  domain.UserRepository
	tokens domain.TokenService
}

type AuthResult struct {
	Token string
	User  *domain.User
}

func NewAuthUseCase(users domain.UserRepository, tokens domain.TokenService) *AuthUseCase {
	return &AuthUseCase{
		users:  users,
		tokens: tokens,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	if err := domain.ValidateRegisterInput(email, password); err != nil {
		return nil, err
	}

	existing, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        domain.NormalizeEmail(email),
		PasswordHash: string(passwordHash),
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := uc.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""

	return &AuthResult{
		Token: token,
		User:  user,
	}, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	if err := domain.ValidateLoginInput(email, password); err != nil {
		return nil, err
	}

	user, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := uc.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""

	return &AuthResult{
		Token: token,
		User:  user,
	}, nil
}
