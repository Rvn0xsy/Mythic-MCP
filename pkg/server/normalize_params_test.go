package server

import (
	"testing"
)

func TestNormalizeParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "plain string path",
			input:    "/tmp/newdir",
			expected: "/tmp/newdir",
		},
		{
			name:     "plain string command",
			input:    "whoami",
			expected: "whoami",
		},
		{
			name:     "plain string with spaces",
			input:    "ls -la /tmp",
			expected: "ls -la /tmp",
		},
		{
			name:     "single-key JSON with string value - the core bug",
			input:    `{"path": "/tmp/mythic-should-work-now"}`,
			expected: "/tmp/mythic-should-work-now",
		},
		{
			name:     "single-key JSON with different key",
			input:    `{"command": "whoami"}`,
			expected: "whoami",
		},
		{
			name:     "single-key JSON with number value",
			input:    `{"pid": 1234}`,
			expected: "1234",
		},
		{
			name:     "single-key JSON with permissions string",
			input:    `{"permissions": "755 /tmp/file"}`,
			expected: "755 /tmp/file",
		},
		{
			name:     "multi-key JSON passthrough",
			input:    `{"host": "192.168.1.1", "port": 8080}`,
			expected: `{"host": "192.168.1.1", "port": 8080}`,
		},
		{
			name:     "JSON array passthrough (not an object)",
			input:    `[1, 2, 3]`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "number-only string (not JSON object)",
			input:    "1234",
			expected: "1234",
		},
		{
			name:     "boolean-like plain string",
			input:    "true",
			expected: "true",
		},
		{
			name:     "quoted JSON string (not an object)",
			input:    `"hello"`,
			expected: `"hello"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeParams(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeParams(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
