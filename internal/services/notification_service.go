package services

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
	"google.golang.org/api/option"
)

type NotificationService struct {
	client *messaging.Client
}

func NewNotificationService() *NotificationService {
	credentialPath := config.AppConfig.FirebaseServiceAccountPath
	if credentialPath == "" {
		// Fallback xuống relative config nếu thiếu env var
		credentialPath = "configs/fnb-2026-firebase-adminsdk-fbsvc-049b24c831.json"
	}

	if _, err := os.Stat(credentialPath); os.IsNotExist(err) {
		logger.Log.Warn(fmt.Sprintf("Firebase credentials not found at %s. Push notifications disabled.", credentialPath))
		return &NotificationService{client: nil}
	}

	opt := option.WithCredentialsFile(credentialPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("Failed to initialize Firebase app: %v", err))
		return &NotificationService{client: nil}
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		logger.Log.Error(fmt.Sprintf("Failed to initialize Firebase Messaging client: %v", err))
		return &NotificationService{client: nil}
	}

	logger.Log.Info("Firebase Messaging initialized successfully")
	return &NotificationService{client: client}
}

func (s *NotificationService) SendPush(topic string, payload domain.PushNotificationPayload) error {
	if s.client == nil {
		return nil
	}

	if payload.Data == nil {
		payload.Data = make(map[string]string)
	}
	payload.Data["type"] = payload.Type
	if payload.TargetID != "" {
		payload.Data["target_id"] = payload.TargetID
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Data:  payload.Data,
		Topic: topic,
	}

	_, err := s.client.Send(context.Background(), message)
	return err
}

// SendTopicNotification keeps backward compatibility
func (s *NotificationService) SendTopicNotification(topic, title, body string, rawData map[string]string) error {
	payload := domain.PushNotificationPayload{
		Type:  domain.NotiTypeSystem,
		Title: title,
		Body:  body,
		Data:  rawData,
	}
	return s.SendPush(topic, payload)
}

func (s *NotificationService) NotifyNewOrder(tableID, tableName string) error {
	payload := domain.PushNotificationPayload{
		Type:     domain.NotiTypeNewOrder,
		Title:    "Đơn hàng mới!",
		Body:     fmt.Sprintf("Bàn %s vừa tạo đơn hàng mới.", tableName),
		TargetID: tableID,
	}
	return s.SendPush(domain.TopicKitchen, payload)
}

func (s *NotificationService) NotifyCallStaff(tableID, tableName string) error {
	payload := domain.PushNotificationPayload{
		Type:     domain.NotiTypeCallStaff,
		Title:    "Khách gọi nhân viên!",
		Body:     fmt.Sprintf("Bàn %s yêu cầu hỗ trợ.", tableName),
		TargetID: tableID,
	}
	return s.SendPush(domain.TopicStaff, payload)
}
