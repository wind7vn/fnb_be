package services

import (
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"gorm.io/gorm"
)

type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(r ports.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

type ProductRequest struct {
	Category    string  `json:"category"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	IsAvailable bool    `json:"is_available"`
}

func (s *ProductService) Create(tenantID string, req ProductRequest) (*domain.Product, *errors.AppError) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Tenant ID không hợp lệ", err)
	}

	product := &domain.Product{
		TenantID:    tid,
		Category:    req.Category,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		IsAvailable: req.IsAvailable,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return product, nil
}

func (s *ProductService) GetAll(tenantID string, category *string) ([]domain.Product, *errors.AppError) {
	products, err := s.repo.FindAllByTenant(tenantID, category)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return products, nil
}

func (s *ProductService) Update(tenantID string, productID string, req ProductRequest) (*domain.Product, *errors.AppError) {
	product, err := s.repo.FindByID(productID, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBadRequest(errors.ErrCodeItemOutOfStock, "Sản phẩm không tồn tại", err)
		}
		return nil, errors.NewInternalServer(err)
	}

	product.Category = req.Category
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.ImageURL = req.ImageURL
	product.IsAvailable = req.IsAvailable

	if err := s.repo.Update(product); err != nil {
		return nil, errors.NewInternalServer(err)
	}
	return product, nil
}

func (s *ProductService) Delete(tenantID string, productID string) *errors.AppError {
	if err := s.repo.Delete(productID, tenantID); err != nil {
		return errors.NewInternalServer(err)
	}
	return nil
}
