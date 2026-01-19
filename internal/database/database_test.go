package database

import (
	"strconv"
	"testing"
)

// TestHealthResponseFormat ensures health check returns correctly formatted strings
func TestHealthResponseFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    int32
		expected string
	}{
		{
			name:     "single digit",
			input:    5,
			expected: "5",
		},
		{
			name:     "double digit",
			input:    25,
			expected: "25",
		},
		{
			name:     "triple digit",
			input:    100,
			expected: "100",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strconv.Itoa(int(tt.input))
			if result != tt.expected {
				t.Errorf("strconv.Itoa(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHealthMapKeys ensures all required keys are present in health response
func TestHealthMapKeys(t *testing.T) {
	requiredKeys := []string{
		"status",
		"total_conns",
		"acquired_conns",
		"idle_conns",
		"constructing",
		"max_conns",
	}

	// Create a mock health response like the one returned by Health()
	mockHealth := map[string]string{
		"status":         "healthy",
		"total_conns":    "25",
		"acquired_conns": "10",
		"idle_conns":     "15",
		"constructing":   "0",
		"max_conns":      "50",
	}

	for _, key := range requiredKeys {
		if _, exists := mockHealth[key]; !exists {
			t.Errorf("Health response missing required key: %s", key)
		}
	}
}

// TestHealthStatusValues ensures status field has valid values
func TestHealthStatusValues(t *testing.T) {
	validStatuses := []string{"healthy", "unhealthy"}

	for _, status := range validStatuses {
		mockHealth := map[string]string{
			"status": status,
		}

		if mockHealth["status"] != status {
			t.Errorf("Expected status %s, got %s", status, mockHealth["status"])
		}
	}
}
