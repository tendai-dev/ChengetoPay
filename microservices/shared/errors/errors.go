package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ErrorCode string

const (
	ErrCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeInternal          ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout           ErrorCode = "TIMEOUT"
	ErrCodeRateLimit         ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"
	ErrCodeInvalidTransaction ErrorCode = "INVALID_TRANSACTION"
)

type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	StatusCode int                    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}
}

func NewValidationError(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

func NewNotFoundError(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewUnauthorizedError(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NewConflictError(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

func NewInternalError(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

func NewServiceUnavailableError(service string) *AppError {
	return New(ErrCodeServiceUnavailable, fmt.Sprintf("%s service unavailable", service), http.StatusServiceUnavailable)
}

func NewTimeoutError(operation string) *AppError {
	return New(ErrCodeTimeout, fmt.Sprintf("%s operation timed out", operation), http.StatusRequestTimeout)
}

func NewRateLimitError() *AppError {
	return New(ErrCodeRateLimit, "Rate limit exceeded", http.StatusTooManyRequests)
}

func NewInsufficientFundsError(available, required float64, currency string) *AppError {
	err := New(ErrCodeInsufficientFunds, "Insufficient funds", http.StatusPaymentRequired)
	err.Details = map[string]interface{}{
		"available": available,
		"required":  required,
		"currency":  currency,
	}
	return err
}

func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

func (e *AppError) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	json.NewEncoder(w).Encode(e)
}

func HandleError(w http.ResponseWriter, err error, requestID string) {
	if appErr, ok := err.(*AppError); ok {
		appErr.WithRequestID(requestID).Send(w)
		return
	}

	// Default to internal error
	NewInternalError(err.Error()).WithRequestID(requestID).Send(w)
}
