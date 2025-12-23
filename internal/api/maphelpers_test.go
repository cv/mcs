package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     map[string]any
		key       string
		wantValue string
		wantOk    bool
	}{
		{
			name:      "valid string",
			input:     map[string]any{"key": "value"},
			key:       "key",
			wantValue: "value",
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]any{"other": "value"},
			key:       "key",
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]any{"key": 123},
			key:       "key",
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "nil map",
			input:     nil,
			key:       "key",
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "empty string",
			input:     map[string]any{"key": ""},
			key:       "key",
			wantValue: "",
			wantOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getString(tt.input, tt.key)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestGetInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     map[string]any
		key       string
		wantValue int
		wantOk    bool
	}{
		{
			name:      "valid int",
			input:     map[string]any{"key": 42},
			key:       "key",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "valid float64",
			input:     map[string]any{"key": 42.0},
			key:       "key",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "float64 with decimal",
			input:     map[string]any{"key": 42.7},
			key:       "key",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]any{"other": 42},
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]any{"key": "not a number"},
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "nil map",
			input:     nil,
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getInt(tt.input, tt.key)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestGetFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     map[string]any
		key       string
		wantValue float64
		wantOk    bool
	}{
		{
			name:      "valid float64",
			input:     map[string]any{"key": 42.5},
			key:       "key",
			wantValue: 42.5,
			wantOk:    true,
		},
		{
			name:      "zero float",
			input:     map[string]any{"key": 0.0},
			key:       "key",
			wantValue: 0.0,
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]any{"other": 42.5},
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]any{"key": "not a number"},
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "nil map",
			input:     nil,
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getFloat64(tt.input, tt.key)
			assert.InDelta(t, tt.wantValue, gotValue, 0.0001)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestGetBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     map[string]any
		key       string
		wantValue bool
		wantOk    bool
	}{
		{
			name:      "valid true",
			input:     map[string]any{"key": true},
			key:       "key",
			wantValue: true,
			wantOk:    true,
		},
		{
			name:      "valid false",
			input:     map[string]any{"key": false},
			key:       "key",
			wantValue: false,
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]any{"other": true},
			key:       "key",
			wantValue: false,
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]any{"key": 1},
			key:       "key",
			wantValue: false,
			wantOk:    false,
		},
		{
			name:      "nil map",
			input:     nil,
			key:       "key",
			wantValue: false,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getBool(tt.input, tt.key)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestGetMap(t *testing.T) {
	t.Parallel()
	nested := map[string]any{"nested": "value"}

	tests := []struct {
		name   string
		input  map[string]any
		key    string
		wantOk bool
	}{
		{
			name:   "valid map",
			input:  map[string]any{"key": nested},
			key:    "key",
			wantOk: true,
		},
		{
			name:   "missing key",
			input:  map[string]any{"other": nested},
			key:    "key",
			wantOk: false,
		},
		{
			name:   "wrong type",
			input:  map[string]any{"key": "not a map"},
			key:    "key",
			wantOk: false,
		},
		{
			name:   "nil map",
			input:  nil,
			key:    "key",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getMap(tt.input, tt.key)
			assert.Equal(t, tt.wantOk, gotOk)
			if tt.wantOk {
				assert.NotNil(t, gotValue)
			}

		})
	}
}

func TestGetSlice(t *testing.T) {
	t.Parallel()
	slice := []any{"a", "b", "c"}

	tests := []struct {
		name   string
		input  map[string]any
		key    string
		wantOk bool
	}{
		{
			name:   "valid slice",
			input:  map[string]any{"key": slice},
			key:    "key",
			wantOk: true,
		},
		{
			name:   "missing key",
			input:  map[string]any{"other": slice},
			key:    "key",
			wantOk: false,
		},
		{
			name:   "wrong type",
			input:  map[string]any{"key": "not a slice"},
			key:    "key",
			wantOk: false,
		},
		{
			name:   "nil map",
			input:  nil,
			key:    "key",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getSlice(tt.input, tt.key)
			assert.Equal(t, tt.wantOk, gotOk)
			if tt.wantOk {
				assert.NotNil(t, gotValue)
			}

		})
	}
}

func TestGetMapSlice(t *testing.T) {
	t.Parallel()
	mapSlice := []any{
		map[string]any{"a": 1},
		map[string]any{"b": 2},
	}

	tests := []struct {
		name    string
		input   map[string]any
		key     string
		wantLen int
		wantOk  bool
	}{
		{
			name:    "valid map slice",
			input:   map[string]any{"key": mapSlice},
			key:     "key",
			wantLen: 2,
			wantOk:  true,
		},
		{
			name:    "empty slice",
			input:   map[string]any{"key": []any{}},
			key:     "key",
			wantLen: 0,
			wantOk:  true,
		},
		{
			name:    "missing key",
			input:   map[string]any{"other": mapSlice},
			key:     "key",
			wantLen: 0,
			wantOk:  false,
		},
		{
			name:    "slice with non-map element",
			input:   map[string]any{"key": []any{"not a map"}},
			key:     "key",
			wantLen: 0,
			wantOk:  false,
		},
		{
			name:    "wrong type",
			input:   map[string]any{"key": "not a slice"},
			key:     "key",
			wantLen: 0,
			wantOk:  false,
		},
		{
			name:    "nil map",
			input:   nil,
			key:     "key",
			wantLen: 0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getMapSlice(tt.input, tt.key)
			assert.Equal(t, tt.wantOk, gotOk)
			assert.Len(t, gotValue, tt.wantLen)
		})
	}
}

func TestGetMapFromSlice(t *testing.T) {
	t.Parallel()
	slice := []any{
		map[string]any{"a": 1},
		map[string]any{"b": 2},
		"not a map",
	}

	tests := []struct {
		name   string
		slice  []any
		index  int
		wantOk bool
	}{
		{
			name:   "valid index 0",
			slice:  slice,
			index:  0,
			wantOk: true,
		},
		{
			name:   "valid index 1",
			slice:  slice,
			index:  1,
			wantOk: true,
		},
		{
			name:   "non-map element",
			slice:  slice,
			index:  2,
			wantOk: false,
		},
		{
			name:   "negative index",
			slice:  slice,
			index:  -1,
			wantOk: false,
		},
		{
			name:   "index out of bounds",
			slice:  slice,
			index:  10,
			wantOk: false,
		},
		{
			name:   "nil slice",
			slice:  nil,
			index:  0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotOk := getMapFromSlice(tt.slice, tt.index)
			assert.Equal(t, tt.wantOk, gotOk)
			if tt.wantOk {
				assert.NotNil(t, gotValue)
			}

		})
	}
}
