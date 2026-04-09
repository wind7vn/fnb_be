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

func (r *userRepository) FindByPhone(phone string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("phone_number = ?", phone).First(&user).Error
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
