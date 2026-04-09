package domain

// Tenant represents a Restaurant or an independent Store unit in our SaaS platform.
type Tenant struct {
	BaseModel
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Timezone string `gorm:"type:varchar(100);not null;default:'UTC'" json:"timezone"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`

	// Settings stored dynamically as JSONB (e.g. Printer IPs, Tax Configs)
	// Example: {"printer_ips": ["192.168.1.10"]}
	Metadata string `gorm:"type:jsonb" json:"metadata"`
}
