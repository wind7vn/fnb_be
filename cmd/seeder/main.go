package main

import (
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
	"github.com/wind7vn/fnb_be/pkg/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
)

func main() {
	config.LoadConfig()
	logger.InitLogger(config.AppConfig.Env)
	db.ConnectDB()

	logger.Log.Info("Starting Seeder...")

	hashCost := bcrypt.DefaultCost
	hashPW, _ := bcrypt.GenerateFromPassword([]byte("123456"), hashCost)

	// 1. Create Superadmin
	superadmin := domain.User{
		Role:         domain.RoleSuperadmin,
		PhoneNumber:  "0999999999",
		FullName:     "System Admin",
		PasswordHash: string(hashPW),
		Metadata:     datatypes.JSON("{}"),
	}
	if err := db.DB.FirstOrCreate(&superadmin, domain.User{PhoneNumber: "0999999999"}).Error; err != nil {
		logger.Log.Error("Superadmin creation failed: " + err.Error())
	}

	// 2. Create Tenant (Restaurant A)
	tenantID, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")
	tenant := domain.Tenant{
		BaseModel: domain.BaseModel{ID: tenantID}, // Force fixed UUID for easy testing
		Name:      "Demo Restaurant A",
		Timezone:  "Asia/Ho_Chi_Minh",
		Metadata:  `{"printer_ips": ["192.168.1.50"]}`,
	}
	db.DB.FirstOrCreate(&tenant, domain.Tenant{Name: "Demo Restaurant A"})

	// 3. Create Owner & Staff
	owner := domain.User{
		TenantID:     &tenant.ID,
		Role:         domain.RoleOwner,
		PhoneNumber:  "0888888888",
		FullName:     "John Owner",
		PasswordHash: string(hashPW),
		Metadata:     datatypes.JSON("{}"),
	}
	db.DB.FirstOrCreate(&owner, domain.User{PhoneNumber: "0888888888", TenantID: &tenant.ID})

	staff1 := domain.User{
		TenantID:     &tenant.ID,
		Role:         domain.RoleStaff,
		PhoneNumber:  "0777777777",
		FullName:     "Waiter Alice",
		PasswordHash: string(hashPW),
		Metadata:     datatypes.JSON("{}"),
	}
	db.DB.FirstOrCreate(&staff1, domain.User{PhoneNumber: "0777777777", TenantID: &tenant.ID})

	// 4. Create Tables
	tables := []domain.Table{
		{TenantID: tenant.ID, Name: "T01", Status: domain.TableStatusAvailable},
		{TenantID: tenant.ID, Name: "T02", Status: domain.TableStatusAvailable},
		{TenantID: tenant.ID, Name: "VIP1", Status: domain.TableStatusOccupied},
	}
	for _, t := range tables {
		db.DB.FirstOrCreate(&t, domain.Table{TenantID: tenant.ID, Name: t.Name})
	}

	// 5. Create Product
	product := domain.Product{
		TenantID:    tenant.ID,
		Name:        "Phở Bò Kobe",
		Description: "Nước dùng ngọt thanh, bò Kobe thượng hạng",
		Price:       150000,
		IsAvailable: true,
	}
	db.DB.FirstOrCreate(&product, domain.Product{TenantID: tenant.ID, Name: "Phở Bò Kobe"})

	logger.Log.Info("Seeding Complete. Try login with Superadmin 0999999999 or Owner 0888888888 (pw: 123456)")
}
