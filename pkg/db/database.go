package db

import (
	"fmt"
	"time"

	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC",
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBName,
		config.AppConfig.DBSSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // Output plural table names like `users`
		},
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		logger.Log.Fatal("Failed to connect to database: " + err.Error())
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.Fatal("Failed to get DB instance: " + err.Error())
	}

	// Optimize Connection Pool
	sqlDB.SetMaxIdleConns(config.AppConfig.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.AppConfig.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Log.Info("Database connection established & pool configured.")
}

// TenantScope is a critical GORM scope applied to EVERY repository operation
// to dynamically inject `WHERE tenant_id = ?` into queries to isolate tenant data.
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID == "" {
			// Protect against Postgres invalid input syntax for type uuid: "" error.
			// Currently if tenantID is empty, it means System Auth Context. 
			// We fallback to a dummy fake UUID so it correctly returns 0 records
			// instead of crashing Postgres. (Or you could return db to fetch ALL, 
			// but we want safety first). We'll use Nil UUID.
			return db.Where("tenant_id = ?", "00000000-0000-0000-0000-000000000000")
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}
