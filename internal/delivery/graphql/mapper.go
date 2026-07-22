package graphql

import (
	"context"
	"time"

	"github.com/efad/prueba2-ordenes/internal/delivery/graphql/ctxkey"
	"github.com/efad/prueba2-ordenes/internal/delivery/graphql/model"
	"github.com/efad/prueba2-ordenes/internal/domain"
)

func requireUserID(ctx context.Context) (string, error) {
	userID, ok := ctxkey.UserID(ctx)
	if !ok {
		return "", domain.ErrUnauthorized
	}
	return userID, nil
}

func toGraphQLUser(user domain.User) *model.User {
	return &model.User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toGraphQLOrder(order domain.Order) *model.Order {
	items := make([]*model.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &model.OrderItem{
			Product:   &model.Product{ID: item.ProductID},
			Quantity:  int32(item.Quantity),
			UnitPrice: item.UnitPrice,
		})
	}

	return &model.Order{
		ID:        order.ID,
		User:      &model.User{ID: order.UserID},
		Items:     items,
		Total:     order.Total,
		Status:    model.OrderStatus(order.Status),
		CreatedAt: order.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toGraphQLOrders(orders []domain.Order) []*model.Order {
	items := make([]*model.Order, 0, len(orders))
	for _, order := range orders {
		copyOrder := order
		items = append(items, toGraphQLOrder(copyOrder))
	}
	return items
}

func toDomainCreateOrderItems(items []*model.CreateOrderItemInput) []domain.CreateOrderItemInput {
	result := make([]domain.CreateOrderItemInput, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, domain.CreateOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  int(item.Quantity),
		})
	}
	return result
}

func toGraphQLProduct(product domain.Product) *model.Product {
	return &model.Product{
		ID:    product.ID,
		Name:  product.Name,
		Price: product.Price,
		Stock: int32(product.Stock),
	}
}

func toGraphQLProducts(products []domain.Product) []*model.Product {
	items := make([]*model.Product, 0, len(products))
	for _, product := range products {
		items = append(items, toGraphQLProduct(product))
	}
	return items
}

func toDomainProductFilter(filter *model.ProductFilter) domain.ProductFilter {
	if filter == nil {
		return domain.ProductFilter{}
	}

	return domain.ProductFilter{
		Name:     filter.Name,
		MinPrice: filter.MinPrice,
		MaxPrice: filter.MaxPrice,
	}
}

func int32Value(value *int32, fallback int) int {
	if value == nil {
		return fallback
	}
	return int(*value)
}
