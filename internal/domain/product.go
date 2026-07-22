package domain

type Product struct {
	ID    string
	Name  string
	Price float64
	Stock int
}

type ProductFilter struct {
	Name     *string
	MinPrice *float64
	MaxPrice *float64
}

func ValidateProductFilter(filter ProductFilter) error {
	if filter.MinPrice != nil && *filter.MinPrice < 0 {
		return invalidInput("el precio minimo no puede ser negativo")
	}
	if filter.MaxPrice != nil && *filter.MaxPrice < 0 {
		return invalidInput("el precio maximo no puede ser negativo")
	}
	if filter.MinPrice != nil && filter.MaxPrice != nil && *filter.MinPrice > *filter.MaxPrice {
		return invalidInput("el precio minimo no puede ser mayor al maximo")
	}
	return nil
}

func NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
