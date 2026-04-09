package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByPhoneAndTenant(phone string, tenantID *string) (*domain.User, error) {
	var user domain.User
	query := r.db.Where("phone_number = ?", phone)

	// If tenantID is provided, filter by it. Otherwise, assume global search (e.g. initial login)
	if tenantID != nil && *tenantID != "" {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	err := query.First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) SaveDeviceToken(device *domain.UserDevice) error {
	// Upsert Logic based on DeviceID and UserID
	return r.db.Where(domain.UserDevice{UserID: device.UserID, DeviceID: device.DeviceID}).
		Assign(domain.UserDevice{FCMToken: device.FCMToken, Platform: device.Platform, LastActive: device.LastActive}).
		FirstOrCreate(device).Error
}

func (r *userRepository) FindStaffByTenant(tenantID string) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

func (r *userRepository) FindAllByPhone(phone string) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Where("phone_number = ?", phone).Find(&users).Error
	return users, err
}
