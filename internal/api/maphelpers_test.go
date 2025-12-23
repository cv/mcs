package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]interface{}
		key       string
		wantValue string
		wantOk    bool
	}{
		{
			name:      "valid string",
			input:     map[string]interface{}{"key": "value"},
			key:       "key",
			wantValue: "value",
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]interface{}{"other": "value"},
			key:       "key",
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]interface{}{"key": 123},
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
			input:     map[string]interface{}{"key": ""},
			key:       "key",
			wantValue: "",
			wantOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := getString(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantValue, gotValue, "getString() value = %v, want %v")
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getString() ok = %v, want %v")
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]interface{}
		key       string
		wantValue int
		wantOk    bool
	}{
		{
			name:      "valid int",
			input:     map[string]interface{}{"key": 42},
			key:       "key",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "valid float64",
			input:     map[string]interface{}{"key": 42.0},
			key:       "key",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "float64 with decimal",
			input:     map[string]interface{}{"key": 42.7},
			key:       "key",
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]interface{}{"other": 42},
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]interface{}{"key": "not a number"},
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
			gotValue, gotOk := getInt(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantValue, gotValue, "getInt() value = %v, want %v")
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getInt() ok = %v, want %v")
		})
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]interface{}
		key       string
		wantValue float64
		wantOk    bool
	}{
		{
			name:      "valid float64",
			input:     map[string]interface{}{"key": 42.5},
			key:       "key",
			wantValue: 42.5,
			wantOk:    true,
		},
		{
			name:      "zero float",
			input:     map[string]interface{}{"key": 0.0},
			key:       "key",
			wantValue: 0.0,
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]interface{}{"other": 42.5},
			key:       "key",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]interface{}{"key": "not a number"},
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
			gotValue, gotOk := getFloat64(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantValue, gotValue, "getFloat64() value = %v, want %v")
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getFloat64() ok = %v, want %v")
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]interface{}
		key       string
		wantValue bool
		wantOk    bool
	}{
		{
			name:      "valid true",
			input:     map[string]interface{}{"key": true},
			key:       "key",
			wantValue: true,
			wantOk:    true,
		},
		{
			name:      "valid false",
			input:     map[string]interface{}{"key": false},
			key:       "key",
			wantValue: false,
			wantOk:    true,
		},
		{
			name:      "missing key",
			input:     map[string]interface{}{"other": true},
			key:       "key",
			wantValue: false,
			wantOk:    false,
		},
		{
			name:      "wrong type",
			input:     map[string]interface{}{"key": 1},
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
			gotValue, gotOk := getBool(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantValue, gotValue, "getBool() value = %v, want %v")
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getBool() ok = %v, want %v")
		})
	}
}

func TestGetMap(t *testing.T) {
	nested := map[string]interface{}{"nested": "value"}

	tests := []struct {
		name   string
		input  map[string]interface{}
		key    string
		wantOk bool
	}{
		{
			name:   "valid map",
			input:  map[string]interface{}{"key": nested},
			key:    "key",
			wantOk: true,
		},
		{
			name:   "missing key",
			input:  map[string]interface{}{"other": nested},
			key:    "key",
			wantOk: false,
		},
		{
			name:   "wrong type",
			input:  map[string]interface{}{"key": "not a map"},
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
			gotValue, gotOk := getMap(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getMap() ok = %v, want %v")
			if tt.wantOk && gotValue == nil {
				t.Errorf("getMap() returned nil when ok=true")
			}
		})
	}
}

func TestGetSlice(t *testing.T) {
	slice := []interface{}{"a", "b", "c"}

	tests := []struct {
		name   string
		input  map[string]interface{}
		key    string
		wantOk bool
	}{
		{
			name:   "valid slice",
			input:  map[string]interface{}{"key": slice},
			key:    "key",
			wantOk: true,
		},
		{
			name:   "missing key",
			input:  map[string]interface{}{"other": slice},
			key:    "key",
			wantOk: false,
		},
		{
			name:   "wrong type",
			input:  map[string]interface{}{"key": "not a slice"},
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
			gotValue, gotOk := getSlice(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getSlice() ok = %v, want %v")
			if tt.wantOk && gotValue == nil {
				t.Errorf("getSlice() returned nil when ok=true")
			}
		})
	}
}

func TestGetMapSlice(t *testing.T) {
	mapSlice := []interface{}{
		map[string]interface{}{"a": 1},
		map[string]interface{}{"b": 2},
	}

	tests := []struct {
		name    string
		input   map[string]interface{}
		key     string
		wantLen int
		wantOk  bool
	}{
		{
			name:    "valid map slice",
			input:   map[string]interface{}{"key": mapSlice},
			key:     "key",
			wantLen: 2,
			wantOk:  true,
		},
		{
			name:    "empty slice",
			input:   map[string]interface{}{"key": []interface{}{}},
			key:     "key",
			wantLen: 0,
			wantOk:  true,
		},
		{
			name:    "missing key",
			input:   map[string]interface{}{"other": mapSlice},
			key:     "key",
			wantLen: 0,
			wantOk:  false,
		},
		{
			name:    "slice with non-map element",
			input:   map[string]interface{}{"key": []interface{}{"not a map"}},
			key:     "key",
			wantLen: 0,
			wantOk:  false,
		},
		{
			name:    "wrong type",
			input:   map[string]interface{}{"key": "not a slice"},
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
			gotValue, gotOk := getMapSlice(tt.input, tt.key)
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getMapSlice() ok = %v, want %v")
			assert.EqualValuesf(t, tt.wantLen, len(gotValue), "getMapSlice() len = %v, want %v")
		})
	}
}

func TestGetMapFromSlice(t *testing.T) {
	slice := []interface{}{
		map[string]interface{}{"a": 1},
		map[string]interface{}{"b": 2},
		"not a map",
	}

	tests := []struct {
		name   string
		slice  []interface{}
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
			gotValue, gotOk := getMapFromSlice(tt.slice, tt.index)
			assert.EqualValuesf(t, tt.wantOk, gotOk, "getMapFromSlice() ok = %v, want %v")
			if tt.wantOk && gotValue == nil {
				t.Errorf("getMapFromSlice() returned nil when ok=true")
			}
		})
	}
}
