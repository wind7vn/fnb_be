package services

import (
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/wind7vn/fnb_be/pkg/cache"
)

type EventPayload struct {
	TenantID string      `json:"tenant_id"`
	Type     string      `json:"type"` // e.g., KDS_ITEM_UPDATED, TABLE_STATUS_CHANGED
	Data     interface{} `json:"data"`
}

type PubSubService struct{}

func NewPubSubService() *PubSubService {
	return &PubSubService{}
}

func (s *PubSubService) PublishEvent(channel string, payload EventPayload) error {
	if cache.RedisClient == nil {
		return nil // Graceful degrade if redis isn't present
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return cache.RedisClient.Publish(cache.Ctx, channel, bytes).Err()
}

func (s *PubSubService) Subscribe(channel string) *redis.PubSub {
	return cache.RedisClient.Subscribe(cache.Ctx, channel)
}
