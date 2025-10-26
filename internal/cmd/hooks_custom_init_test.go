package cmd

import (
	"testing"
)

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid simple filename",
			fileName: "test",
			wantErr:  false,
		},
		{
			name:     "valid filename with extension",
			fileName: "test.yml",
			wantErr:  false,
		},
		{
			name:     "valid filename with yaml extension",
			fileName: "test.yaml",
			wantErr:  false,
		},
		{
			name:     "empty filename",
			fileName: "",
			wantErr:  true,
			errMsg:   "invalid filename",
		},
		{
			name:     "dot filename",
			fileName: ".",
			wantErr:  true,
			errMsg:   "invalid filename",
		},
		{
			name:     "double dot filename",
			fileName: "..",
			wantErr:  true,
			errMsg:   "invalid filename",
		},
		{
			name:     "absolute path unix",
			fileName: "/etc/passwd",
			wantErr:  true,
			errMsg:   "absolute paths not allowed",
		},
		{
			name:     "absolute path windows",
			fileName: "C:\\Windows\\System32",
			wantErr:  true,
			errMsg:   "path separators not allowed", // On Unix, backslash is caught by separator check
		},
		{
			name:     "path traversal with slash",
			fileName: "../../../etc/passwd",
			wantErr:  true,
			errMsg:   "path separators not allowed",
		},
		{
			name:     "path traversal with backslash",
			fileName: "..\\..\\..\\Windows\\System32",
			wantErr:  true,
			errMsg:   "path separators not allowed",
		},
		{
			name:     "forward slash in name",
			fileName: "test/config",
			wantErr:  true,
			errMsg:   "path separators not allowed",
		},
		{
			name:     "backslash in name",
			fileName: "test\\config",
			wantErr:  true,
			errMsg:   "path separators not allowed",
		},
		{
			name:     "relative path with dot",
			fileName: "./test",
			wantErr:  true,
			errMsg:   "path separators not allowed",
		},
		{
			name:     "hidden file with dot prefix",
			fileName: ".hidden",
			wantErr:  false, // dot prefix is allowed, just not "." or ".." alone
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeFileName(tt.fileName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("sanitizeFileName(%q) expected error containing %q, got nil", tt.fileName, tt.errMsg)
				} else if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("sanitizeFileName(%q) error = %v, want error containing %q", tt.fileName, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("sanitizeFileName(%q) unexpected error: %v", tt.fileName, err)
				}
				// Verify the result has a proper extension
				if got != "" && !hasSuffix(got, ".yml") && !hasSuffix(got, ".yaml") {
					t.Errorf("sanitizeFileName(%q) = %q, expected .yml or .yaml extension", tt.fileName, got)
				}
			}
		})
	}
}

func TestSanitizeFileName_AddsExtension(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", "test.yml"},
		{"config", "config.yml"},
		{"my-hook", "my-hook.yml"},
		{"test.yml", "test.yml"},
		{"test.yaml", "test.yaml"},
		{"test.YML", "test.YML"},   // preserves case
		{"test.YAML", "test.YAML"}, // preserves case
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := sanitizeFileName(tt.input)
			if err != nil {
				t.Fatalf("sanitizeFileName(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("sanitizeFileName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstringInString(s, substr)
}

func findSubstringInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	// Case-insensitive comparison for extensions
	sLower := toLower(s[len(s)-len(suffix):])
	suffixLower := toLower(suffix)
	return sLower == suffixLower
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
