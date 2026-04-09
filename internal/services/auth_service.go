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
	memberRepo ports.TenantMemberRepository
}

func NewAuthService(repo ports.UserRepository, tenantRepo ports.TenantRepository, memberRepo ports.TenantMemberRepository) *AuthService {
	return &AuthService{userRepo: repo, tenantRepo: tenantRepo, memberRepo: memberRepo}
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
	SystemRole   string `json:"system_role"`
	FullName     string `json:"full_name"`
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, *errors.AppError) {
	// Find user uniquely by phone
	user, err := s.userRepo.FindByPhone(req.PhoneNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBadRequest(errors.ErrCodeUserNotFound, "Số điện thoại không tồn tại.", err)
		}
		return nil, errors.NewInternalServer(err)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeWrongPassword, "Mật khẩu không chính xác.", err)
	}

	// Find all tenants the user belongs to
	members, err := s.memberRepo.FindRolesByUser(user.ID.String())
	if (err != nil || len(members) == 0) && (user.SystemRole != domain.SystemRoleAdmin && user.SystemRole != domain.SystemRoleSuperuser) {
		return nil, errors.NewBadRequest(errors.ErrCodeUnauthorized, "Bạn không có quyền truy cập cửa hàng nào.", err)
	}

	var activeRole string
	var activeTenant string

	if user.SystemRole != domain.SystemRoleAdmin && user.SystemRole != domain.SystemRoleSuperuser {
		// Normal User login flow
		var activeMember *domain.TenantMember
		if req.TenantID != nil && *req.TenantID != "" {
			for _, m := range members {
				if m.TenantID.String() == *req.TenantID {
					activeMember = &m
					break
				}
			}
			if activeMember == nil {
				return nil, errors.NewBadRequest(errors.ErrCodeUnauthorized, "Không có quyền vào cửa hàng này.", nil)
			}
		} else {
			activeMember = &members[0]
		}
		activeRole = activeMember.Role
		activeTenant = activeMember.TenantID.String()
	} else {
		// System Role login flow
		if req.TenantID != nil && *req.TenantID != "" {
			activeRole = domain.RoleOwner // Impersonate as Owner
			activeTenant = *req.TenantID
		} else {
			activeRole = "System"
			activeTenant = ""
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID.String(),
		"tenant_id":   activeTenant,
		"role":        activeRole,
		"system_role": string(user.SystemRole),
		"exp":         time.Now().Add(time.Duration(config.AppConfig.JWTExpireMinutes) * time.Minute).Unix(),
	})

	tokenString, err2 := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err2 != nil {
		return nil, errors.NewInternalServer(err2)
	}

	return &AuthResponse{
		Token:      tokenString,
		Role:       activeRole,
		SystemRole: string(user.SystemRole),
		FullName:   user.FullName,
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
		return nil, errors.NewBadRequest(errors.ErrCodeUserNotFound, "Không tìm thấy user", err)
	}

	var results []map[string]interface{}

	if user.SystemRole == domain.SystemRoleSuperuser || user.SystemRole == domain.SystemRoleAdmin {
		// Admin/Superuser sees ALL tenants with role "Owner"
		tenants, err := s.tenantRepo.FindAll()
		if err == nil {
			for _, tenant := range tenants {
				results = append(results, map[string]interface{}{
					"tenant_id":   tenant.ID.String(),
					"tenant_name": tenant.Name,
					"role":        domain.RoleOwner,
				})
			}
		}
	} else {
		// Normal user
		members, err := s.memberRepo.FindRolesByUser(userID)
		if err != nil {
			return nil, errors.NewInternalServer(err)
		}
		for _, acc := range members {
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

	// Verify the user has access to the requested tenant via Pivot table (Skip if admin/superuser)
	var activeRole string
	var activeTenantID string
	if user.SystemRole == domain.SystemRoleSuperuser || user.SystemRole == domain.SystemRoleAdmin {
		activeRole = domain.RoleOwner
		activeTenantID = targetTenantID
	} else {
		member, err := s.memberRepo.FindByUserAndTenant(userID, targetTenantID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.NewBadRequest(errors.ErrCodeUnauthorized, "Bạn không có quyền truy cập cửa hàng này", err)
			}
			return nil, errors.NewInternalServer(err)
		}
		activeRole = member.Role
		activeTenantID = member.TenantID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID.String(),
		"tenant_id":   activeTenantID,
		"role":        activeRole,
		"system_role": string(user.SystemRole),
		"exp":         time.Now().Add(time.Duration(config.AppConfig.JWTExpireMinutes) * time.Minute).Unix(),
	})

	tokenString, err2 := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err2 != nil {
		return nil, errors.NewInternalServer(err2)
	}

	return &AuthResponse{
		Token:        tokenString,
		RefreshToken: "",
		Role:         activeRole,
		SystemRole:   string(user.SystemRole),
		FullName:     user.FullName,
	}, nil
}
