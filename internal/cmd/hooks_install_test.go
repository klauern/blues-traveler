package cmd

import (
	"testing"

	"github.com/klauern/blues-traveler/internal/config"
)

func TestBuildInstallHookCommand(t *testing.T) {
	tests := []struct {
		name     string
		hookType string
		flags    installFlags
		want     string
	}{
		{
			name:     "basic command without logging",
			hookType: "security",
			flags: installFlags{
				logEnabled: false,
			},
			want: "hooks run security",
		},
		{
			name:     "command with jsonl logging",
			hookType: "security",
			flags: installFlags{
				logEnabled: true,
				logFormat:  config.LoggingFormatJSONL,
			},
			want: "hooks run security --log",
		},
		{
			name:     "command with pretty logging",
			hookType: "format",
			flags: installFlags{
				logEnabled: true,
				logFormat:  "pretty",
			},
			want: "hooks run format --log --log-format pretty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildInstallHookCommand(tt.hookType, tt.flags)
			if err != nil {
				t.Fatalf("buildInstallHookCommand() error = %v", err)
			}

			// Check if the command contains expected parts (excluding full path)
			if !contains(got, tt.want) {
				t.Errorf("buildInstallHookCommand() = %v, should contain %v", got, tt.want)
			}
		})
	}
}

func TestHandleDuplicateHookResult(t *testing.T) {
	tests := []struct {
		name   string
		result config.MergeResult
		want   bool
	}{
		{
			name: "not a duplicate",
			result: config.MergeResult{
				WasDuplicate: false,
			},
			want: false,
		},
		{
			name: "duplicate with replacement",
			result: config.MergeResult{
				WasDuplicate:  true,
				DuplicateInfo: "Replaced existing security hook",
			},
			want: false,
		},
		{
			name: "duplicate without changes",
			result: config.MergeResult{
				WasDuplicate:  true,
				DuplicateInfo: "Hook already exists",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleDuplicateHookResult(tt.result)
			if got != tt.want {
				t.Errorf("handleDuplicateHookResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInstallFlags(t *testing.T) {
	tests := []struct {
		name      string
		logFormat string
		wantErr   bool
	}{
		{
			name:      "valid jsonl format",
			logFormat: "jsonl",
			wantErr:   false,
		},
		{
			name:      "valid pretty format",
			logFormat: "pretty",
			wantErr:   false,
		},
		{
			name:      "empty format defaults to jsonl",
			logFormat: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Full testing would require mocking cli.Command
			// This is a placeholder for the test structure
			if tt.wantErr {
				// Test error cases
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findInString(s, substr)
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
