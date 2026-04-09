package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/db"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(database *gorm.DB) ports.NotificationRepository {
	return &notificationRepository{db: database}
}

func (r *notificationRepository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetUnreadByTenant(tenantID string, limit int) ([]domain.Notification, error) {
	var results []domain.Notification
	err := db.TenantScope(tenantID)(r.db).
		Where("is_read = ?", false).
		Order("created_at desc").
		Limit(limit).
		Find(&results).Error
	return results, err
}

func (r *notificationRepository) MarkAsRead(tenantID string, notificationID string) error {
	return db.TenantScope(tenantID)(r.db).
		Model(&domain.Notification{}).
		Where("id = ?", notificationID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(tenantID string) error {
	return db.TenantScope(tenantID)(r.db).
		Model(&domain.Notification{}).
		Where("is_read = ?", false).
		Update("is_read", true).Error
}
