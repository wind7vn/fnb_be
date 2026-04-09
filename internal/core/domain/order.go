package domain

import (
	"github.com/google/uuid"
)

type Order struct {
	BaseModel
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	TableID    *uuid.UUID `gorm:"type:uuid;index" json:"table_id"` // Nullable for takeaway
	Code       string     `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Status     string     `gorm:"type:varchar(50);not null" json:"status"` // Enum mapping needed
	TotalPrice float64     `gorm:"type:decimal(12,2);not null;default:0" json:"total_price"`
	SessionID  string     `gorm:"type:varchar(255)" json:"session_id"` // Identifies the guest session/device
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	BaseModel
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Status    string    `gorm:"type:varchar(50);not null" json:"status"` // KDS Status: Pending, Cooking, Ready, Served
	Note      string    `gorm:"type:text" json:"note"`
	SubTotal  float64   `gorm:"type:decimal(10,2);not null" json:"sub_total"`
}
