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
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).Where("id = ?", id).
		Preload("Items").
		Preload("Items.Product").
		Preload("Table").
		First(&order).Error
	return &order, err
}

func (r *orderRepository) FindActiveByTable(tableID string, tenantID string) (*domain.Order, error) {
	var order domain.Order
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).
		Where("table_id = ? AND status NOT IN ('Paid', 'Cancelled')", tableID).
		Preload("Items").
		Preload("Items.Product").
		Preload("Table").
		First(&order).Error
	return &order, err
}

func (r *orderRepository) FindAllActive(tenantID string) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.dbConn.Scopes(db.TenantScope(tenantID)).
		Where("status NOT IN ('Paid', 'Cancelled')").
		Preload("Items").
		Preload("Items.Product").
		Preload("Table").
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

func (r *orderRepository) DeleteItem(itemID string, tenantID string) error {
	return r.dbConn.Where("id = ? AND tenant_id = ?", itemID, tenantID).Delete(&domain.OrderItem{}).Error
}

func (r *orderRepository) UpdateItemQuantity(itemID string, tenantID string, quantity int, newSubTotal float64) error {
	return r.dbConn.Model(&domain.OrderItem{}).Where("id = ? AND tenant_id = ?", itemID, tenantID).Updates(map[string]interface{}{
		"quantity":  quantity,
		"sub_total": newSubTotal,
	}).Error
}

func (r *orderRepository) FindItemByID(itemID string) (*domain.OrderItem, error) {
	var item domain.OrderItem
	err := r.dbConn.Where("id = ?", itemID).
		Preload("Product").
		First(&item).Error
	return &item, err
}

