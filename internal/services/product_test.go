package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/repositories"
	"github.com/wind7vn/fnb_be/internal/services"
	"github.com/wind7vn/fnb_be/pkg/testutils"
)

func TestProductTenantIsolation(t *testing.T) {
	db := testutils.SetupMockDB()
	tenantARepo := repositories.NewProductRepository(db)
	svc := services.NewProductService(tenantARepo)

	// Create Tenant A & B IDs
	tenantAID, _ := uuid.NewV7()
	tenantBID, _ := uuid.NewV7()

	// Direct DB injection of product belonging to Tenant B
	productID, _ := uuid.NewV7()
	db.Create(&domain.Product{
		BaseModel:   domain.BaseModel{ID: productID},
		TenantID:    tenantBID,
		Name:        "Tra Sua B",
		Description: "Belongs to Tenant B",
		Price:       35000,
	})

	// Tenant A tries to Get, Update, or Delete Tenant B's product
	t.Run("Get All should filter", func(t *testing.T) {
		products, err := svc.GetAll(tenantAID.String(), nil)
		assert.Nil(t, err)
		assert.Len(t, products, 0) // Shielded!
	})

	t.Run("Update should fail finding record", func(t *testing.T) {
		_, err := svc.Update(tenantAID.String(), productID.String(), services.ProductRequest{
			Name: "Hacked",
		})
		assert.NotNil(t, err)
		assert.Equal(t, 400, err.StatusCode) // Not Found -> BadRequest translation
	})

	t.Run("Filtering by category", func(t *testing.T) {
		catA := "Drink"
		catB := "Food"
		db.Create(&domain.Product{
			BaseModel:   domain.BaseModel{ID: uuid.New()},
			TenantID:    tenantAID,
			Category:    catA,
			Name:        "Tra Sua A",
			Price:       35000,
		})
		db.Create(&domain.Product{
			BaseModel:   domain.BaseModel{ID: uuid.New()},
			TenantID:    tenantAID,
			Category:    catB,
			Name:        "Com Chien A",
			Price:       55000,
		})

		// Get only Drink
		products, err := svc.GetAll(tenantAID.String(), &catA)
		assert.Nil(t, err)
		assert.Len(t, products, 1)
		assert.Equal(t, "Tra Sua A", products[0].Name)
	})
}
