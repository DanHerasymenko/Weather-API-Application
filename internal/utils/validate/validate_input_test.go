package validate

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsValidCity(t *testing.T) {
	tests := []struct {
		name string
		city string
		want bool
	}{
		{"empty", "", false},
		{"invalid", "Kyiv", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidCity(tt.city)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"empty", "", false},
		{"invalid", "@rg@", false},
		{"invalid", "email.com", false},
		{"valid", "somemail@test.com", true},
		{"plus", "user+tag@example.co.uk", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidFrequency(t *testing.T) {
	tests := []struct {
		name      string
		frequency string
		want      bool
	}{
		{"empty", "", false},
		{"invalid", "wrgeg", false},
		{"valid", "Hourly", true},
		{"valid", "DAILY", true},
		{"plus", "hourly", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidFrequency(tt.frequency)
			require.Equal(t, tt.want, got)
		})
	}
}
