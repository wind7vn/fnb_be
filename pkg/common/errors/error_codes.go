package errors

// Common Error Codes mapped to Mobile/Frontend enum handling
const (
	ErrCodeValidationFailed      = "ERR_VALIDATION_FAILED"
	ErrCodeUnauthorized          = "ERR_UNAUTHORIZED"
	ErrCodeForbidden             = "ERR_FORBIDDEN"
	ErrCodeInternalSystemError   = "ERR_INTERNAL_SYSTEM"
	ErrCodeTenantIsolationBreach = "ERR_TENANT_ISOLATION_BREACH"

	// Business Codes
	ErrCodeUserNotFound   = "ERR_USER_NOT_FOUND"
	ErrCodeWrongPassword  = "ERR_WRONG_PASSWORD"
	ErrCodeItemOutOfStock = "ERR_ITEM_OUT_OF_STOCK"
	ErrCodeTableOccupied  = "ERR_TABLE_OCCUPIED"
)
