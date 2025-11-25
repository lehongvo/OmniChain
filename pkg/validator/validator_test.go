package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Age      int    `json:"age" validate:"min=18,max=100"`
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Email:    "test@example.com",
				Password: "password123",
				Age:      25,
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Email:    "invalid-email",
				Password: "password123",
				Age:      25,
			},
			wantErr: true,
		},
		{
			name: "password too short",
			input: TestStruct{
				Email:    "test@example.com",
				Password: "short",
				Age:      25,
			},
			wantErr: true,
		},
		{
			name: "age too low",
			input: TestStruct{
				Email:    "test@example.com",
				Password: "password123",
				Age:      15,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateStruct(tt.input)
			if tt.wantErr {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

