package domain_test

import (
	"testing"

	"github.com/efad/prueba2-ordenes/internal/domain"
)

func TestValidateRegisterInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "rechaza email vacio",
			email:    "",
			password: "password1",
			wantErr:  true,
		},
		{
			name:     "rechaza contrasena corta",
			email:    "user@example.com",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "acepta datos validos",
			email:    "user@example.com",
			password: "password1",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateRegisterInput(tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRegisterInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	got := domain.NormalizeEmail("  User@Example.COM ")
	want := "user@example.com"

	if got != want {
		t.Fatalf("NormalizeEmail() = %q, want %q", got, want)
	}
}

func TestValidateProductFilter(t *testing.T) {
	t.Parallel()

	minPrice := 100.0
	maxPrice := 50.0

	err := domain.ValidateProductFilter(domain.ProductFilter{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err == nil {
		t.Fatal("ValidateProductFilter() esperaba error cuando min > max")
	}
}
