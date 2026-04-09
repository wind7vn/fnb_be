package errors

// AppError represents a unified sanitized error struct destined for the frontend JSON payload.
// It deliberately omits raw SQL syntax errors or Stack Traces.
type AppError struct {
	Success    bool   `json:"success"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"` // HTTP status
	RawError   error  `json:"-"`           // Internal tracking (omitted from JSON)
}

func (e *AppError) Error() string {
	return e.Message
}

// Global Factory Functions

func NewBadRequest(code, message string, raw error) *AppError {
	return &AppError{
		Success:    false,
		ErrorCode:  code,
		Message:    message,
		StatusCode: 400,
		RawError:   raw,
	}
}

func NewUnauthorized(code, message string, raw error) *AppError {
	return &AppError{
		Success:    false,
		ErrorCode:  code,
		Message:    message,
		StatusCode: 401,
		RawError:   raw,
	}
}

func NewForbidden(code, message string, raw error) *AppError {
	return &AppError{
		Success:    false,
		ErrorCode:  code,
		Message:    message,
		StatusCode: 403,
		RawError:   raw,
	}
}

func NewInternalServer(raw error) *AppError {
	return &AppError{
		Success:    false,
		ErrorCode:  ErrCodeInternalSystemError,
		Message:    "Hệ thống đang gặp sự cố. Vui lòng thử lại sau.",
		StatusCode: 500,
		RawError:   raw,
	}
}
