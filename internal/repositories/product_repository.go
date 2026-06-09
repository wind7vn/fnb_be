package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/db"
	"gorm.io/gorm"
)

type productRepository struct {
	dbConn *gorm.DB
}

func NewProductRepository(dbConn *gorm.DB) *productRepository {
	return &productRepository{dbConn: dbConn}
}

func (r *productRepository) Create(product *domain.Product) error {
	return r.dbConn.Create(product).Error
}

func (r *productRepository) FindAllByTenant(tenantID string, category *string) ([]domain.Product, error) {
	var products []domain.Product
	query := r.dbConn.Scopes(db.TenantScope(tenantID))
	if category != nil && *category != "" {
		query = query.Where("category = ?", *category)
	}
	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) FindByID(id string, tenantID string) (*domain.Product, error) {
	var product domain.Product
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).Where("id = ?", id).First(&product).Error
	return &product, err
}

func (r *productRepository) Update(product *domain.Product) error {
	return r.dbConn.Save(product).Error
}

func (r *productRepository) Delete(id string, tenantID string) error {
	return r.dbConn.Scopes(db.TenantScope(tenantID)).Where("id = ?", id).Delete(&domain.Product{}).Error
}

func (r *productRepository) BatchUpdateStatus(tenantID string, updates map[string]bool) error {
	var trueIDs []string
	var falseIDs []string
	for id, status := range updates {
		if status {
			trueIDs = append(trueIDs, id)
		} else {
			falseIDs = append(falseIDs, id)
		}
	}

	return r.dbConn.Transaction(func(tx *gorm.DB) error {
		if len(trueIDs) > 0 {
			if e := tx.Scopes(db.TenantScope(tenantID)).Where("id IN ?", trueIDs).Model(&domain.Product{}).Update("is_available", true).Error; e != nil {
				return e
			}
		}
		if len(falseIDs) > 0 {
			if e := tx.Scopes(db.TenantScope(tenantID)).Where("id IN ?", falseIDs).Model(&domain.Product{}).Update("is_available", false).Error; e != nil {
				return e
			}
		}
		return nil
	})
}
