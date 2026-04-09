package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"gorm.io/gorm"
)

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *tenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(tenant *domain.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *tenantRepository) FindAll() ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	err := r.db.Find(&tenants).Error
	return tenants, err
}

func (r *tenantRepository) FindByID(id string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.Where("id = ?", id).First(&tenant).Error
	return &tenant, err
}

func (r *tenantRepository) Update(tenant *domain.Tenant) error {
	return r.db.Save(tenant).Error
}
