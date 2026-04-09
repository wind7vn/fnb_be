package ports

import "github.com/wind7vn/fnb_be/internal/core/domain"

type ProductRepository interface {
	Create(product *domain.Product) error
	FindAllByTenant(tenantID string, category *string) ([]domain.Product, error)
	FindByID(id string, tenantID string) (*domain.Product, error)
	Update(product *domain.Product) error
	Delete(id string, tenantID string) error
}
