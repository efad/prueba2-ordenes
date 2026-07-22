package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	if domainErr, ok := domain.AsDomainError(err); ok {
		return &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: domainErr.Message,
			Extensions: map[string]any{
				"code": domainErr.Code,
			},
		}
	}

	return graphql.DefaultErrorPresenter(ctx, err)
}
