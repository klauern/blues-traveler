package cursor

import (
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

func TestGenerateWrapperWithWildcardMatcher(t *testing.T) {
	config := WrapperConfig{
		HookKey:     "security",
		CursorEvent: BeforeShellExecution,
		Matcher:     "*",
		BinaryPath:  "/usr/local/bin/blues-traveler",
		Description: "Security check with wildcard matcher",
	}

	wrapper, err := GenerateWrapper(config)
	if err != nil {
		t.Fatalf("GenerateWrapper failed: %v", err)
	}

	// Check that the wrapper converts * to .* for valid ERE
	if !strings.Contains(wrapper, `if [[ "$matcher" == "*" ]]; then`) {
		t.Error("Wrapper should contain wildcard conversion logic")
	}

	if !strings.Contains(wrapper, `matcher=".*"`) {
		t.Error("Wrapper should convert * to .* for valid regex")
	}

	// Verify the matcher variable is set correctly
	if !strings.Contains(wrapper, `matcher="*"`) {
		t.Error("Wrapper should set initial matcher value")
	}

	// Should still contain core functionality
	required := []string{
		"#!/bin/bash",
		"security",
		"beforeShellExecution",
		"--cursor-mode",
		"grep -E",
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

	// Normalize path separators for cross-platform testing
	normalizedPath := filepath.ToSlash(path)

	if !strings.Contains(normalizedPath, ".cursor/hooks") {
		t.Errorf("Path should contain .cursor/hooks, got: %s", normalizedPath)
	}

	if !strings.Contains(normalizedPath, "security") {
		t.Errorf("Path should contain hook key 'security', got: %s", normalizedPath)
	}

	if !strings.Contains(normalizedPath, BeforeShellExecution) {
		t.Errorf("Path should contain event name, got: %s", normalizedPath)
	}
}
