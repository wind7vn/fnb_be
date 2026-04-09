package main

import (
	"log"

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
	db.DB.AutoMigrate(&domain.User{}, &domain.Tenant{}, &domain.TenantMember{})

	logger.Log.Info("Starting Seeder for User 0982651922 with Centralized Model...")

	hashCost := bcrypt.DefaultCost
	hashPW, _ := bcrypt.GenerateFromPassword([]byte("123"), hashCost)

	// 1. Tạo User (Sếp Tùng)
	userPhone := "0982651922"
	var user domain.User
	if err := db.DB.Where("phone_number = ?", userPhone).First(&user).Error; err != nil {
		user = domain.User{
			PhoneNumber:  userPhone,
			FullName:     "Sếp Tùng",
			PasswordHash: string(hashPW),
			Metadata:     datatypes.JSON("{}"),
		}
		if dbErr := db.DB.Create(&user).Error; dbErr != nil {
			log.Fatal("Failed to create User:", dbErr)
		}
		log.Println("Created User:", userPhone)
	} else {
		log.Println("User already exists:", userPhone)
	}

	// 1.5. Tạo Superuser (Trùm Cuối)
	adminPhone := "0999999999"
	var admin domain.User
	if err := db.DB.Where("phone_number = ?", adminPhone).First(&admin).Error; err != nil {
		admin = domain.User{
			PhoneNumber:  adminPhone,
			FullName:     "Trùm Cuối Hệ Thống",
			PasswordHash: string(hashPW),
			SystemRole:   domain.SystemRoleSuperuser,
			Metadata:     datatypes.JSON("{}"),
		}
		if dbErr := db.DB.Create(&admin).Error; dbErr != nil {
			log.Fatal("Failed to create Superuser:", dbErr)
		}
		log.Println("Created Superuser:", adminPhone)
	} else {
		admin.SystemRole = domain.SystemRoleSuperuser
		db.DB.Save(&admin)
		log.Println("Superuser already exists:", adminPhone)
	}

	// 2. Định nghĩa các chi nhánh
	tenantsData := []string{
		"Bún Đậu Làng Mơ - Chi Nhánh Cầu Giấy (Quyền Owner)",
		"Bún Đậu Làng Mơ - Chi Nhánh Hai Bà Trưng (Quyền Manager)",
		"Bún Đậu Làng Mơ - Chi Nhánh Hà Đông (Quyền Staff)",
	}
	
	roles := []string{
		domain.RoleOwner,
		domain.RoleManager,
		domain.RoleStaff,
	}

	for i, tName := range tenantsData {
		var tenant domain.Tenant
		if err := db.DB.Where("name = ?", tName).First(&tenant).Error; err != nil {
			tenant = domain.Tenant{
				Name:     tName,
				Timezone: "Asia/Ho_Chi_Minh",
				IsActive: true,
				Metadata: "{}",
			}
			if dbErr := db.DB.Create(&tenant).Error; dbErr != nil {
				log.Fatal("Failed to create Tenant:", dbErr)
			}
			log.Println("Created Tenant:", tName)
		}

		// 3. Liên kết User vào Tenant qua bảng TenantMember
		var member domain.TenantMember
		err := db.DB.Where("user_id = ? AND tenant_id = ?", user.ID, tenant.ID).First(&member).Error
		if err != nil {
			member = domain.TenantMember{
				UserID:   user.ID,
				TenantID: tenant.ID,
				Role:     roles[i],
			}
			if dbErr := db.DB.Create(&member).Error; dbErr != nil {
				log.Fatal("Failed to create TenantMember:", dbErr)
			}
			log.Printf("Mapped User %s as %s in %s\n", userPhone, roles[i], tName)
		} else {
			log.Printf("User %s already mapped as %s in %s\n", userPhone, member.Role, tName)
		}

		// 4. Tạo Bàn (Tables)
		tables := []string{"Bàn 1", "Bàn 2", "Bàn 3", "Bàn 4", "Bàn 5", "Bàn 6", "VIP 1", "VIP 2", "VIP 3"}
		for _, tbName := range tables {
			var table domain.Table
			if db.DB.Where("tenant_id = ? AND name = ?", tenant.ID, tbName).First(&table).Error != nil {
				table = domain.Table{
					TenantID: tenant.ID,
					Name:     tbName,
					Status:   "Available",
				}
				db.DB.Create(&table)
			}
		}

		// 5. Tạo Món ăn (Products)
		products := []struct{
			Name  string
			Price float64
		}{
			{"Bún Đậu Thập Cẩm", 55000},
			{"Bún Đậu Trứng Chiên", 35000},
			{"Bún Đậu Nhồi Thịt", 45000},
			{"Nem Chua Rán", 40000},
			{"Dồi Sụn Nướng", 50000},
			{"Trà Tắc khổng lồ", 15000},
			{"Nước Mơ", 12000},
			{"Bia Sài Gòn", 20000},
			{"Cháo Quẩy", 30000},
			{"Mắm Tôm Chén Thêm", 5000},
		}

		for _, p := range products {
			var prod domain.Product
			if db.DB.Where("tenant_id = ? AND name = ?", tenant.ID, p.Name).First(&prod).Error != nil {
				prod = domain.Product{
					TenantID:    tenant.ID,
					Name:        p.Name,
					Price:       p.Price,
					IsAvailable: true,
					Description: "Đặc sản chuẩn vị Làng Mơ",
				}
				db.DB.Create(&prod)
			}
		}
		log.Printf("Created 9 Tables and 10 Products for %s\n", tName)
	}

	logger.Log.Info("Seeding Complete. Vui lòng đăng nhập với SĐT: 0982651922 và Password: 123")
}
