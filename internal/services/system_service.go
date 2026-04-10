package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/config"
	"gorm.io/datatypes"
)

type SystemService struct {
	actionLogRepo ports.ActionLogRepository
	notiRepo      ports.NotificationRepository
	pushNoti      *NotificationService // injected Firebase FCM service
}

func NewSystemService(al ports.ActionLogRepository, no ports.NotificationRepository, push *NotificationService) *SystemService {
	return &SystemService{actionLogRepo: al, notiRepo: no, pushNoti: push}
}

// LogAction is meant to be called in a goroutine from other handlers or services.
func (s *SystemService) LogAction(tenantID string, userID string, role string, action string, entityTable string, entityID string, metadata interface{}) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return
	}

	uid, _ := uuid.Parse(userID) // Might be empty if sys/guest

	var puid *uuid.UUID
	if uid != uuid.Nil {
		puid = &uid
	}

	b, _ := json.Marshal(metadata)

	log := &domain.ActionLog{
		TenantID:    tid,
		UserID:      puid,
		Role:        role,
		Action:      action,
		EntityTable: entityTable,
		EntityID:    entityID,
		Metadata:    datatypes.JSON(b),
	}

	_ = s.actionLogRepo.LogAction(log)
}

func (s *SystemService) GetLogs(tenantID string, limit, offset int) ([]domain.ActionLog, *errors.AppError) {
	logs, err := s.actionLogRepo.GetLogsByTenant(tenantID, limit, offset)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return logs, nil
}

func (s *SystemService) GetUnreadNotifications(tenantID string, limit int) ([]domain.Notification, *errors.AppError) {
	notis, err := s.notiRepo.GetUnreadByTenant(tenantID, limit)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return notis, nil
}

func (s *SystemService) MarkRead(tenantID string, notiID string) *errors.AppError {
	err := s.notiRepo.MarkAsRead(tenantID, notiID)
	if err != nil {
		return errors.NewInternalServer(err)
	}
	return nil
}

func (s *SystemService) MarkAllRead(tenantID string) *errors.AppError {
	err := s.notiRepo.MarkAllAsRead(tenantID)
	if err != nil {
		return errors.NewInternalServer(err)
	}
	return nil
}

type BankMapping struct {
	ID          string `json:"id"`
	Bin         string `json:"bin"`
	ShortName   string `json:"short_name"`
	Name        string `json:"name"`
	BankLogoURL string `json:"logo_url"`
	IsBankQr    bool   `json:"is_bank_qr"`
}

var cachedBanks []BankMapping

func (s *SystemService) GetSupportedBanks() ([]BankMapping, *errors.AppError) {
	if len(cachedBanks) > 0 {
		return cachedBanks, nil
	}

	// Đọc tệp JSON gốc đã được pull về máy
	bodyBytes, err := os.ReadFile("./data/momo_banks.json")
	if err != nil {
		// Thử tìm ở thư mục cha nếu chạy lệnh từ thư mục con (ví dụ cmd_server)
		bodyBytes, err = os.ReadFile("../data/momo_banks.json")
		if err != nil {
			return nil, errors.NewInternalServer(fmt.Errorf("không thể tìm thấy file cấu hình ngân hàng: %v", err))
		}
	}

	var momoData map[string]struct {
		Bin         string `json:"bin"`
		ShortName   string `json:"shortName"`
		Name        string `json:"name"`
		BankLogoUrl string `json:"bankLogoUrl"`
		IsBankQr    bool   `json:"isBankQr"`
	}

	if err := json.Unmarshal(bodyBytes, &momoData); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	var results []BankMapping
	for id, bank := range momoData {
		results = append(results, BankMapping{
			ID:          id,
			Bin:         bank.Bin,
			ShortName:   bank.ShortName,
			Name:        bank.Name,
			BankLogoURL: config.AppConfig.AppDomain + bank.BankLogoUrl,
			IsBankQr:    bank.IsBankQr,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ShortName < results[j].ShortName
	})

	cachedBanks = results
	return results, nil
}

// CreateNotification saves to DB and asynchronously pushes via FCM or Redis
func (s *SystemService) CreateNotification(tenantID string, userID string, title, message, notiType string, data map[string]interface{}) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return
	}
	uid, _ := uuid.Parse(userID)
	
	b, _ := json.Marshal(data)

	noti := &domain.Notification{
		TenantID: tid,
		UserID:   uid,
		Title:    title,
		Body:     message,
		Type:     notiType,
		IsRead:   false,
		Data:     datatypes.JSON(b),
	}

	_ = s.notiRepo.Create(noti)

	// Stub out actual FCM logic for now. 
	// In production, fetch UserDevice tokens by UserID/TenantID, and use firebase.google.com/go/v4
	if s.pushNoti != nil {
		topicStr := "tenant_" + tenantID + "_staff"
		dataStr := make(map[string]string)
		for k, v := range data {
			if strVal, ok := v.(string); ok {
				dataStr[k] = strVal
			}
		}
		go s.pushNoti.SendTopicNotification(topicStr, title, message, dataStr)
	}
}
