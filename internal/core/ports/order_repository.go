package ports

import "github.com/wind7vn/fnb_be/internal/core/domain"

type OrderRepository interface {
	Create(order *domain.Order) error
	FindByID(id string, tenantID string) (*domain.Order, error)
	FindActiveByTable(tableID string, tenantID string) (*domain.Order, error)
	FindAllActive(tenantID string) ([]domain.Order, error)
	Update(order *domain.Order) error
	UpdateStatus(orderID string, status string) error
	UpdateItemStatus(itemID string, status string) error
}
