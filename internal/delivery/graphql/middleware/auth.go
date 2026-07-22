package middleware

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/efad/prueba2-ordenes/internal/delivery/graphql/ctxkey"
	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func Auth(tokens domain.TokenService) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx != nil && isPublicOperation(opCtx) {
			return next(ctx)
		}

		userID, err := userIDFromHeaders(ctx, tokens)
		if err != nil {
			addDomainError(ctx, err)
			return func(ctx context.Context) *graphql.Response {
				return &graphql.Response{
					Errors: graphql.GetErrors(ctx),
				}
			}
		}

		return next(ctxkey.WithUserID(ctx, userID))
	}
}

func isPublicOperation(opCtx *graphql.OperationContext) bool {
	if opCtx.Operation == nil {
		return false
	}

	switch opCtx.Operation.Operation {
	case ast.Query:
		return isIntrospectionQuery(opCtx.Operation)
	case ast.Mutation:
		return isAuthMutation(opCtx.Operation)
	default:
		return false
	}
}

func isIntrospectionQuery(op *ast.OperationDefinition) bool {
	for _, selection := range op.SelectionSet {
		field, ok := selection.(*ast.Field)
		if !ok {
			continue
		}
		if field.Name == "__schema" || field.Name == "__type" {
			return true
		}
	}
	return false
}

func isAuthMutation(op *ast.OperationDefinition) bool {
	for _, selection := range op.SelectionSet {
		field, ok := selection.(*ast.Field)
		if !ok {
			continue
		}
		if field.Name == "register" || field.Name == "login" {
			return true
		}
	}
	return false
}

func userIDFromHeaders(ctx context.Context, tokens domain.TokenService) (string, error) {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return "", domain.ErrUnauthorized
	}

	authHeader := opCtx.Headers.Get("Authorization")
	if authHeader == "" {
		return "", domain.ErrUnauthorized
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", domain.ErrUnauthorized
	}

	return tokens.Parse(strings.TrimSpace(parts[1]))
}

func addDomainError(ctx context.Context, err error) {
	if domainErr, ok := domain.AsDomainError(err); ok {
		graphql.AddError(ctx, &gqlerror.Error{
			Message: domainErr.Message,
			Extensions: map[string]any{
				"code": domainErr.Code,
			},
		})
		return
	}

	graphql.AddError(ctx, err)
}
