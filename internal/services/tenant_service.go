package services

import (
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"golang.org/x/crypto/bcrypt"
)

type TenantService struct {
	tenantRepo ports.TenantRepository
	userRepo   ports.UserRepository
	memberRepo ports.TenantMemberRepository
}

func NewTenantService(tenantRepo ports.TenantRepository, userRepo ports.UserRepository, memberRepo ports.TenantMemberRepository) *TenantService {
	return &TenantService{tenantRepo: tenantRepo, userRepo: userRepo, memberRepo: memberRepo}
}

type CreateTenantRequest struct {
	Name            string `json:"name"`
	Timezone        string `json:"timezone"`
	OwnerPhone      string `json:"owner_phone"`
	OwnerFullName   string `json:"owner_full_name"`
	OwnerPassword   string `json:"owner_password"` // Initial password
}

func (s *TenantService) CreateTenant(req CreateTenantRequest) (*domain.Tenant, *errors.AppError) {
	// Create Tenant
	tenant := &domain.Tenant{
		Name:     req.Name,
		Timezone: req.Timezone,
		IsActive: true,
	}

	if err := s.tenantRepo.Create(tenant); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	// Check if user already exists
	owner, errUser := s.userRepo.FindByPhone(req.OwnerPhone)
	if errUser != nil || owner == nil {
		hashPW, _ := bcrypt.GenerateFromPassword([]byte(req.OwnerPassword), bcrypt.DefaultCost)
		owner = &domain.User{
			PhoneNumber:  req.OwnerPhone,
			FullName:     req.OwnerFullName,
			PasswordHash: string(hashPW),
		}

		if err := s.userRepo.Create(owner); err != nil {
			return nil, errors.NewInternalServer(err)
		}
	}

	// Create map between Owner and Tenant
	member := &domain.TenantMember{
		UserID:   owner.ID, // User returned from create/find
		TenantID: tenant.ID,
		Role:     domain.RoleOwner,
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	return tenant, nil
}

type CreateAdminRequest struct {
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	Password    string `json:"password"`
}

func (s *TenantService) CreateSystemAdmin(req CreateAdminRequest) (*domain.User, *errors.AppError) {
	existing, errUser := s.userRepo.FindByPhone(req.PhoneNumber)
	if errUser == nil && existing != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Số điện thoại đã tồn tại", nil)
	}

	hashPW, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	adminUser := &domain.User{
		PhoneNumber:  req.PhoneNumber,
		FullName:     req.FullName,
		PasswordHash: string(hashPW),
		SystemRole:   domain.SystemRoleAdmin,
	}

	if err := s.userRepo.Create(adminUser); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	return adminUser, nil
}

func (s *TenantService) GetAllTenants() ([]domain.Tenant, *errors.AppError) {
	tenants, err := s.tenantRepo.FindAll()
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return tenants, nil
}

type CreateStaffRequest struct {
	Role        string `json:"role"`
	PhoneNumber string `json:"phone_number"`
	FullName    string `json:"full_name"`
	Password    string `json:"password"`
}

func (s *TenantService) CreateStaff(tenantID string, req CreateStaffRequest) (*domain.User, *errors.AppError) {
	// Must be Owner or Manager doing this. Role checking is in handler.
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Mã cửa hàng không hợp lệ", err)
	}

	// Check if user already exists
	staff, errUser := s.userRepo.FindByPhone(req.PhoneNumber)
	if errUser != nil || staff == nil {
		hashPW, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		staff = &domain.User{
			PhoneNumber:  req.PhoneNumber,
			FullName:     req.FullName,
			PasswordHash: string(hashPW),
		}

		if err := s.userRepo.Create(staff); err != nil {
			return nil, errors.NewInternalServer(err)
		}
	}

	// Map to Tenant
	member := &domain.TenantMember{
		UserID:   staff.ID,
		TenantID: tid,
		Role:     req.Role, // Should validate if it's "Staff" or "Manager"
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	return staff, nil
}

func (s *TenantService) GetStaff(tenantID string) ([]map[string]interface{}, *errors.AppError) {
	members, err := s.memberRepo.FindStaffByTenant(tenantID)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}

	var results []map[string]interface{}
	for _, m := range members {
		user, _ := s.userRepo.FindByID(m.UserID.String())
		if user != nil {
			results = append(results, map[string]interface{}{
				"id":           user.ID,
				"phone_number": user.PhoneNumber,
				"full_name":    user.FullName,
				"role":         m.Role,
			})
		}
	}

	return results, nil
}

type UpdateSettingsRequest struct {
	Metadata string `json:"metadata"` // JSON string payload
}

func (s *TenantService) UpdateSettings(tenantID string, req UpdateSettingsRequest) *errors.AppError {
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return errors.NewBadRequest(errors.ErrCodeValidationFailed, "Không tìm thấy dữ liệu cửa hàng", err)
	}

	tenant.Metadata = req.Metadata
	if err := s.tenantRepo.Update(tenant); err != nil {
		return errors.NewInternalServer(err)
	}
	return nil
}

func (s *TenantService) GetSettings(tenantID string) (string, *errors.AppError) {
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return "", errors.NewBadRequest(errors.ErrCodeValidationFailed, "Không tìm thấy dữ liệu cửa hàng", err)
	}
	return tenant.Metadata, nil
}
