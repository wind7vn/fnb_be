package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   ports.UserRepository
	tenantRepo ports.TenantRepository
}

func NewAuthService(repo ports.UserRepository, tenantRepo ports.TenantRepository) *AuthService {
	return &AuthService{userRepo: repo, tenantRepo: tenantRepo}
}

type LoginRequest struct {
	PhoneNumber string  `json:"phone_number"`
	Password    string  `json:"password"`
	TenantID    *string `json:"tenant_id,omitempty"` // Nullable for Superadmin login
}

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Role         string `json:"role"`
	FullName     string `json:"full_name"`
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, *errors.AppError) {
	// Find user
	user, err := s.userRepo.FindByPhoneAndTenant(req.PhoneNumber, req.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBadRequest(errors.ErrCodeUserNotFound, "Số điện thoại không tồn tại hoặc không thuộc quyền quản lý.", err)
		}
		return nil, errors.NewInternalServer(err)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeWrongPassword, "Mật khẩu không chính xác.", err)
	}

	// Generate JWT
	tenantStr := ""
	if user.TenantID != nil {
		tenantStr = user.TenantID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID.String(),
		"tenant_id": tenantStr,
		"role":      user.Role,
		"exp":       time.Now().Add(time.Duration(config.AppConfig.JWTExpireMinutes) * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}

	// Optional: Handle Refresh Token here later

	return &AuthResponse{
		Token:    tokenString,
		Role:     user.Role,
		FullName: user.FullName,
	}, nil
}

func (s *AuthService) RegisterDevice(userID string, deviceID string, fcmToken string, platform string) *errors.AppError {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.NewBadRequest(errors.ErrCodeValidationFailed, "ID người dùng không hợp lệ", err)
	}

	device := &domain.UserDevice{
		UserID:     uid,
		DeviceID:   deviceID,
		FCMToken:   fcmToken,
		Platform:   platform,
		LastActive: time.Now(),
	}

	if err := s.userRepo.SaveDeviceToken(device); err != nil {
		return errors.NewInternalServer(err)
	}
	return nil
}

func (s *AuthService) GetMe(userID string) (*domain.User, *errors.AppError) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeUserNotFound, "Không tìm thấy người dùng", err)
	}
	return user, nil
}

func (s *AuthService) GenerateGuestToken(tenantID, tableID string) (string, *errors.AppError) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"tenant_id": tenantID,
		"table_id":  tableID,
		"role":      "Guest",
		"exp":       time.Now().Add(12 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", errors.NewInternalServer(err)
	}

	return tokenString, nil
}

func (s *AuthService) GetMyTenants(userID string) ([]map[string]interface{}, *errors.AppError) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeUserNotFound, "Không tìm thấy người dùng", err)
	}

	accounts, err := s.userRepo.FindAllByPhone(user.PhoneNumber)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}

	var results []map[string]interface{}
	for _, acc := range accounts {
		if acc.TenantID != nil {
			tenant, tErr := s.tenantRepo.FindByID(acc.TenantID.String())
			if tErr == nil {
				results = append(results, map[string]interface{}{
					"tenant_id":   tenant.ID.String(),
					"tenant_name": tenant.Name,
					"role":        acc.Role,
				})
			}
		}
	}
	return results, nil
}

func (s *AuthService) SwitchTenant(userID string, targetTenantID string) (*AuthResponse, *errors.AppError) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeUserNotFound, "Không tìm thấy người dùng", err)
	}

	targetUser, err := s.userRepo.FindByPhoneAndTenant(user.PhoneNumber, &targetTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBadRequest(errors.ErrCodeUnauthorized, "Bạn không có quyền truy cập cửa hàng này", err)
		}
		return nil, errors.NewInternalServer(err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   targetUser.ID.String(),
		"tenant_id": targetUser.TenantID.String(),
		"role":      targetUser.Role,
		"exp":       time.Now().Add(time.Duration(config.AppConfig.JWTExpireMinutes) * time.Minute).Unix(),
	})

	tokenString, err2 := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err2 != nil {
		return nil, errors.NewInternalServer(err2)
	}

	return &AuthResponse{
		Token:        tokenString,
		RefreshToken: "", // Can implement refresh token logic later
		Role:         targetUser.Role,
		FullName:     targetUser.FullName,
	}, nil
}
