package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides standard audit trails and soft-delete capabilities.
// This is embedded in all entity structs per architecture guidelines.
type BaseModel struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	ModifiedBy *uuid.UUID     `gorm:"type:uuid" json:"modified_by"`
	IsDeleted  bool           `gorm:"not null;default:false;index" json:"is_deleted"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM hook to natively inject UUIDv7 (Time-Ordered Sequential UUID)
func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		uid, err := uuid.NewV7()
		if err == nil {
			base.ID = uid
		}
	}
	return
}
