package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/google/uuid"
)

type ProductRepository struct {
	db *DB
}

func NewProductRepository(db *DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, filter domain.ProductFilter, page, pageSize int) ([]domain.Product, int, error) {
	page, pageSize = domain.NormalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	where, args := buildProductFilter(filter)
	countQuery := "SELECT COUNT(*) FROM products" + where
	var total int
	if err := r.db.querier(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	listArgs := append(append([]any{}, args...), pageSize, offset)
	listQuery := `
		SELECT id, name, price, stock
		FROM products
	` + where + `
		ORDER BY name ASC
		LIMIT $` + fmt.Sprint(len(args)+1) + `
		OFFSET $` + fmt.Sprint(len(args)+2)

	rows, err := r.db.querier(ctx).Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, *product)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func buildProductFilter(filter domain.ProductFilter) (string, []any) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if filter.Name != nil && strings.TrimSpace(*filter.Name) != "" {
		args = append(args, "%"+strings.TrimSpace(*filter.Name)+"%")
		clauses = append(clauses, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	if filter.MinPrice != nil {
		args = append(args, *filter.MinPrice)
		clauses = append(clauses, fmt.Sprintf("price >= $%d", len(args)))
	}
	if filter.MaxPrice != nil {
		args = append(args, *filter.MaxPrice)
		clauses = append(clauses, fmt.Sprintf("price <= $%d", len(args)))
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	row := r.db.querier(ctx).QueryRow(ctx, `
		SELECT id, name, price, stock
		FROM products
		WHERE id = $1
	`, id)

	product, err := scanProduct(row)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("find product by id: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) FindByIDs(ctx context.Context, ids []string) ([]domain.Product, error) {
	if len(ids) == 0 {
		return []domain.Product{}, nil
	}

	rows, err := r.db.querier(ctx).Query(ctx, `
		SELECT id, name, price, stock
		FROM products
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("find products by ids: %w", err)
	}
	defer rows.Close()

	productsByID := make(map[string]domain.Product, len(ids))
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		productsByID[product.ID] = *product
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	products := make([]domain.Product, 0, len(ids))
	for _, id := range ids {
		if product, ok := productsByID[id]; ok {
			products = append(products, product)
		}
	}

	return products, nil
}

func (r *ProductRepository) DecrementStock(ctx context.Context, productID string, quantity int) error {
	tag, err := r.db.querier(ctx).Exec(ctx, `
		UPDATE products
		SET stock = stock - $2
		WHERE id = $1
		  AND stock >= $2
	`, productID, quantity)
	if err != nil {
		return fmt.Errorf("decrement stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		product, findErr := r.FindByID(ctx, productID)
		if findErr != nil {
			return findErr
		}
		if product.Stock < quantity {
			return domain.ErrInsufficientStock
		}
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) RestoreStock(ctx context.Context, productID string, quantity int) error {
	tag, err := r.db.querier(ctx).Exec(ctx, `
		UPDATE products
		SET stock = stock + $2
		WHERE id = $1
	`, productID, quantity)
	if err != nil {
		return fmt.Errorf("restore stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	if product.ID == "" {
		product.ID = uuid.NewString()
	}

	_, err := r.db.querier(ctx).Exec(ctx, `
		INSERT INTO products (id, name, price, stock)
		VALUES ($1, $2, $3, $4)
	`, product.ID, product.Name, product.Price, product.Stock)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	return nil
}

func scanProduct(row scannable) (*domain.Product, error) {
	var product domain.Product
	if err := row.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
		return nil, err
	}
	return &product, nil
}
