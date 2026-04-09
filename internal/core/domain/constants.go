package domain

// Roles
const (
	RoleSuperadmin = "Superadmin"
	RoleAdmin      = "Admin"
	RoleOwner      = "Owner"
	RoleManager    = "Manager"
	RoleStaff      = "Staff"
	RoleUser       = "User" // Or whatever role exist
)

// Table Statuses
const (
	TableStatusAvailable = "Available"
	TableStatusOccupied  = "Occupied"
)

// Order Statuses
const (
	OrderStatusPending   = "Pending"
	OrderStatusCooking   = "Cooking"
	OrderStatusReady     = "Ready"
	OrderStatusCompleted = "Completed"
	OrderStatusCancelled = "Cancelled"
	OrderStatusPaid      = "Paid"
)

// Event Types (WebSocket/PubSub)
const (
	EventItemStatusUpdated   = "ITEM_STATUS_UPDATED"
	EventOrderCreated        = "ORDER_CREATED"
	EventOrderUpdated        = "ORDER_UPDATED"
	EventTableCallStaff      = "TABLE_CALL_STAFF"
)

