package main

import (
	"log"

	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
	"github.com/wind7vn/fnb_be/pkg/db"
)

func main() {
	// 1. Initial Setup
	config.LoadConfig()
	logger.InitLogger(config.AppConfig.Env)
	defer logger.Sync()

	db.ConnectDB()

	logger.Log.Info("Starting Database Migrations...")

	// 2. Pre-Migration: Raw SQL for UUIDs and Extensions
	execRawSQL(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)

	// 3. AutoMigrate Entities (GORM)
	err := db.DB.AutoMigrate(
		&domain.Tenant{},
		&domain.User{},
		&domain.UserDevice{},
		&domain.Table{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.ActionLog{},
		&domain.Notification{},
	)

	if err != nil {
		logger.Log.Fatal("AutoMigrate failed: " + err.Error())
	}

	logger.Log.Info("Core Entity Schema AutoMigrate successful.")

	// 4. Post-Migration: Advanced DB Constraints / Roles / RLS via Raw SQL
	// e.g. Unique composite index: A phonenumber is unique per Tenant
	execRawSQL(`CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_phone ON "user" (tenant_id, phone_number) WHERE tenant_id IS NOT NULL;`)

	logger.Log.Info("Migrations Applied Successfully!")
}

func execRawSQL(query string) {
	if err := db.DB.Exec(query).Error; err != nil {
		log.Printf("Warning during RawSQL execution: %v (Query: %s)\n", err, query)
	}
}
