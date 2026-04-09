package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/db"
	"gorm.io/gorm"
)

type orderRepository struct {
	dbConn *gorm.DB
}

func NewOrderRepository(dbConn *gorm.DB) *orderRepository {
	return &orderRepository{dbConn: dbConn}
}

func (r *orderRepository) Create(order *domain.Order) error {
	return r.dbConn.Create(order).Error
}

func (r *orderRepository) FindByID(id string, tenantID string) (*domain.Order, error) {
	var order domain.Order
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).Where("id = ?", id).Preload("Items").Preload("Items.Product").First(&order).Error
	return &order, err
}

func (r *orderRepository) FindActiveByTable(tableID string, tenantID string) (*domain.Order, error) {
	var order domain.Order
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).
		Where("table_id = ? AND status != 'Paid'", tableID).
		Preload("Items").
		Preload("Items.Product").
		First(&order).Error
	return &order, err
}

func (r *orderRepository) FindAllActive(tenantID string) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).
		Where("status != 'Paid'").
		Preload("Items").
		Preload("Items.Product").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Update(order *domain.Order) error {
	return r.dbConn.Session(&gorm.Session{FullSaveAssociations: true}).Save(order).Error
}

func (r *orderRepository) UpdateStatus(orderID string, status string) error {
	return r.dbConn.Model(&domain.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

func (r *orderRepository) UpdateItemStatus(itemID string, status string) error {
	return r.dbConn.Model(&domain.OrderItem{}).Where("id = ?", itemID).Update("status", status).Error
}
