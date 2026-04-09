package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/wind7vn/fnb_be/internal/core/domain"
	"github.com/wind7vn/fnb_be/internal/core/ports"
	"github.com/wind7vn/fnb_be/pkg/common/errors"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo   ports.OrderRepository
	productRepo ports.ProductRepository
	tableRepo   ports.TableRepository
	pubSub      *PubSubService
	system      *SystemService
}

func NewOrderService(order ports.OrderRepository, product ports.ProductRepository, table ports.TableRepository, ps *PubSubService, sys *SystemService) *OrderService {
	return &OrderService{orderRepo: order, productRepo: product, tableRepo: table, pubSub: ps, system: sys}
}

type OpenSessionRequest struct {
	TableID   string `json:"table_id"`
	SessionID string `json:"session_id"` // Identifies the guest session
}

func (s *OrderService) GetActiveOrderByTable(tenantID string, tableID string) (*domain.Order, *errors.AppError) {
	_, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Tenant ID invalid", err)
	}

	_, err = uuid.Parse(tableID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Table ID invalid", err)
	}

	order, err := s.orderRepo.FindActiveByTable(tableID, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Active order not found for this table", err)
		}
		return nil, errors.NewInternalServer(err)
	}

	return order, nil
}

func (s *OrderService) GetActiveOrders(tenantID string) ([]domain.Order, *errors.AppError) {
	_, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Tenant ID invalid", err)
	}

	orders, err := s.orderRepo.FindAllActive(tenantID)
	if err != nil {
		return nil, errors.NewInternalServer(err)
	}

	return orders, nil
}

func (s *OrderService) OpenSession(tenantID string, req OpenSessionRequest) (*domain.Order, *errors.AppError) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Tenant ID invalid", err)
	}

	tabID, err := uuid.Parse(req.TableID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Table ID invalid", err)
	}

	// Check if already exist active order
	_, err = s.orderRepo.FindActiveByTable(req.TableID, tenantID)
	if err == nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Bàn đang có người ngồi", nil)
	} else if err != gorm.ErrRecordNotFound {
		return nil, errors.NewInternalServer(err)
	}

	// Generate a unique code
	codeStr := "ORD-" + uuid.New().String()[:8]

	order := &domain.Order{
		TenantID: tid,
		TableID:  &tabID,
		Code:     codeStr,
		Status:   domain.OrderStatusPending,
		TotalPrice: 0,
		SessionID: req.SessionID,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	// Update table status to Occupied
	table, err := s.tableRepo.FindByID(req.TableID, tenantID)
	if err == nil {
		table.Status = "Occupied"
		errUpdate := s.tableRepo.Update(table)
		if errUpdate != nil {
			logger.Log.Error("Failed to update table status: " + errUpdate.Error())
		}
	} else {
		logger.Log.Error("Failed to find table to update status: " + err.Error())
	}

	if s.system != nil {
		go s.system.LogAction(tenantID, "", "Staff", "CREATE_ORDER", "orders", order.ID.String(), map[string]interface{}{"table_id": req.TableID})
	}

	return order, nil
}

func (s *OrderService) AddItems(tenantID string, orderID string, sessionID string, items []struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}) (*domain.Order, *errors.AppError) {
	order, err := s.orderRepo.FindByID(orderID, tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Hoá đơn không hợp lệ", err)
	}

	if order.Status == "Paid" {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Hoá đơn đã thanh toán, không thể thêm món", nil)
	}

	// If sessionID is provided (which means it's a Guest), it must match the order's session ID (unless the order has no session ID, which means it was created by staff)
	if sessionID != "" {
		if order.SessionID == "" {
			order.SessionID = sessionID
		} else if sessionID != order.SessionID {
			return nil, errors.NewUnauthorized(errors.ErrCodeUnauthorized, "Bàn đang có khách khác order, bạn không thể thêm món", nil)
		}
	}

	// Fetch product prices and build cart slice
	for _, itemReq := range items {
		product, err := s.productRepo.FindByID(itemReq.ProductID, tenantID)
		if err != nil {
			return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Sản phẩm không hợp lệ: "+itemReq.ProductID, err)
		}

		pid := product.ID

		orderItem := domain.OrderItem{
			TenantID:  order.TenantID,
			OrderID:   order.ID,
			ProductID: pid,
			Quantity:  itemReq.Quantity,
			Price:     product.Price,
			SubTotal:  product.Price * float64(itemReq.Quantity),
			Status:    "Pending",
		}
		
		order.Items = append(order.Items, orderItem)
		order.TotalPrice += orderItem.SubTotal
	}

	if err := s.orderRepo.Update(order); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	// Đảm bảo Table luôn Occupied nếu có đơn thêm món
	if order.TableID != nil {
		table, errTable := s.tableRepo.FindByID(order.TableID.String(), tenantID)
		if errTable == nil && table.Status != "Occupied" {
			table.Status = "Occupied"
			_ = s.tableRepo.Update(table)
		}
	}

	if s.pubSub != nil {
		_ = s.pubSub.PublishEvent("KDS_EVENTS", EventPayload{
			TenantID: tenantID,
			Type:     domain.EventOrderCreated,
			Data:     order, // Send full updated order roughly
		})
	}

	return order, nil
}

func (s *OrderService) GenerateGuestToken(tenantID string, tableID string) (string, *errors.AppError) {
	// Only valid if there's an active order
	_, err := s.orderRepo.FindActiveByTable(tableID, tenantID)
	if err != nil {
		return "", errors.NewBadRequest(errors.ErrCodeValidationFailed, "Chưa mở bàn, không thể tạo mã QR", err)
	}

	claims := jwt.MapClaims{
		"tenant_id": tenantID,
		"table_id":  tableID,
		"role":      "Guest",
		"exp":       time.Now().Add(4 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, jwtErr := token.SignedString([]byte(config.AppConfig.JWTSecret))
	
	if jwtErr != nil {
		return "", errors.NewInternalServer(jwtErr)
	}

	return tokenStr, nil
}

func (s *OrderService) Checkout(tenantID string, orderID string) (*domain.Order, *errors.AppError) {
	order, err := s.orderRepo.FindByID(orderID, tenantID)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Hoá đơn không hợp lệ", err)
	}

	if order.Status == "Paid" {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Hoá đơn đã thanh toán", nil)
	}

	order.Status = domain.OrderStatusPaid
	if err := s.orderRepo.UpdateStatus(order.ID.String(), domain.OrderStatusPaid); err != nil {
		return nil, errors.NewInternalServer(err)
	}

	// Update table status to Available
	if order.TableID != nil {
		table, err := s.tableRepo.FindByID(order.TableID.String(), tenantID)
		if err == nil {
			table.Status = "Available"
			_ = s.tableRepo.Update(table)
		}
	}

	return order, nil
}

func (s *OrderService) UpdateItemStatus(tenantID string, itemID string, status string) *errors.AppError {
	// Fast path DB update
	err := s.orderRepo.UpdateItemStatus(itemID, status)
	if err != nil {
		return errors.NewInternalServer(err)
	}

	if s.pubSub != nil {
		_ = s.pubSub.PublishEvent("KDS_EVENTS", EventPayload{
			TenantID: tenantID,
			Type:     domain.EventItemStatusUpdated,
			Data: map[string]string{
				"item_id": itemID,
				"status":  status,
			},
		})
	}

	if status == domain.OrderStatusReady && s.system != nil {
		s.system.CreateNotification(
			tenantID,
			"", // Assuming broadcast to tenant initially
			"Món ăn đã xong",
			"Một món ăn trong đơn hàng đã hoàn tất",
			"KDS_ITEM_READY",
			map[string]interface{}{"item_id": itemID},
		)
	}

	return nil
}
