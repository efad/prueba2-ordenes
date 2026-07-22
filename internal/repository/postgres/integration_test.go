//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/efad/prueba2-ordenes/internal/repository/postgres"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://orders:orders@localhost:5432/orders?sslmode=disable"
	}

	return databaseURL
}

func setupRepos(t *testing.T) (*postgres.DB, *postgres.UserRepository, *postgres.ProductRepository, *postgres.OrderRepository, *postgres.TransactionManager) {
	t.Helper()

	ctx := context.Background()
	db, err := postgres.NewDB(ctx, testDatabaseURL(t))
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}

	t.Cleanup(db.Close)

	return db,
		postgres.NewUserRepository(db),
		postgres.NewProductRepository(db),
		postgres.NewOrderRepository(db),
		postgres.NewTransactionManager(db)
}

func TestCreateOrderTransactionDecrementsStock(t *testing.T) {
	_, users, products, orders, txManager := setupRepos(t)
	ctx := context.Background()

	user := &domain.User{
		Email:        "buyer-" + time.Now().Format("150405.000000") + "@example.com",
		PasswordHash: "hash",
	}
	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("Create user error = %v", err)
	}

	product := &domain.Product{
		Name:  "Producto test",
		Price: 25.5,
		Stock: 10,
	}
	if err := products.Create(ctx, product); err != nil {
		t.Fatalf("Create product error = %v", err)
	}

	order := &domain.Order{
		UserID: user.ID,
		Items: []domain.OrderItem{
			{ProductID: product.ID, Quantity: 3, UnitPrice: product.Price},
		},
		Total: domain.CalculateOrderTotal([]domain.OrderItem{
			{ProductID: product.ID, Quantity: 3, UnitPrice: product.Price},
		}),
	}

	err := txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := products.DecrementStock(txCtx, product.ID, 3); err != nil {
			return err
		}
		return orders.Create(txCtx, order)
	})
	if err != nil {
		t.Fatalf("RunInTransaction() error = %v", err)
	}

	updated, err := products.FindByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByID product error = %v", err)
	}
	if updated.Stock != 7 {
		t.Fatalf("stock = %d, want 7", updated.Stock)
	}

	savedOrder, err := orders.FindByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("FindByID order error = %v", err)
	}
	if savedOrder.Total != 76.5 {
		t.Fatalf("total = %v, want 76.5", savedOrder.Total)
	}
}

func TestCancelOrderRestoresStock(t *testing.T) {
	_, users, products, orders, txManager := setupRepos(t)
	ctx := context.Background()

	user := &domain.User{
		Email:        "cancel-" + time.Now().Format("150405.000000") + "@example.com",
		PasswordHash: "hash",
	}
	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("Create user error = %v", err)
	}

	product := &domain.Product{
		Name:  "Producto cancel",
		Price: 10,
		Stock: 5,
	}
	if err := products.Create(ctx, product); err != nil {
		t.Fatalf("Create product error = %v", err)
	}

	order := &domain.Order{
		UserID: user.ID,
		Items: []domain.OrderItem{
			{ProductID: product.ID, Quantity: 2, UnitPrice: product.Price},
		},
		Total: 20,
	}

	err := txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := products.DecrementStock(txCtx, product.ID, 2); err != nil {
			return err
		}
		return orders.Create(txCtx, order)
	})
	if err != nil {
		t.Fatalf("create order tx error = %v", err)
	}

	cancelled, err := orders.Cancel(ctx, order.ID, user.ID)
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if cancelled.Status != domain.OrderStatusCancelled {
		t.Fatalf("status = %s, want CANCELLED", cancelled.Status)
	}
	if cancelled.DeletedAt == nil {
		t.Fatal("deleted_at expected on cancelled order")
	}

	updated, err := products.FindByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("FindByID product error = %v", err)
	}
	if updated.Stock != 5 {
		t.Fatalf("stock = %d, want 5", updated.Stock)
	}
}
