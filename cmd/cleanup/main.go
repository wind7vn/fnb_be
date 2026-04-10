package main

import (
	"fmt"
	"log"

	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
	"github.com/wind7vn/fnb_be/pkg/db"
	"gorm.io/gorm"
)

func main() {
	config.LoadConfig()
	logger.InitLogger(config.AppConfig.Env)
	defer logger.Sync()
	
	db.ConnectDB()

	// Danh sách tất cả các Models tham gia hệ thống
	models := []interface{}{
		&domain.Tenant{},
		&domain.User{},
		&domain.UserDevice{},
		&domain.TenantMember{},
		&domain.Table{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.ActionLog{},
		&domain.Notification{},
	}

	fmt.Println("Bắt đầu truy quét các cột bị dư thừa trên Database...")
	
	for _, model := range models {
		cleanupTable(db.DB, model)
	}

	fmt.Println("==========================================")
	fmt.Println("🎉 Hoàn tất dọn dẹp các thông tin dư thừa!")
	fmt.Println("==========================================")
}

func cleanupTable(database *gorm.DB, model interface{}) {
	stmt := &gorm.Statement{DB: database}
	err := stmt.Parse(model)
	if err != nil {
		log.Printf("Lỗi parse model: %v", err)
		return
	}

	tableName := stmt.Schema.Table
	
	// Bỏ qua nếu bảng chưa tồn tại
	if !database.Migrator().HasTable(model) {
		return
	}

	dbColumns, err := database.Migrator().ColumnTypes(model)
	if err != nil {
		log.Printf("Lỗi lấy thông tin cột của bảng %s: %v", tableName, err)
		return
	}

	// Đưa tất cả các Field Name của Struct (được khai báo) vào một Map để so sánh
	structCols := make(map[string]bool)
	for _, field := range stmt.Schema.Fields {
		structCols[field.DBName] = true
	}

	// Duyệt qua từng cột đang có thật dưới database
	// Nếu không có trong Struct -> Tức là bạn đã Xoá/Comment -> Quét & Drop nó.
	for _, dbCol := range dbColumns {
		colName := dbCol.Name()
		if !structCols[colName] {
			fmt.Printf("⚠️ Phát hiện cột thừa: Bảng [%s] -> Cột [%s]. Đang tiến hành Drop...\n", tableName, colName)
			errDrop := database.Migrator().DropColumn(model, colName)
			if errDrop != nil {
				fmt.Printf("❌ Lỗi khi tải Drop %s.%s: %v\n", tableName, colName, errDrop)
			} else {
				fmt.Printf("✅ Đã xoá thành công %s.%s\n", tableName, colName)
			}
		}
	}
}
