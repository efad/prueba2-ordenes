package graphql

import (
	"github.com/efad/prueba2-ordenes/internal/delivery/graphql/model"
	"github.com/efad/prueba2-ordenes/internal/domain"
)

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
