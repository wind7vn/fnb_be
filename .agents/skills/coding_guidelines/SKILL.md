---
name: F&B Backend Golang Conventions
description: Strict coding guidelines, architectural constraints, and conventions for the F&B Multi-Tenant backend.
---

# F&B Backend Golang Conventions

## 1. Directory Structure (Clean Architecture)
- Do not mix business logic with HTTP transport. All controllers go to `internal/handlers`.
- Business logic goes to `internal/services`.
- Database operations go to `internal/repositories`.
- Models and Domain types go to `internal/core/domain`.

## 2. Shared Common logic & Enums
- **No Magic Strings**: All categorical data (e.g., `OrderStatus`) must use explicit Enums stored in `pkg/common/enums`.
- Use Go tags rigorously (`validate:"required,oneof=PENDING READY"`).
- All shared error handling, success wrappers, and utility files go inside `pkg/common/`.

## 3. Database & Models
- Every struct representing a DB table must embed `BaseModel` which contains `created_at`, `updated_at`, `modified_by`, `is_deleted`, `deleted_at`.
- All tenant-owned tables must have `TenantID`.
- You MUST NOT write raw SQL that bypasses tenant scoping. Use the `TenantScope()` or ensure `tenant_id = ?` is present.

## 4. Exception Handling
- **Frontend payloads**: Return standardized JSON: `{"success": false, "error_code": "ERR_CODE", "message": "friendly message", "status_code": 400}`.
- **Log generation**: Use structured logging (Zap logger) to log stack traces and raw DB errors internally. Do not expose SQL errors to the frontend.
