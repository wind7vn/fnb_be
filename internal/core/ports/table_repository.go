package ports

import "github.com/wind7vn/fnb_be/internal/core/domain"

type TableRepository interface {
	Create(table *domain.Table) error
	FindAllByTenant(tenantID string) ([]domain.Table, error)
	FindByID(id string, tenantID string) (*domain.Table, error)
	Update(table *domain.Table) error
	Delete(id string, tenantID string) error
}
