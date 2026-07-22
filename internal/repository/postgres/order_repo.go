package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/google/uuid"
)

type OrderRepository struct {
	db *DB
}

func NewOrderRepository(db *DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	if order.ID == "" {
		order.ID = uuid.NewString()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	if order.Status == "" {
		order.Status = domain.OrderStatusPending
	}

	_, err := r.db.querier(ctx).Exec(ctx, `
		INSERT INTO orders (id, user_id, total, status, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, order.ID, order.UserID, order.Total, order.Status, order.CreatedAt, order.DeletedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for _, item := range order.Items {
		itemID := uuid.NewString()
		_, err := r.db.querier(ctx).Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5)
		`, itemID, order.ID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.querier(ctx).QueryRow(ctx, `
		SELECT id, user_id, total, status, created_at, deleted_at
		FROM orders
		WHERE id = $1
		  AND deleted_at IS NULL
	`, id)

	order, err := scanOrder(row)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("find order by id: %w", err)
	}

	items, err := r.loadOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]domain.Order, int, error) {
	page, pageSize = domain.NormalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	var total int
	if err := r.db.querier(ctx).QueryRow(ctx, `
		SELECT COUNT(*)
		FROM orders
		WHERE user_id = $1
		  AND deleted_at IS NULL
	`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := r.db.querier(ctx).Query(ctx, `
		SELECT id, user_id, total, status, created_at, deleted_at
		FROM orders
		WHERE user_id = $1
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, err
		}
		order.Items = items
		orders = append(orders, *order)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepository) Cancel(ctx context.Context, orderID, userID string) (*domain.Order, error) {
	row := r.db.querier(ctx).QueryRow(ctx, `
		SELECT id, user_id, total, status, created_at, deleted_at
		FROM orders
		WHERE id = $1
		  AND deleted_at IS NULL
		FOR UPDATE
	`, orderID)

	order, err := scanOrder(row)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("lock order for cancel: %w", err)
	}

	if err := order.CanCancel(userID); err != nil {
		return nil, err
	}

	items, err := r.loadOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	productRepo := NewProductRepository(r.db)
	for _, item := range order.Items {
		if err := productRepo.RestoreStock(ctx, item.ProductID, item.Quantity); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	order.Status = domain.OrderStatusCancelled
	order.DeletedAt = &now

	_, err = r.db.querier(ctx).Exec(ctx, `
		UPDATE orders
		SET status = $2,
		    deleted_at = $3
		WHERE id = $1
	`, order.ID, order.Status, order.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("update cancelled order: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) loadOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.db.querier(ctx).Query(ctx, `
		SELECT product_id, quantity, unit_price
		FROM order_items
		WHERE order_id = $1
		ORDER BY product_id
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("load order items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func scanOrder(row scannable) (*domain.Order, error) {
	var order domain.Order
	if err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Total,
		&order.Status,
		&order.CreatedAt,
		&order.DeletedAt,
	); err != nil {
		return nil, err
	}
	return &order, nil
}
