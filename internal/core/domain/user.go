package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// User maps to employee mappings (Admin, Owner, Manager, Staff)
type User struct {
	BaseModel
	TenantID     *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"` // Nullable for Superadmin
	Role         string     `gorm:"type:varchar(50);not null" json:"role"`
	PhoneNumber  string     `gorm:"type:varchar(20);not null;index" json:"phone_number"` // Login ID
	FullName     string     `gorm:"type:varchar(255);not null" json:"full_name"`
	AvatarURL    string     `gorm:"type:varchar(500)" json:"avatar_url"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"` // Omit on output
	Metadata     datatypes.JSON     `gorm:"type:jsonb" json:"metadata"`

	// Associations constraint: A phone number should be unique INSIDE a specific Tenant.
	// That constraint should be implemented in DB Migration via unique composite index (tenant_id, phone_number).
}

// UserDevice tracks push notification tokens (FCM/APNS) for the employee's active device
type UserDevice struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceID   string    `gorm:"type:varchar(255);not null" json:"device_id"`
	FCMToken   string    `gorm:"type:varchar(500);not null" json:"fcm_token"`
	Platform   string    `gorm:"type:varchar(20);not null" json:"platform"` // ios, android, web
	LastActive time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"last_active"`
}
