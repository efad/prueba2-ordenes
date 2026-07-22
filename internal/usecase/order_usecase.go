package usecase

import (
	"context"

	"github.com/efad/prueba2-ordenes/internal/domain"
)

type OrderUseCase struct {
	orders   domain.OrderRepository
	products domain.ProductRepository
	tx       domain.TransactionManager
}

type OrderListResult struct {
	Items      []domain.Order
	TotalCount int
	Page       int
	PageSize   int
}

func NewOrderUseCase(
	orders domain.OrderRepository,
	products domain.ProductRepository,
	tx domain.TransactionManager,
) *OrderUseCase {
	return &OrderUseCase{
		orders:   orders,
		products: products,
		tx:       tx,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID string, items []domain.CreateOrderItemInput) (*domain.Order, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	if err := domain.ValidateCreateOrderItems(items); err != nil {
		return nil, err
	}

	order := &domain.Order{
		UserID: userID,
		Status: domain.OrderStatusPending,
	}

	err := uc.tx.RunInTransaction(ctx, func(txCtx context.Context) error {
		order.Items = make([]domain.OrderItem, 0, len(items))

		for _, input := range items {
			product, err := uc.products.FindByID(txCtx, input.ProductID)
			if err != nil {
				return err
			}
			if product.Stock < input.Quantity {
				return domain.ErrInsufficientStock
			}

			if err := uc.products.DecrementStock(txCtx, input.ProductID, input.Quantity); err != nil {
				return err
			}

			order.Items = append(order.Items, domain.OrderItem{
				ProductID: input.ProductID,
				Quantity:  input.Quantity,
				UnitPrice: product.Price,
			})
		}

		order.Total = domain.CalculateOrderTotal(order.Items)
		return uc.orders.Create(txCtx, order)
	})
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) MyOrders(ctx context.Context, userID string, page, pageSize int) (*OrderListResult, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	page, pageSize = domain.NormalizePagination(page, pageSize)
	orders, total, err := uc.orders.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &OrderListResult{
		Items:      orders,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, userID, orderID string) (*domain.Order, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	order, err := uc.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if err := order.CanView(userID); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, userID, orderID string) (*domain.Order, error) {
	if userID == "" {
		return nil, domain.ErrUnauthorized
	}

	return uc.orders.Cancel(ctx, orderID, userID)
}
