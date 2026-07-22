package usecase_test

import (
	"context"
	"testing"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/efad/prueba2-ordenes/internal/usecase"
)

type mockOrderRepository struct {
	orders map[string]*domain.Order
}

func (m *mockOrderRepository) Create(_ context.Context, order *domain.Order) error {
	if m.orders == nil {
		m.orders = make(map[string]*domain.Order)
	}
	copyOrder := *order
	if copyOrder.ID == "" {
		copyOrder.ID = "order-1"
	}
	m.orders[copyOrder.ID] = &copyOrder
	*order = copyOrder
	return nil
}

func (m *mockOrderRepository) FindByID(_ context.Context, id string) (*domain.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	copyOrder := *order
	return &copyOrder, nil
}

func (m *mockOrderRepository) ListByUser(_ context.Context, userID string, _, _ int) ([]domain.Order, int, error) {
	orders := make([]domain.Order, 0)
	for _, order := range m.orders {
		if order.UserID == userID {
			orders = append(orders, *order)
		}
	}
	return orders, len(orders), nil
}

func (m *mockOrderRepository) Cancel(_ context.Context, orderID, userID string) (*domain.Order, error) {
	order, err := m.FindByID(context.Background(), orderID)
	if err != nil {
		return nil, err
	}
	if err := order.CanCancel(userID); err != nil {
		return nil, err
	}
	order.Status = domain.OrderStatusCancelled
	return order, nil
}

type mockTxManager struct{}

func (mockTxManager) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestOrderUseCaseCreateOrderSuccess(t *testing.T) {
	t.Parallel()

	products := &mockProductRepository{
		products: []domain.Product{
			{ID: "prod-1", Name: "Teclado", Price: 10, Stock: 5},
		},
	}
	orders := &mockOrderRepository{}
	orderUC := usecase.NewOrderUseCase(orders, products, mockTxManager{})

	order, err := orderUC.CreateOrder(context.Background(), "user-1", []domain.CreateOrderItemInput{
		{ProductID: "prod-1", Quantity: 2},
	})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if order.Total != 20 {
		t.Fatalf("total = %v, want 20", order.Total)
	}
	if products.products[0].Stock != 3 {
		t.Fatalf("stock = %d, want 3", products.products[0].Stock)
	}
}

func TestOrderUseCaseCreateOrderInsufficientStock(t *testing.T) {
	t.Parallel()

	products := &mockProductRepository{
		products: []domain.Product{
			{ID: "prod-1", Name: "Teclado", Price: 10, Stock: 1},
		},
	}
	orderUC := usecase.NewOrderUseCase(&mockOrderRepository{}, products, mockTxManager{})

	_, err := orderUC.CreateOrder(context.Background(), "user-1", []domain.CreateOrderItemInput{
		{ProductID: "prod-1", Quantity: 2},
	})
	if err != domain.ErrInsufficientStock {
		t.Fatalf("CreateOrder() error = %v, want ErrInsufficientStock", err)
	}
}

func TestOrderUseCaseGetOrderNotOwned(t *testing.T) {
	t.Parallel()

	orders := &mockOrderRepository{
		orders: map[string]*domain.Order{
			"order-1": {
				ID:     "order-1",
				UserID: "user-1",
				Status: domain.OrderStatusPending,
			},
		},
	}
	orderUC := usecase.NewOrderUseCase(orders, &mockProductRepository{}, mockTxManager{})

	_, err := orderUC.GetOrder(context.Background(), "user-2", "order-1")
	if err != domain.ErrOrderNotOwned {
		t.Fatalf("GetOrder() error = %v, want ErrOrderNotOwned", err)
	}
}

func TestOrderUseCaseCancelOrderSuccess(t *testing.T) {
	t.Parallel()

	orders := &mockOrderRepository{
		orders: map[string]*domain.Order{
			"order-1": {
				ID:     "order-1",
				UserID: "user-1",
				Status: domain.OrderStatusPending,
			},
		},
	}
	orderUC := usecase.NewOrderUseCase(orders, &mockProductRepository{}, mockTxManager{})

	order, err := orderUC.CancelOrder(context.Background(), "user-1", "order-1")
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if order.Status != domain.OrderStatusCancelled {
		t.Fatalf("status = %s, want CANCELLED", order.Status)
	}
}
