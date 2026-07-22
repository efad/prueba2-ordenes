package jwt

import (
	"fmt"
	"time"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret     []byte
	expiration time.Duration
}

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewService(secret string, expiration time.Duration) (*Service, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret es obligatorio")
	}
	if expiration <= 0 {
		expiration = 24 * time.Hour
	}

	return &Service{
		secret:     []byte(secret),
		expiration: expiration,
	}, nil
}

func (s *Service) Generate(userID string) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
		},
	})

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("firmar jwt: %w", err)
	}

	return signed, nil
}

func (s *Service) Parse(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("metodo de firma invalido")
		}
		return s.secret, nil
	})
	if err != nil {
		return "", domain.ErrUnauthorized
	}

	parsedClaims, ok := token.Claims.(*claims)
	if !ok || !token.Valid || parsedClaims.UserID == "" {
		return "", domain.ErrUnauthorized
	}

	return parsedClaims.UserID, nil
}
