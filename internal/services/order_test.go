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

func TestOrderCartCalculations(t *testing.T) {
	db := testutils.SetupMockDB()
	
	orderRepo := repositories.NewOrderRepository(db)
	productRepo := repositories.NewProductRepository(db)
	tableRepo := repositories.NewTableRepository(db)
	svc := services.NewOrderService(orderRepo, productRepo, tableRepo, nil, nil, nil)

	tenantID, _ := uuid.NewV7()
	tableID, _ := uuid.NewV7()

	// Seed Products
	prod1ID, _ := uuid.NewV7()
	prod2ID, _ := uuid.NewV7()
	db.Create(&domain.Product{BaseModel: domain.BaseModel{ID: prod1ID}, TenantID: tenantID, Name: "Coffee", Price: 25000})
	db.Create(&domain.Product{BaseModel: domain.BaseModel{ID: prod2ID}, TenantID: tenantID, Name: "Milk", Price: 15000})

	// Open Table Session
	order, _ := svc.OpenSession(tenantID.String(), services.OpenSessionRequest{TableID: tableID.String()})

	// Add items to simulate cart manipulation
	items := []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}{
		{ProductID: prod1ID.String(), Quantity: 2}, // 50,000
		{ProductID: prod2ID.String(), Quantity: 3}, // 45,000
	}

	updatedOrder, err := svc.AddItems(tenantID.String(), order.ID.String(), "", items)

	assert.Nil(t, err)
	assert.Equal(t, float64(95000), updatedOrder.TotalPrice)
	assert.Len(t, updatedOrder.Items, 2)
	assert.Equal(t, float64(50000), updatedOrder.Items[0].SubTotal)
	
	// Checkout test
	finalOrder, checkErr := svc.Checkout(tenantID.String(), order.ID.String())
	assert.Nil(t, checkErr)
	assert.Equal(t, "Paid", finalOrder.Status)
}
