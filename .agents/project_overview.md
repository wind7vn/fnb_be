# F&B Multi-Tenant Architecture - Brain Context

## 1. System Objective
A robust, Multi-Tenant Food & Beverage Management System (POS + Kitchen Display System). Built with Golang, PostgreSQL, and Redis. SaaS model where multiple independent restaurants (Tenants) share the backend with **strictly isolated data**.

## 2. Core Architecture & Standards
- **Language**: Golang 1.26+ following Hexagonal/Clean architecture logic.
- **Database**: PostgreSQL with universally applied `BaseModel` (`created_at`, `modified_by`, `is_deleted`).
- **Data Isolation**: Every tenant-owned query expects `tenant_id`. Middleware handles this via `c.Locals("tenant_id")`.
- **RBAC Hierarchy**: `Superadmin` (System) -> `Owner` (Tenant) -> (`Manager`, `Staff`).
- **Dual-Layer Error Handling**: Raw Zap/Logrus logs for trace -> Sanitized JSON (`error_code`, `message`) for UI.
- **Enums Strategy**: Prevent magic strings. Declare core types in `pkg/common/enums`.

## 3. Real-Time & Event Syncs
- **Push Notification System**: Driven by `USER_DEVICE` (`device_id`, `fcm_token`) and internal `NOTIFICATION` tables.
- **WebSockets Gateway**: Powered by Redis PubSub using namespaced channels `tenant:<id>:events`.
- **Standardized WS Events**: `TABLE_STATUS_CHANGED`, `ORDER_CREATED`, `KDS_NEW_TICKET`, `KDS_ITEM_READY`, `PAYMENT_COMPLETED`.

## 4. Nuanced API Endpoints Context
- `POST /system/users`: Superadmin initializes an Owner.
- `POST /tenants/users`: Owners spin up local `Manager` or `Staff` accounts.
- `POST /orders/guest`: Tokenless customer QR orders leveraging generated `Guest Token`.
- `PUT /auth/me`: Independent user profile management.
- `POST /auth/devices`: Devices register their Push Token / FCM ID here.
