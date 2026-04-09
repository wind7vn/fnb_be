package domain

import "github.com/google/uuid"

// TenantMember maps a User to a Tenant with a specific Role (M:M Pivot Table)
type TenantMember struct {
	BaseModel
	UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_tenant" json:"user_id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_tenant" json:"tenant_id"`
	Role     string    `gorm:"type:varchar(50);not null" json:"role"` // e.g. Owner, Manager, Staff
}
