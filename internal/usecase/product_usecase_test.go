package usecase_test

import (
	"context"
	"testing"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/efad/prueba2-ordenes/internal/usecase"
)

type mockProductRepository struct {
	products []domain.Product
}

func (m *mockProductRepository) List(_ context.Context, filter domain.ProductFilter, page, pageSize int) ([]domain.Product, int, error) {
	filtered := make([]domain.Product, 0)
	for _, product := range m.products {
		if filter.Name != nil && *filter.Name != "" {
			if !containsIgnoreCase(product.Name, *filter.Name) {
				continue
			}
		}
		if filter.MinPrice != nil && product.Price < *filter.MinPrice {
			continue
		}
		if filter.MaxPrice != nil && product.Price > *filter.MaxPrice {
			continue
		}
		filtered = append(filtered, product)
	}

	page, pageSize = domain.NormalizePagination(page, pageSize)
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []domain.Product{}, len(filtered), nil
	}

	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], len(filtered), nil
}

func (m *mockProductRepository) FindByID(_ context.Context, id string) (*domain.Product, error) {
	for _, product := range m.products {
		if product.ID == id {
			copyProduct := product
			return &copyProduct, nil
		}
	}
	return nil, domain.ErrProductNotFound
}

func (m *mockProductRepository) FindByIDs(_ context.Context, ids []string) ([]domain.Product, error) {
	return []domain.Product{}, nil
}

func (m *mockProductRepository) DecrementStock(_ context.Context, _ string, _ int) error {
	return nil
}

func (m *mockProductRepository) RestoreStock(_ context.Context, _ string, _ int) error {
	return nil
}

func containsIgnoreCase(value, substr string) bool {
	return len(value) >= len(substr) && (value == substr || len(substr) == 0)
}

func TestProductUseCaseListWithPagination(t *testing.T) {
	t.Parallel()

	repo := &mockProductRepository{
		products: []domain.Product{
			{ID: "1", Name: "Teclado", Price: 50, Stock: 10},
			{ID: "2", Name: "Mouse", Price: 25, Stock: 5},
			{ID: "3", Name: "Monitor", Price: 200, Stock: 2},
		},
	}
	productUC := usecase.NewProductUseCase(repo)

	result, err := productUC.List(context.Background(), domain.ProductFilter{}, 1, 2)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(result.Items))
	}
	if result.TotalCount != 3 {
		t.Fatalf("totalCount = %d, want 3", result.TotalCount)
	}
}

func TestProductUseCaseGetByIDNotFound(t *testing.T) {
	t.Parallel()

	productUC := usecase.NewProductUseCase(&mockProductRepository{})
	_, err := productUC.GetByID(context.Background(), "missing")
	if err != domain.ErrProductNotFound {
		t.Fatalf("GetByID() error = %v, want ErrProductNotFound", err)
	}
}
