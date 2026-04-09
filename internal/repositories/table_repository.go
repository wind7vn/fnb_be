package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/db"
	"gorm.io/gorm"
)

type tableRepository struct {
	dbConn *gorm.DB
}

func NewTableRepository(dbConn *gorm.DB) *tableRepository {
	return &tableRepository{dbConn: dbConn}
}

func (r *tableRepository) Create(table *domain.Table) error {
	return r.dbConn.Create(table).Error
}

func (r *tableRepository) FindAllByTenant(tenantID string) ([]domain.Table, error) {
	var tables []domain.Table
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).Find(&tables).Error
	return tables, err
}

func (r *tableRepository) FindByID(id string, tenantID string) (*domain.Table, error) {
	var table domain.Table
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).Where("id = ?", id).First(&table).Error
	return &table, err
}

func (r *tableRepository) Update(table *domain.Table) error {
	return r.dbConn.Save(table).Error
}
