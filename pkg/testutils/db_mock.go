package testutils

import (
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupMockDB creates an in-memory SQLite database specifically tuned for Unit Testing
// It automatically migrates the models so repositories can test standard CRUD cleanly.
func SetupMockDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to in-memory mocked DB: " + err.Error())
	}

	// Auto-migrate essential tables for isolated execution
	_ = db.AutoMigrate(
		&domain.Tenant{},
		&domain.User{},
		&domain.UserDevice{},
		&domain.Table{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
	)

	return db
}
