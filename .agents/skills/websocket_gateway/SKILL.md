---
name: F&B Backend WebSocket & Real-Time Gateway
description: Guidelines and architecture rules for managing WebSocket connections, event payloads, and Redis Pub/Sub sync.
---

# Skill: WebSocket Gateway & Real-Time Sync

The KDS (Kitchen Display System) and table occupancy dashboard require real-time synchronization across devices (POS, Waiter App, Chef Display).

## 1. Gateway Connection Lifecycle
- Clients establish connection via `wss://<api_domain>/ws?token=<jwt_token>`
- The gateway extracts `tenant_id` from the JWT claims.
- The connection is registered with the localized in-memory `Hub` and subscribed to the tenant's Redis Pub/Sub channel.

## 2. Pub/Sub Channel Namespace
- Redis channel names must follow the format: `tenant:<tenant_id>:events`
- All nodes in the cluster listen to these channels to broadcast events to locally connected TCP sockets.

## 3. Standard WebSocket Event Payload
Every event must conform to the standard JSON envelope structure:
```json
{
  "type": "EVENT_TYPE_STRING",
  "payload": {
    "key": "value"
  }
}
```

Predefined standard event types (`internal/core/domain/constants.go`):
- `TABLE_STATUS_CHANGED`: Live updates for occupancy grid.
- `ORDER_CREATED` / `ORDER_UPDATED`: Trigger ticket reload.
- `KDS_NEW_TICKET` / `KDS_ITEM_READY`: Sync order items between POS, Waiter, and Kitchen.
- `PAYMENT_COMPLETED`: Close table session.
