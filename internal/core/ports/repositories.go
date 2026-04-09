package ports

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
)

// UserRepository handles user persistence.
type UserRepository interface {
	FindByPhone(phone string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
	SaveDeviceToken(device *domain.UserDevice) error
}

type TenantMemberRepository interface {
	Create(member *domain.TenantMember) error
	FindByUserAndTenant(userID string, tenantID string) (*domain.TenantMember, error)
	FindRolesByUser(userID string) ([]domain.TenantMember, error)
	FindStaffByTenant(tenantID string) ([]domain.TenantMember, error)
}
