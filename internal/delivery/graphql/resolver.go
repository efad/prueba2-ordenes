package graphql

import "github.com/efad/prueba2-ordenes/internal/usecase"

type Resolver struct {
	AuthUC    *usecase.AuthUseCase
	ProductUC *usecase.ProductUseCase
}
