package cursor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateWrapper(t *testing.T) {
	config := WrapperConfig{
		HookKey:     "security",
		CursorEvent: BeforeShellExecution,
		Matcher:     ".*\\.go$",
		BinaryPath:  "/usr/local/bin/blues-traveler",
		Description: "Security check for dangerous commands",
	}

	wrapper, err := GenerateWrapper(config)
	if err != nil {
		t.Fatalf("GenerateWrapper failed: %v", err)
	}

	// Check that the wrapper contains key elements
	required := []string{
		"#!/bin/bash",
		"security",
		"beforeShellExecution",
		"--cursor-mode",
		"jq -r",
		".*\\.go$",
	}

	for _, req := range required {
		if !strings.Contains(wrapper, req) {
			t.Errorf("Generated wrapper missing required content: %q", req)
		}
	}
}

func TestGenerateWrapperWithoutMatcher(t *testing.T) {
	config := WrapperConfig{
		HookKey:     "debug",
		CursorEvent: Stop,
		Matcher:     "",
		BinaryPath:  "/usr/local/bin/blues-traveler",
		Description: "Debug logging",
	}

	wrapper, err := GenerateWrapper(config)
	if err != nil {
		t.Fatalf("GenerateWrapper failed: %v", err)
	}

	// Check that the wrapper doesn't include matcher logic
	if strings.Contains(wrapper, "matcher=") {
		t.Error("Wrapper should not contain matcher logic when no matcher is provided")
	}

	// Should still contain core functionality
	required := []string{
		"#!/bin/bash",
		"debug",
		"stop",
		"--cursor-mode",
	}

	for _, req := range required {
		if !strings.Contains(wrapper, req) {
			t.Errorf("Generated wrapper missing required content: %q", req)
		}
	}
}

func TestWrapperScriptPath(t *testing.T) {
	path, err := WrapperScriptPath("security", BeforeShellExecution)
	if err != nil {
		t.Fatalf("WrapperScriptPath failed: %v", err)
	}

	// Get the actual home directory for comparison
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	// Verify path starts with actual home directory, not literal "~"
	if !strings.HasPrefix(path, home) {
		t.Errorf("Path should start with home directory %q, got: %s", home, path)
	}

	// Verify path does NOT contain literal "~"
	if strings.Contains(path, "/~/") || strings.HasSuffix(filepath.Dir(path), "~") {
		t.Errorf("Path should not contain literal tilde, got: %s", path)
	}

	path = filepath.ToSlash(path)

	if !strings.Contains(path, ".cursor/hooks") {
		t.Errorf("Path should contain .cursor/hooks, got: %s", path)
	}

	if !strings.Contains(path, "security") {
		t.Errorf("Path should contain hook key 'security', got: %s", path)
	}

	if !strings.Contains(path, BeforeShellExecution) {
		t.Errorf("Path should contain event name, got: %s", path)
	}

	// Verify the complete expected path structure
	expectedSuffix := filepath.Join(".cursor", "hooks", "blues-traveler-security-"+BeforeShellExecution+".sh")
	if !strings.HasSuffix(path, filepath.ToSlash(expectedSuffix)) {
		t.Errorf("Path should end with %q, got: %s", expectedSuffix, path)
	}
}
