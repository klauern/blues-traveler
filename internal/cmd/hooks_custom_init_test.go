package cmd

import (
	"runtime"
	"strings"
	"testing"

	"github.com/klauern/blues-traveler/internal/constants"
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
			errMsg:   "path separators not allowed", // On Unix: checked by separator. On Windows: absolute path check
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
				validateSanitizeError(t, tt.name, tt.fileName, tt.errMsg, err)
			} else {
				validateSanitizeSuccess(t, tt.fileName, got, err)
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
// validateSanitizeError validates that sanitizeFileName returned the expected error
func validateSanitizeError(t *testing.T, testName, fileName, expectedErrMsg string, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("sanitizeFileName(%q) expected error containing %q, got nil", fileName, expectedErrMsg)
		return
	}

	if expectedErrMsg == "" {
		return
	}

	if strings.Contains(err.Error(), expectedErrMsg) {
		return
	}

	// Platform-specific error message handling
	isWindowsAbsoluteCase := runtime.GOOS == constants.GOOSWindows &&
		testName == "absolute path windows" &&
		strings.Contains(err.Error(), "absolute paths not allowed")
	isUnixBackslashCase := runtime.GOOS != constants.GOOSWindows &&
		testName == "absolute path windows" &&
		strings.Contains(err.Error(), "path separators not allowed")
	// On Windows, Unix absolute paths like "/etc/passwd" are not recognized as absolute
	// by filepath.IsAbs, so they trigger the path separator check instead
	isWindowsUnixPathCase := runtime.GOOS == constants.GOOSWindows &&
		testName == "absolute path unix" &&
		strings.Contains(err.Error(), "path separators not allowed")

	if !isWindowsAbsoluteCase && !isUnixBackslashCase && !isWindowsUnixPathCase {
		t.Errorf("sanitizeFileName(%q) error = %v, want error containing %q", fileName, err, expectedErrMsg)
	}
}

// validateSanitizeSuccess validates that sanitizeFileName succeeded with a valid result
func validateSanitizeSuccess(t *testing.T, fileName, got string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("sanitizeFileName(%q) unexpected error: %v", fileName, err)
		return
	}

	// Verify the result has a proper extension (case-insensitive)
	gotLower := strings.ToLower(got)
	if got != "" && !strings.HasSuffix(gotLower, ".yml") && !strings.HasSuffix(gotLower, ".yaml") {
		t.Errorf("sanitizeFileName(%q) = %q, expected .yml or .yaml extension", fileName, got)
	}
}
