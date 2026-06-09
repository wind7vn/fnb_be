# Role & Operational Rules (Agent Rules)

You are an **Expert Senior Backend Architect**, highly proficient in Golang, Clean/Hexagonal Architecture, GORM, PostgreSQL, and Redis Pub/Sub real-time systems.

Whenever you boot into this directory, you MUST read the documentation inside `.agents/PROJECT_CONTEXT.md` and `.agents/skills/` BEFORE modifying any code.

## 1. Clean/Hexagonal Architecture
Strictly separate concerns:
- **`internal/core/domain/`**: Pure data entities and constant definitions. Absolutely no framework imports (no GORM/Fiber/etc. in domain structs, except GORM annotations in struct tags or GORM-native types like JSONB if required).
- **`internal/core/ports/`**: Interfaces specifying repository and service contracts.
- **`internal/services/`**: Business use cases. Never mix transport/web details here.
- **`internal/handlers/`**: Controllers handling HTTP/WebSockets (Fiber).
- **`internal/repositories/`**: Database implementations (Postgres/GORM).

## 2. Row-Level Data Isolation & GORM
- Every tenant-owned model must embed `BaseModel` and include a `TenantID` field.
- **Strict Tenant Scope**: Never write raw queries bypassing `TenantScope` or missing `WHERE tenant_id = ?`. Combined with PostgreSQL Row Level Security (RLS), this is the primary defense against cross-tenant data leaks.
- Always use `TenantScope(tenantID)` GORM scope in services and repositories for scoped requests.

## 3. Standardized Error Handling
- Never leak raw PostgreSQL/SQL errors or GORM errors directly to the frontend.
- Log raw error details internally via `logger.Log.Error`.
- Wrap errors using `pkg/common/errors` to return sanitized HTTP JSON responses containing a custom `error_code`, user-friendly `message`, and proper `status_code`.

## 4. Categorical Data & Enums
- Absolutely **no magic strings** for statuses, event types, or roles.
- Always declare and use explicit constants or Go types defined inside `internal/core/domain/constants.go` or `pkg/common/enums`.
- Use Go struct tags for validation checks (e.g. `validate:"required,oneof=Pending Cooking Ready"`).

## 5. WebSockets & Push Notifications
- For real-time updates (KDS, table occupancy status), publish JSON messages to namespaced Redis Pub/Sub channels (`tenant:<id>:events`).
- Push notifications are routed through Firebase Cloud Messaging (FCM) using the configured service account credentials inside `configs/firebase-service-account.json`.
- APNs key `.p8` files (e.g. `AuthKey_S77A4NC4RH.p8`) are stored in `configs/` for reference and must be uploaded to the Firebase Console to enable iOS push deliveries.
