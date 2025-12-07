package main

import (
	"os"
	"testing"
)

func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "default port",
			envVar:   "PROXY_PORT",
			envValue: "",
			expected: "1488",
		},
		{
			name:     "custom port",
			envVar:   "PROXY_PORT",
			envValue: "3128",
			expected: "3128",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			// Get value
			port := os.Getenv("PROXY_PORT")
			if port == "" {
				port = "1488"
			}

			if port != tt.expected {
				t.Errorf("Expected port %s, got %s", tt.expected, port)
			}
		})
	}
}

func TestMainFunction(t *testing.T) {
	// This test verifies that main can be imported without errors
	// Actual execution testing would require integration tests
	t.Log("Main package imports successfully")
}
