package handlers

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/wind7vn/fnb_be/internal/middlewares"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
)

type WSHandler struct {
	pubSub *services.PubSubService
	// TenantID -> Array of WS connections
	clients    map[string][]*websocket.Conn
	clientsMtx sync.RWMutex
}

func NewWSHandler(ps *services.PubSubService) *WSHandler {
	h := &WSHandler{
		pubSub:  ps,
		clients: make(map[string][]*websocket.Conn),
	}
	
	// Start consuming from Redis
	go h.listenRedisEvents()
	
	return h
}

func (h *WSHandler) SetupRoutes(router fiber.Router) {
	// Upgrade check middleware
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wsGroup := router.Group("/ws")
	// Use JWT middleware to inject tenant_id and role into ctx
	// But Fiber websockets can't access c.Locals *inside* the websocket handler directly as easily, 
	// actually it CAN, we extract it before creating the connection handler.
	wsGroup.Use(middlewares.JWTMiddleware())
	
	wsGroup.Get("/kds", websocket.New(h.HandleKDSConnection))
}

func (h *WSHandler) HandleKDSConnection(c *websocket.Conn) {
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok {
		c.Close()
		return
	}

	h.addClient(tenantID, c)
	defer h.removeClient(tenantID, c)

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break // Connection closed or errored
		}
	}
}

func (h *WSHandler) addClient(tenantID string, c *websocket.Conn) {
	h.clientsMtx.Lock()
	defer h.clientsMtx.Unlock()
	h.clients[tenantID] = append(h.clients[tenantID], c)
}

func (h *WSHandler) removeClient(tenantID string, c *websocket.Conn) {
	h.clientsMtx.Lock()
	defer h.clientsMtx.Unlock()
	conns := h.clients[tenantID]
	for i, conn := range conns {
		if conn == c {
			h.clients[tenantID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *WSHandler) listenRedisEvents() {
	if h.pubSub == nil {
		return
	}

	pubsub := h.pubSub.Subscribe("KDS_EVENTS")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var payload services.EventPayload
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			logger.Log.Sugar().Errorf("WS Redis Unmarshal Error: %v", err)
			continue
		}

		// Broadcast to specific tenant's clients
		h.clientsMtx.RLock()
		tenantClients := h.clients[payload.TenantID]
		h.clientsMtx.RUnlock()

		for _, client := range tenantClients {
			if err := client.WriteJSON(payload); err != nil {
				client.Close() // Will be removed in next lifecycle
			}
		}
	}
}
