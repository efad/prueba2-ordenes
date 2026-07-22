package domain

import (
	"strings"
	"time"
)

const minPasswordLength = 8

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func ValidateRegisterInput(email, password string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	return validatePassword(password)
}

func ValidateLoginInput(email, password string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if strings.TrimSpace(password) == "" {
		return invalidInput("la contrasena es obligatoria")
	}
	return nil
}

func validateEmail(email string) error {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return invalidInput("el email es obligatorio")
	}
	if !strings.Contains(normalized, "@") {
		return invalidInput("el email no es valido")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return invalidInput("la contrasena debe tener al menos 8 caracteres")
	}
	return nil
}

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
