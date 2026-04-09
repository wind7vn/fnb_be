package domain

import (
	"github.com/google/uuid"
)

type Product struct {
	BaseModel
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Category    string    `gorm:"type:varchar(100);not null;index" json:"category"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	ImageURL    string    `gorm:"type:varchar(500)" json:"image_url"`
	IsAvailable bool      `gorm:"not null;default:true" json:"is_available"`
}
