---
name: F&B Backend Database Utilities & Migrations
description: Core conventions and instructions for database migrations, model schema mapping, and GORM database cleanup tasks.
---

# Skill: GORM Database Migrations & Cleanups

When designing, updating, or auditing database tables in the F&B backend, you must follow these standards.

## 1. GORM BaseModel & Audit Fields
Every GORM database model must embed `BaseModel` from `internal/core/domain/base_model.go`:
```go
type BaseModel struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	ModifiedBy *uuid.UUID     `gorm:"type:uuid" json:"modified_by,omitempty"`
	IsDeleted  bool           `gorm:"not null;default:false" json:"is_deleted"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## 2. GORM Migrations (`cmd/migrator`)
- Run migrations: `go run cmd/migrator/main.go`
- This script adds missing tables, columns, indices, and constraints natively via GORM's `AutoMigrate()`.
- **Warning**: It never automatically deletes or drops columns to avoid data loss.

## 3. Structural Pruning (`cmd/cleanup`)
- Run structural cleanup: `go run cmd/cleanup/main.go`
- This script leverages Go reflection to compare Golang domain struct definitions against database column lists.
- It safely drops columns that have been removed or renamed in the Go models to prune DB bloat in staging and dev.
- **Caution**: Run with care on staging/production to ensure no critical data is pruned.
