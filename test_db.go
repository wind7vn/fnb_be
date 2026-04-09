package main
import (
	"fmt"
	"github.com/wind7vn/fnb_be/pkg/config"
	"github.com/wind7vn/fnb_be/pkg/db"
	"github.com/wind7vn/fnb_be/internal/core/domain"
)
func main() {
	config.LoadConfig()
	db.ConnectDB()
	var user domain.User
	db.DB.Where("phone_number = ?", "0999999999").First(&user)
	fmt.Printf("Phone: %s, SystemRole: '%s'\n", user.PhoneNumber, user.SystemRole)
}
