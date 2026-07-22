package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

type Order struct {
	ID        string
	UserID    string
	Items     []OrderItem
	Total     float64
	Status    OrderStatus
	CreatedAt time.Time
	DeletedAt *time.Time
}

type CreateOrderItemInput struct {
	ProductID string
	Quantity  int
}

func ValidateCreateOrderItems(items []CreateOrderItemInput) error {
	if len(items) == 0 {
		return invalidInput("la orden debe tener al menos un item")
	}

	seenProducts := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.ProductID == "" {
			return invalidInput("el producto es obligatorio")
		}
		if item.Quantity <= 0 {
			return invalidInput("la cantidad debe ser mayor a cero")
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return invalidInput("no se permiten productos duplicados en la misma orden")
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	return nil
}

func CalculateOrderTotal(items []OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total
}

func (o *Order) CanCancel(userID string) error {
	if o.UserID != userID {
		return ErrOrderNotOwned
	}
	if o.Status != OrderStatusPending {
		return ErrOrderNotCancellable
	}
	return nil
}

func (o *Order) CanView(userID string) error {
	if o.UserID != userID {
		return ErrOrderNotOwned
	}
	return nil
}
