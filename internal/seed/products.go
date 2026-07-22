package seed

import (
	"context"
	"fmt"

	"github.com/efad/prueba2-ordenes/internal/domain"
	"github.com/efad/prueba2-ordenes/internal/repository/postgres"
)

var defaultProducts = []domain.Product{
	{Name: "Teclado mecanico", Price: 89.99, Stock: 25},
	{Name: "Mouse inalambrico", Price: 39.50, Stock: 40},
	{Name: "Monitor 27 pulgadas", Price: 249.99, Stock: 12},
	{Name: "Webcam HD", Price: 59.00, Stock: 18},
	{Name: "Auriculares Bluetooth", Price: 74.25, Stock: 30},
	{Name: "Hub USB-C", Price: 29.99, Stock: 50},
	{Name: "Silla ergonomica", Price: 199.00, Stock: 8},
	{Name: "Escritorio ajustable", Price: 320.00, Stock: 5},
	{Name: "Lampara LED", Price: 24.99, Stock: 35},
	{Name: "Soporte para laptop", Price: 45.00, Stock: 22},
}

func Products(ctx context.Context, db *postgres.DB) error {
	var count int
	if err := db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count); err != nil {
		return fmt.Errorf("contar productos: %w", err)
	}
	if count > 0 {
		return nil
	}

	repo := postgres.NewProductRepository(db)
	for _, product := range defaultProducts {
		copyProduct := product
		if err := repo.Create(ctx, &copyProduct); err != nil {
			return fmt.Errorf("seed producto %s: %w", product.Name, err)
		}
	}

	return nil
}
