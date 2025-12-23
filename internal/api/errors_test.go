package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckResultCode tests the checkResultCode helper function.
func TestCheckResultCode(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			err := checkResultCode(tt.resultCode, tt.operation)
			if tt.wantErr {
				require.Error(t, err, "checkResultCode() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "checkResultCode() error = %v, wantErr %v")
			}

			if tt.wantErr {
				assert.EqualError(t, err, tt.errMessage)
			}

		})
	}
}

// TestCheckResultCode_ReturnType tests that checkResultCode returns ResultCodeError.
func TestCheckResultCode_ReturnType(t *testing.T) {
	t.Parallel()
	err := checkResultCode("500E00", "test operation")
	require.Error(t, err, "Expected error, got nil")

	resultCodeErr := &ResultCodeError{}
	ok := errors.As(err, &resultCodeErr)
	require.Truef(t, ok, "Expected *ResultCodeError, got %T", err)

	assert.Equalf(t, "500E00", resultCodeErr.ResultCode, "Expected ResultCode '500E00', got '%s'", resultCodeErr.ResultCode)

	assert.Equalf(t, "test operation", resultCodeErr.Operation, "Expected Operation 'test operation', got '%s'", resultCodeErr.Operation)
}

// TestNewResultCodeError tests the ResultCodeError constructor.
func TestNewResultCodeError(t *testing.T) {
	t.Parallel()
	err := NewResultCodeError("400E01", "unlock doors")

	assert.Equalf(t, "400E01", err.ResultCode, "Expected ResultCode '400E01', got '%s'", err.ResultCode)

	assert.Equalf(t, "unlock doors", err.Operation, "Expected Operation 'unlock doors', got '%s'", err.Operation)

	expectedMsg := "failed to unlock doors: result code 400E01"
	assert.Equal(t, expectedMsg, err.Error())
}
