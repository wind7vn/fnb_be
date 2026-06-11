package repositories

import (
	"time"

	"github.com/google/uuid"
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
	// 1. Deactivate other users' device tokens on the same device_id OR fcm_token
	now := time.Now()
	err := r.db.Model(&domain.UserDevice{}).
		Where("(device_id = ? OR fcm_token = ?) AND user_id != ? AND is_deleted = false", device.DeviceID, device.FCMToken, device.UserID).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": gorm.DeletedAt{Time: now, Valid: true},
		}).Error
	if err != nil {
		return err
	}

	// 2. Upsert/Reactivate the current user's device token
	return r.db.Unscoped().
		Where(domain.UserDevice{UserID: device.UserID, DeviceID: device.DeviceID}).
		Assign(domain.UserDevice{
			FCMToken:   device.FCMToken,
			Platform:   device.Platform,
			LastActive: device.LastActive,
			BaseModel: domain.BaseModel{
				IsDeleted: false,
				DeletedAt: gorm.DeletedAt{Valid: false},
			},
		}).
		FirstOrCreate(device).Error
}

func (r *userRepository) DeleteDeviceToken(userID string, deviceID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	// Step 1: Update is_deleted to true
	err = r.db.Model(&domain.UserDevice{}).
		Where("user_id = ? AND device_id = ?", uid, deviceID).
		Update("is_deleted", true).Error
	if err != nil {
		return err
	}
	// Step 2: Soft delete the record
	return r.db.Where("user_id = ? AND device_id = ?", uid, deviceID).
		Delete(&domain.UserDevice{}).Error
}

func (r *userRepository) FindDeviceTokensByUserID(userID string) ([]domain.UserDevice, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	var devices []domain.UserDevice
	err = r.db.Where("user_id = ? AND is_deleted = false", uid).Find(&devices).Error
	return devices, err
}
