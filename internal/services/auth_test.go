package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/repositories"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/config"
	"github.com/wind7vn/fnb_be/pkg/testutils"
	"golang.org/x/crypto/bcrypt"
)

func setupTest() *services.AuthService {
	config.AppConfig.JWTSecret = "test_secret"
	config.AppConfig.JWTExpireMinutes = 60

	db := testutils.SetupMockDB()

	// Create mock user
	hashCost := bcrypt.DefaultCost
	hashPW, _ := bcrypt.GenerateFromPassword([]byte("password123"), hashCost)

	uid, _ := uuid.NewV7()
	user := domain.User{
		BaseModel:    domain.BaseModel{ID: uid},
		Role:         "Staff",
		PhoneNumber:  "0123456789",
		FullName:     "Test User",
		PasswordHash: string(hashPW),
	}
	db.Create(&user)

	repo := repositories.NewUserRepository(db)
	tenantRepo := repositories.NewTenantRepository(db)
	return services.NewAuthService(repo, tenantRepo)
}

func TestLogin_Success(t *testing.T) {
	svc := setupTest()

	resp, err := svc.Login(services.LoginRequest{
		PhoneNumber: "0123456789",
		Password:    "password123",
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "Staff", resp.Role)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := setupTest()

	resp, err := svc.Login(services.LoginRequest{
		PhoneNumber: "0123456789",
		Password:    "wrongpass",
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, errors.ErrCodeWrongPassword, err.ErrorCode)
}
