package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	TenantID    string `json:"tenant_id,omitempty"`
}

type LoginResponse struct {
	Data struct {
		Token string `json:"token"`
		User  struct {
			TenantID string `json:"tenant_id"`
		} `json:"user"`
	} `json:"data"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1"

	// 1. Login to get token
	log.Println("1. Logging in...")
	loginReq := LoginRequest{
		PhoneNumber: "0888888888",
		Password:    "123456",
		TenantID:    "11111111-1111-1111-1111-111111111111",
	}
	body, _ := json.Marshal(loginReq)
	
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Login failed with HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
		log.Fatalf("Failed to decode token: %v\nBody:%s", err, string(bodyBytes))
	}

	token := loginResp.Data.Token
	tenantID := "11111111-1111-1111-1111-111111111111"

	if token == "" {
		log.Fatal("Could not get JWT token from response:", string(bodyBytes))
	}
	
	log.Printf("Logged in. Tenant: %s", tenantID)

	// 2. Connect to WebSocket
	wsURL := fmt.Sprintf("ws://localhost:8080/api/v1/ws/kds?token=%s", token)
	log.Printf("2. Connecting to WebSocket: %s", wsURL)

	conn, res, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if res != nil {
			body, _ := io.ReadAll(res.Body)
			log.Fatalf("WebSocket connection failed: %v. Status: %d, Body: %s", err, res.StatusCode, string(body))
		}
		log.Fatalf("WebSocket connection failed: %v", err)
	}
	defer conn.Close()

	// Goroutine to read messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("WS read error:", err)
				return
			}
			log.Printf("\n[✅ WS MESSAGE RECEIVED] %s\n", message)
		}
	}()

	log.Println("3. Connected to WS! Simulating Redis event after 2 seconds...")

	time.Sleep(2 * time.Second)

	// Publish to Redis
	event := map[string]interface{}{
		"tenant_id": tenantID,
		"type":      "ITEM_STATUS_UPDATED",
		"data": map[string]string{
			"item_id": "test-item-1234",
			"status":  "Ready",
		},
	}
	eventJSON, _ := json.Marshal(event)
	
	// Create temporary redis client
	importRedis(tenantID, eventJSON) // I will move this below main

	time.Sleep(2 * time.Second)
	log.Println("Test finished successfully.")
}

func importRedis(tenantID string, eventJSON []byte) {
	log.Printf("Publishing event to Redis: %s", string(eventJSON))

	opt, _ := redis.ParseURL("redis://:WsA8J6PPuf8N6ysVlmNB@localhost:6379/0")
	client := redis.NewClient(opt)
	defer client.Close()

	ctx := context.Background()
	channel := "KDS_EVENTS"
	
	err := client.Publish(ctx, channel, string(eventJSON)).Err()
	if err != nil {
		log.Printf("Redis Publish failed: %v", err)
	} else {
		log.Println("Redis event published!")
	}
}
