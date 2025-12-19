package api

import (
	"testing"
)

// TestCheckResultCode tests the checkResultCode helper function
func TestCheckResultCode(t *testing.T) {
	tests := []struct {
		name       string
		resultCode string
		operation  string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "Success code returns nil",
			resultCode: ResultCodeSuccess,
			operation:  "test operation",
			wantErr:    false,
		},
		{
			name:       "Non-success code returns error",
			resultCode: "500E00",
			operation:  "test operation",
			wantErr:    true,
			errMessage: "failed to test operation: result code 500E00",
		},
		{
			name:       "Empty result code returns error",
			resultCode: "",
			operation:  "test operation",
			wantErr:    true,
			errMessage: "failed to test operation: result code ",
		},
		{
			name:       "Custom operation message",
			resultCode: "400E01",
			operation:  "lock doors",
			wantErr:    true,
			errMessage: "failed to lock doors: result code 400E01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkResultCode(tt.resultCode, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkResultCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMessage {
				t.Errorf("checkResultCode() error message = %v, want %v", err.Error(), tt.errMessage)
			}
		})
	}
}

// TestCheckResultCode_ReturnType tests that checkResultCode returns ResultCodeError
func TestCheckResultCode_ReturnType(t *testing.T) {
	err := checkResultCode("500E00", "test operation")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	resultCodeErr, ok := err.(*ResultCodeError)
	if !ok {
		t.Fatalf("Expected *ResultCodeError, got %T", err)
	}

	if resultCodeErr.ResultCode != "500E00" {
		t.Errorf("Expected ResultCode '500E00', got '%s'", resultCodeErr.ResultCode)
	}

	if resultCodeErr.Operation != "test operation" {
		t.Errorf("Expected Operation 'test operation', got '%s'", resultCodeErr.Operation)
	}
}

// TestNewResultCodeError tests the ResultCodeError constructor
func TestNewResultCodeError(t *testing.T) {
	err := NewResultCodeError("400E01", "unlock doors")

	if err.ResultCode != "400E01" {
		t.Errorf("Expected ResultCode '400E01', got '%s'", err.ResultCode)
	}

	if err.Operation != "unlock doors" {
		t.Errorf("Expected Operation 'unlock doors', got '%s'", err.Operation)
	}

	expectedMsg := "failed to unlock doors: result code 400E01"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
