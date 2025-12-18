package api

import "fmt"

// APIError represents a general API error
type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

// EncryptionError represents an encryption error (error code 600001)
type EncryptionError struct {
	APIError
}

// TokenExpiredError represents a token expiration error (error code 600002)
type TokenExpiredError struct {
	APIError
}

// RequestInProgressError represents a request in progress error (error code 920000, extraCode 400S01)
type RequestInProgressError struct {
	APIError
}

// EngineStartLimitError represents an engine start limit error (error code 920000, extraCode 400S11)
type EngineStartLimitError struct {
	APIError
}

// NewAPIError creates a new API error
func NewAPIError(message string) *APIError {
	return &APIError{Message: message}
}

// NewEncryptionError creates a new encryption error
func NewEncryptionError() *EncryptionError {
	return &EncryptionError{APIError{Message: "Server rejected encrypted request"}}
}

// NewTokenExpiredError creates a new token expired error
func NewTokenExpiredError() *TokenExpiredError {
	return &TokenExpiredError{APIError{Message: "Token expired"}}
}

// NewRequestInProgressError creates a new request in progress error
func NewRequestInProgressError() *RequestInProgressError {
	return &RequestInProgressError{APIError{Message: "Request already in progress, please wait and try again"}}
}

// NewEngineStartLimitError creates a new engine start limit error
func NewEngineStartLimitError() *EngineStartLimitError {
	return &EngineStartLimitError{APIError{Message: "The engine can only be remotely started 2 consecutive times. Please drive the vehicle to reset the counter."}}
}

// NewInvalidRegionError creates a new invalid region error
func NewInvalidRegionError(region string) error {
	return fmt.Errorf("invalid region: %s", region)
}
