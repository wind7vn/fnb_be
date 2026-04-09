package domain

import (
	"github.com/google/uuid"
)

type Table struct {
	BaseModel
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name     string    `gorm:"type:varchar(100);not null" json:"name"`  // e.g., "T1", "VIP-1"
	Zone     string    `gorm:"type:varchar(100);default:'Tất cả'" json:"zone"` // e.g., "Tầng 1"
	Status   string    `gorm:"type:varchar(50);not null" json:"status"` // Available, Occupied
}
