package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByIDs(ctx context.Context, ids []string) ([]User, error)
}

type ProductRepository interface {
	List(ctx context.Context, filter ProductFilter, page, pageSize int) ([]Product, int, error)
	FindByID(ctx context.Context, id string) (*Product, error)
	FindByIDs(ctx context.Context, ids []string) ([]Product, error)
	DecrementStock(ctx context.Context, productID string, quantity int) error
	RestoreStock(ctx context.Context, productID string, quantity int) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]Order, int, error)
	Cancel(ctx context.Context, orderID, userID string) (*Order, error)
}

type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TokenService interface {
	Generate(userID string) (string, error)
	Parse(token string) (userID string, err error)
}
