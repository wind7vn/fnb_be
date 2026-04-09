package ports

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
)

// UserRepository handles user persistence.
type UserRepository interface {
	FindByPhoneAndTenant(phone string, tenantID *string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
	SaveDeviceToken(device *domain.UserDevice) error
	FindStaffByTenant(tenantID string) ([]domain.User, error)
	FindAllByPhone(phone string) ([]domain.User, error)
}
