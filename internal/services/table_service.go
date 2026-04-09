package services

import (
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"gorm.io/gorm"
)

type TableService struct {
	repo ports.TableRepository
}

func NewTableService(r ports.TableRepository) *TableService {
	return &TableService{repo: r}
}

type TableRequest struct {
	Name   string `json:"name"`
	Zone   string `json:"zone"`
	Status string `json:"status"` // Available, Occupied, Reserved
}

func (s *TableService) Create(tenantID string, req TableRequest) (*domain.Table, *errors.AppError) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Tenant ID invalid", err)
	}

	table := &domain.Table{
		TenantID: tid,
		Name:     req.Name,
		Zone:     req.Zone,
		Status:   req.Status,
	}

	if err := s.repo.Create(table); err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return table, nil
}

func (s *TableService) GetAll(tenantID string) ([]domain.Table, *errors.AppError) {
	tables, err := s.repo.FindAllByTenant(tenantID)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return tables, nil
}

func (s *TableService) UpdateStatus(tenantID string, id string, status string) (*domain.Table, *errors.AppError) {
	table, err := s.repo.FindByID(id, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Bàn không tồn tại", err)
		}
		return nil, errors.NewInternalServer(err)
	}

	table.Status = status
	if err := s.repo.Update(table); err != nil {
		return nil, errors.NewInternalServer(err)
	}
	
	// TODO: Emit Redis Pub/Sub WebSocket event "TABLE_STATUS_CHANGED" here in Phase 9
	return table, nil
}
