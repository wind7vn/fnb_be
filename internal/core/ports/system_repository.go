package ports

import "github.com/wind7vn/fnb_be/internal/core/domain"

type ActionLogRepository interface {
	LogAction(log *domain.ActionLog) error
	GetLogsByTenant(tenantID string, limit, offset int) ([]domain.ActionLog, error)
	GetLogsByEntity(tenantID string, entityTable string, entityID string) ([]domain.ActionLog, error)
}

type NotificationRepository interface {
	Create(notification *domain.Notification) error
	GetUnreadByTenant(tenantID string, limit int) ([]domain.Notification, error)
	MarkAsRead(tenantID string, notificationID string) error
	MarkAllAsRead(tenantID string) error
}
