package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"gorm.io/gorm"
)

type tenantMemberRepository struct {
	db *gorm.DB
}

func NewTenantMemberRepository(db *gorm.DB) *tenantMemberRepository {
	return &tenantMemberRepository{db: db}
}

func (r *tenantMemberRepository) Create(member *domain.TenantMember) error {
	return r.db.Create(member).Error
}

func (r *tenantMemberRepository) FindByUserAndTenant(userID string, tenantID string) (*domain.TenantMember, error) {
	var member domain.TenantMember
	err := r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *tenantMemberRepository) FindRolesByUser(userID string) ([]domain.TenantMember, error) {
	var members []domain.TenantMember
	err := r.db.Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

func (r *tenantMemberRepository) FindStaffByTenant(tenantID string) ([]domain.TenantMember, error) {
	var members []domain.TenantMember
	err := r.db.Where("tenant_id = ?", tenantID).Find(&members).Error
	return members, err
}
