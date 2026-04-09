package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SystemRole definitions for Global Permissions
type SystemRole string

const (
	SystemRoleUser      SystemRole = "user"
	SystemRoleAdmin     SystemRole = "admin"
	SystemRoleSuperuser SystemRole = "superuser"
)

// User maps to employee mappings (Admin, Owner, Manager, Staff)
type User struct {
	BaseModel
	PhoneNumber  string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"phone_number"` // Login ID globally unique
	FullName     string         `gorm:"type:varchar(255);not null" json:"full_name"`
	AvatarURL    string         `gorm:"type:varchar(500)" json:"avatar_url"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // Omit on output
	SystemRole   SystemRole     `gorm:"type:varchar(20);not null;default:'user'" json:"system_role"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
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
