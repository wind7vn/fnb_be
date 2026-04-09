package ports

import "github.com/wind7vn/fnb_be/internal/core/domain"

type TenantRepository interface {
	Create(tenant *domain.Tenant) error
	FindAll() ([]domain.Tenant, error)
	FindByID(id string) (*domain.Tenant, error)
	Update(tenant *domain.Tenant) error
}
