package usecase

import (
	"context"

	"github.com/efad/prueba2-ordenes/internal/domain"
)

type ProductUseCase struct {
	products domain.ProductRepository
}

type ProductListResult struct {
	Items      []domain.Product
	TotalCount int
	Page       int
	PageSize   int
}

func NewProductUseCase(products domain.ProductRepository) *ProductUseCase {
	return &ProductUseCase{products: products}
}

func (uc *ProductUseCase) List(
	ctx context.Context,
	filter domain.ProductFilter,
	page, pageSize int,
) (*ProductListResult, error) {
	if err := domain.ValidateProductFilter(filter); err != nil {
		return nil, err
	}

	page, pageSize = domain.NormalizePagination(page, pageSize)

	items, total, err := uc.products.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &ProductListResult{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (uc *ProductUseCase) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, domain.ErrProductNotFound
	}

	return uc.products.FindByID(ctx, id)
}
