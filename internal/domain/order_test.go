package domain_test

import (
	"testing"

	"github.com/efad/prueba2-ordenes/internal/domain"
)

func TestValidateCreateOrderItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		items   []domain.CreateOrderItemInput
		wantErr bool
	}{
		{
			name:    "rechaza lista vacia",
			items:   nil,
			wantErr: true,
		},
		{
			name: "rechaza cantidad invalida",
			items: []domain.CreateOrderItemInput{
				{ProductID: "prod-1", Quantity: 0},
			},
			wantErr: true,
		},
		{
			name: "rechaza productos duplicados",
			items: []domain.CreateOrderItemInput{
				{ProductID: "prod-1", Quantity: 1},
				{ProductID: "prod-1", Quantity: 2},
			},
			wantErr: true,
		},
		{
			name: "acepta items validos",
			items: []domain.CreateOrderItemInput{
				{ProductID: "prod-1", Quantity: 2},
				{ProductID: "prod-2", Quantity: 1},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateCreateOrderItems(tt.items)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCreateOrderItems() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateOrderTotal(t *testing.T) {
	t.Parallel()

	items := []domain.OrderItem{
		{ProductID: "prod-1", Quantity: 2, UnitPrice: 10.5},
		{ProductID: "prod-2", Quantity: 1, UnitPrice: 5.0},
	}

	got := domain.CalculateOrderTotal(items)
	want := 26.0

	if got != want {
		t.Fatalf("CalculateOrderTotal() = %v, want %v", got, want)
	}
}

func TestOrderCanCancel(t *testing.T) {
	t.Parallel()

	order := &domain.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: domain.OrderStatusPending,
	}

	if err := order.CanCancel("user-2"); err != domain.ErrOrderNotOwned {
		t.Fatalf("CanCancel() otro usuario = %v, want %v", err, domain.ErrOrderNotOwned)
	}

	order.Status = domain.OrderStatusConfirmed
	if err := order.CanCancel("user-1"); err != domain.ErrOrderNotCancellable {
		t.Fatalf("CanCancel() estado confirmado = %v, want %v", err, domain.ErrOrderNotCancellable)
	}

	order.Status = domain.OrderStatusPending
	if err := order.CanCancel("user-1"); err != nil {
		t.Fatalf("CanCancel() valido = %v, want nil", err)
	}
}
