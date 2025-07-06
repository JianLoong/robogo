package util

import (
	"testing"
)

func TestRequireArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []interface{}
		min     int
		wantErr bool
	}{
		{
			name:    "valid args",
			args:    []interface{}{"test", 123},
			min:     1,
			wantErr: false,
		},
		{
			name:    "exact args",
			args:    []interface{}{"test"},
			min:     1,
			wantErr: false,
		},
		{
			name:    "insufficient args",
			args:    []interface{}{},
			min:     1,
			wantErr: true,
		},
		{
			name:    "nil args",
			args:    nil,
			min:     1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireArgs(tt.args, tt.min)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireString(t *testing.T) {
	tests := []struct {
		name    string
		args    []interface{}
		index   int
		want    string
		wantErr bool
	}{
		{
			name:    "valid string",
			args:    []interface{}{"test"},
			index:   0,
			want:    "test",
			wantErr: false,
		},
		{
			name:    "insufficient args",
			args:    []interface{}{},
			index:   0,
			want:    "",
			wantErr: true,
		},
		{
			name:    "wrong type",
			args:    []interface{}{123},
			index:   0,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RequireString(tt.args, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RequireString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequireInt(t *testing.T) {
	tests := []struct {
		name    string
		args    []interface{}
		index   int
		want    int
		wantErr bool
	}{
		{
			name:    "valid int",
			args:    []interface{}{123},
			index:   0,
			want:    123,
			wantErr: false,
		},
		{
			name:    "valid float",
			args:    []interface{}{123.0},
			index:   0,
			want:    123,
			wantErr: false,
		},
		{
			name:    "valid string number",
			args:    []interface{}{"123"},
			index:   0,
			want:    123,
			wantErr: false,
		},
		{
			name:    "invalid string",
			args:    []interface{}{"abc"},
			index:   0,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RequireInt(tt.args, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RequireInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name: "with field",
			err: ValidationError{
				Field:   "test_field",
				Message: "test message",
				Value:   "test_value",
			},
			expected: "validation error for field 'test_field': test message (value: test_value)",
		},
		{
			name: "without field",
			err: ValidationError{
				Message: "test message",
				Value:   "test_value",
			},
			expected: "validation error: test message (value: test_value)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}
