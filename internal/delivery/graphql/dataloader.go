package graphql

import (
	"context"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/graph-gophers/dataloader/v7"
)

type loadersKey struct{}

type Loaders struct {
	User    *dataloader.Loader[string, *domain.User]
	Product *dataloader.Loader[string, *domain.Product]
}

func NewLoaders(users domain.UserRepository, products domain.ProductRepository) *Loaders {
	return &Loaders{
		User: dataloader.NewBatchedLoader(func(ctx context.Context, keys []string) []*dataloader.Result[*domain.User] {
			return batchUsers(ctx, users, keys)
		}),
		Product: dataloader.NewBatchedLoader(func(ctx context.Context, keys []string) []*dataloader.Result[*domain.Product] {
			return batchProducts(ctx, products, keys)
		}),
	}
}

func batchUsers(ctx context.Context, users domain.UserRepository, keys []string) []*dataloader.Result[*domain.User] {
	results := make([]*dataloader.Result[*domain.User], len(keys))
	found, err := users.FindByIDs(ctx, keys)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*domain.User]{Error: err}
		}
		return results
	}

	byID := make(map[string]*domain.User, len(found))
	for i := range found {
		user := found[i]
		copyUser := user
		byID[user.ID] = &copyUser
	}

	for i, key := range keys {
		if user, ok := byID[key]; ok {
			results[i] = &dataloader.Result[*domain.User]{Data: user}
			continue
		}
		results[i] = &dataloader.Result[*domain.User]{Error: domain.ErrUnauthorized}
	}

	return results
}

func batchProducts(ctx context.Context, products domain.ProductRepository, keys []string) []*dataloader.Result[*domain.Product] {
	results := make([]*dataloader.Result[*domain.Product], len(keys))
	found, err := products.FindByIDs(ctx, keys)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result[*domain.Product]{Error: err}
		}
		return results
	}

	byID := make(map[string]*domain.Product, len(found))
	for i := range found {
		product := found[i]
		copyProduct := product
		byID[product.ID] = &copyProduct
	}

	for i, key := range keys {
		if product, ok := byID[key]; ok {
			results[i] = &dataloader.Result[*domain.Product]{Data: product}
			continue
		}
		results[i] = &dataloader.Result[*domain.Product]{Error: domain.ErrProductNotFound}
	}

	return results
}

func WithLoaders(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey{}, loaders)
}

func LoadersFromContext(ctx context.Context) *Loaders {
	loaders, _ := ctx.Value(loadersKey{}).(*Loaders)
	return loaders
}
