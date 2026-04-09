package domain

import (
	"gorm.io/datatypes"
	"github.com/google/uuid"
)

type ActionLog struct {
	BaseModel
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID      *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"` // Ptr to allow sys/guest
	Role        string         `gorm:"type:varchar(50);not null" json:"role"`
	Action      string         `gorm:"type:varchar(100);not null" json:"action"`
	EntityTable string         `gorm:"type:varchar(100);not null" json:"entity_table"`
	EntityID    string         `gorm:"type:varchar(100);not null" json:"entity_id"` // Support any ID fmt
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata"` // Diff payload
}

type Notification struct {
	BaseModel
	TenantID uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title    string         `gorm:"type:varchar(255);not null" json:"title"`
	Body     string         `gorm:"type:text" json:"body"`
	Type     string         `gorm:"type:varchar(50);not null" json:"type"`
	IsRead   bool           `gorm:"not null;default:false" json:"is_read"`
	Data     datatypes.JSON `gorm:"type:jsonb" json:"data"`
}
