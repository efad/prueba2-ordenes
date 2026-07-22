package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/efad/prueba2-ordenes/internal/domain"
)

func DataLoaderMiddleware(users domain.UserRepository, products domain.ProductRepository) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		loaders := NewLoaders(users, products)
		return next(WithLoaders(ctx, loaders))
	}
}
