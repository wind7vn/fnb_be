package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/wind7vn/fnb_be/internal/core/domain"
)

func main() {
	dsn := "host=127.0.0.1 user=wind password=WsA8J6PPuf8N6ysVlmNB dbname=fnb port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 1. Create a Tenant
	tenant := domain.Tenant{
		Name:     "Test Restaurant",
		Timezone: "UTC",
		IsActive: true,
		Metadata: `{"tax_percent": 10}`,
	}
	if err := db.Create(&tenant).Error; err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}

	// 2. Create the Owner User
	hashPW, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	owner := domain.User{
		TenantID:     &tenant.ID,
		PhoneNumber:  "0982651922",
		PasswordHash: string(hashPW),
		Role:         domain.RoleOwner,
		FullName:     "Phong Nguyen",
	}
	if err := db.Create(&owner).Error; err != nil {
		log.Fatalf("Failed to create owner user: %v", err)
	}

	// 3. Create some Tables
	tables := []domain.Table{
		{TenantID: tenant.ID, Name: "Bàn 1", Status: domain.TableStatusAvailable},
		{TenantID: tenant.ID, Name: "Bàn 2", Status: domain.TableStatusAvailable},
		{TenantID: tenant.ID, Name: "Bàn 3", Status: domain.TableStatusOccupied},
	}
	for _, t := range tables {
		db.Create(&t)
	}

	// 4. Create some Products
	products := []domain.Product{
		{TenantID: tenant.ID, Name: "Cơm Rang Dưa Bò", Price: 55000, IsAvailable: true},
		{TenantID: tenant.ID, Name: "Phở Bò Kobe", Price: 350000, IsAvailable: true},
		{TenantID: tenant.ID, Name: "Cà phê Sữa Đá", Price: 25000, IsAvailable: true},
		{TenantID: tenant.ID, Name: "Sinh Tố Xoài", Price: 40000, IsAvailable: true},
	}
	for _, p := range products {
		db.Create(&p)
	}

	fmt.Println("Seeding complete! You can now login with 0982651922 / 123456")
}
