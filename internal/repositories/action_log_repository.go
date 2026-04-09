package repositories

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/db"
	"gorm.io/gorm"
)

type actionLogRepository struct {
	db *gorm.DB
}

func NewActionLogRepository(database *gorm.DB) ports.ActionLogRepository {
	return &actionLogRepository{db: database}
}

func (r *actionLogRepository) LogAction(log *domain.ActionLog) error {
	return r.db.Create(log).Error
}

func (r *actionLogRepository) GetLogsByTenant(tenantID string, limit, offset int) ([]domain.ActionLog, error) {
	var logs []domain.ActionLog
	err := db.TenantScope(tenantID)(r.db).Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, err
}

func (r *actionLogRepository) GetLogsByEntity(tenantID string, entityTable string, entityID string) ([]domain.ActionLog, error) {
	var logs []domain.ActionLog
	err := db.TenantScope(tenantID)(r.db).
		Where("entity_table = ? AND entity_id = ?", entityTable, entityID).
		Order("created_at desc").
		Find(&logs).Error
	return logs, err
}
