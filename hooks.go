package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/brads3290/cchooks"
)

// Built-in hook plugin implementations are registered in plugin.go via init().
// This file now only contains the concrete hook runner functions invoked by the registered plugins.

// isPluginEnabled checks (project first, then global) settings to see if a plugin is enabled.
// Defaults to enabled if settings cannot be loaded or plugin key absent.
func isPluginEnabled(pluginKey string) bool {
	// Project settings
	if projectPath, err := getSettingsPath(false); err == nil {
		if s, err := loadSettings(projectPath); err == nil {
			if !s.IsPluginEnabled(pluginKey) {
				return false
			}
		}
	}
	// Global settings fallback
	if globalPath, err := getSettingsPath(true); err == nil {
		if s, err := loadSettings(globalPath); err == nil {
			if !s.IsPluginEnabled(pluginKey) {
				return false
			}
		}
	}
	return true
}

// runSecurityHook implements security blocking logic (enhanced for macOS)
// Strategy:
//  1. Parse Bash command
//  2. Run a series of detectors (token based + regex patterns)
//  3. Block immediately on high‑risk destructive / persistence / system reconfiguration ops
//  4. Provide specific rationale to aid user correction
func runSecurityHook() error {
	if !isPluginEnabled("security") {
		fmt.Println("Security plugin disabled - skipping")
		return nil
	}

	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			if event.ToolName != "Bash" {
				return cchooks.Approve()
			}

			bash, err := event.AsBash()
			if err != nil {
				return cchooks.Block("failed to parse bash command")
			}

			cmdLower := strings.ToLower(bash.Command)
			tokens := strings.Fields(cmdLower)

			// 1. High‑risk pattern list (simple substring)
			staticSubstrings := []string{
				"dd if=",          // raw disk writing
				"mkfs",            // filesystem creation
				"> /dev/",         // redirect into device nodes
				"sudo rm",         // elevated deletion
				"chmod -r 777 /",  // broad perms at root
				"chown -r",        // recursive ownership change (validate later)
				"shutdown -h now", // immediate shutdown
				"shutdown -r now",
				"nvram -c", // clearing NVRAM (EFI vars)
			}
			for _, s := range staticSubstrings {
				if strings.Contains(cmdLower, s) {
					return cchooks.Block(fmt.Sprintf("blocked dangerous command pattern: %s", s))
				}
			}

			// 2. macOS specific critical command regexes
			regexes := map[string]*regexp.Regexp{
				"disk erase / format (diskutil)": regexp.MustCompile(`\bdiskutil\s+(erase(?:disk|volume)|apfs\s+erase)`),
				"asr restore":                    regexp.MustCompile(`\basr\s+restore\b`),
				"csrutil modification":           regexp.MustCompile(`\bcsrutil\b`),
				"gatekeeper disable (spctl)":     regexp.MustCompile(`\bspctl\b.*--master-disable`),
				"launchctl service removal":      regexp.MustCompile(`\blaunchctl\b.*\b(remove|bootout)\b`),
				"systemsetup change":             regexp.MustCompile(`\bsystemsetup\b\s+-set`),
				"host/network config change":     regexp.MustCompile(`\b(scutil|networksetup)\b\s+--?set`),
				"TCC db direct write":            regexp.MustCompile(`sqlite3\s+.*TCC\.db`),
				"keychain dump":                  regexp.MustCompile(`\bsecurity\s+dump-keychain\b`),
			}
			for label, rx := range regexes {
				if rx.MatchString(cmdLower) {
					return cchooks.Block("blocked high-risk macOS command: " + label)
				}
			}

			// 3. Destructive rm heuristic
			if blocked, reason := detectDangerousRm(tokens); blocked {
				return cchooks.Block(reason)
			}

			// 4. Potential full-volume deletion / wiping
			if blocked, reason := detectVolumeWipe(tokens); blocked {
				return cchooks.Block(reason)
			}

			// 5. Broad ownership / permission escalation
			if blocked, reason := detectRecursiveOwnershipOrPerm(tokens); blocked {
				return cchooks.Block(reason)
			}

			// 6. Suspicious exfil patterns (scp/rsync/curl to remote) – block only if wildcard root usage
			if blocked, reason := detectPotentialExfil(tokens, cmdLower); blocked {
				return cchooks.Block(reason)
			}

			return cchooks.Approve()
		},
	}

	runner.Run()
	return nil
}

// detectDangerousRm blocks destructive rm invocations aimed at root / system paths
func detectDangerousRm(tokens []string) (bool, string) {
	if len(tokens) < 2 {
		return false, ""
	}
	if tokens[0] != "rm" {
		return false, ""
	}

	flags := []string{}
	targets := []string{}
	for _, t := range tokens[1:] {
		if strings.HasPrefix(t, "-") {
			flags = append(flags, t)
		} else {
			targets = append(targets, t)
		}
	}
	flagStr := strings.Join(flags, " ")
	if !(strings.Contains(flagStr, "r") || strings.Contains(flagStr, "R")) {
		return false, ""
	}

	// Dangerous root/system targets
	dangerousPrefixes := []string{"/system", "/library", "/applications", "/users", "/private", "/usr", "/bin", "/sbin", "/etc", "/var", "/Volumes"}
	for _, tgt := range targets {
		lt := strings.ToLower(tgt)
		if lt == "/" || lt == "/*" {
			return true, "blocked rm: targets filesystem root"
		}
		for _, p := range dangerousPrefixes {
			if strings.HasPrefix(lt, p) {
				return true, "blocked rm: targets critical path " + tgt
			}
		}
		// Wildcards at root level
		if strings.HasPrefix(lt, "/*") {
			return true, "blocked rm: wildcard at root " + tgt
		}
	}
	return false, ""
}

// detectVolumeWipe identifies diskutil / asr style full volume operations missed by simple substrings
func detectVolumeWipe(tokens []string) (bool, string) {
	if len(tokens) == 0 {
		return false, ""
	}
	if tokens[0] == "diskutil" {
		for _, t := range tokens[1:] {
			if strings.HasPrefix(t, "erase") || strings.Contains(t, "apfs") {
				return true, "blocked diskutil erase / apfs operation"
			}
		}
	}
	if tokens[0] == "asr" {
		for _, t := range tokens[1:] {
			if t == "restore" {
				return true, "blocked asr restore (imaging) operation"
			}
		}
	}
	return false, ""
}

// detectRecursiveOwnershipOrPerm blocks broad recursive chmod/chown at root/system
func detectRecursiveOwnershipOrPerm(tokens []string) (bool, string) {
	if len(tokens) < 2 {
		return false, ""
	}
	if tokens[0] != "chmod" && tokens[0] != "chown" {
		return false, ""
	}
	hasRecursive := false
	for _, t := range tokens[1:] {
		if strings.HasPrefix(t, "-") && strings.Contains(t, "R") {
			hasRecursive = true
			break
		}
	}
	if !hasRecursive {
		return false, ""
	}
	// Last token(s) likely paths; scan all non-flag tokens after flags removed
	for _, t := range tokens[1:] {
		if strings.HasPrefix(t, "-") {
			continue
		}
		lt := strings.ToLower(t)
		if lt == "/" || strings.HasPrefix(lt, "/system") || strings.HasPrefix(lt, "/library") {
			return true, "blocked recursive " + tokens[0] + " on critical path " + t
		}
	}
	return false, ""
}

// detectPotentialExfil identifies scp/rsync/curl with suspicious source path breadth
func detectPotentialExfil(tokens []string, cmdLower string) (bool, string) {
	if len(tokens) == 0 {
		return false, ""
	}
	switch tokens[0] {
	case "scp", "rsync":
		// broad patterns originating at root
		if strings.Contains(cmdLower, " / ") || strings.Contains(cmdLower, " /etc") || strings.Contains(cmdLower, " /var") {
			return true, "blocked potential mass file transfer (" + tokens[0] + ") from system paths"
		}
	case "curl":
		// simplistic heuristic: uploading from system path via -T
		if strings.Contains(cmdLower, "-t /etc") || strings.Contains(cmdLower, "-t /var") {
			return true, "blocked curl upload of system files"
		}
	}
	return false, ""
}

// runFormatHook implements code formatting logic
func runFormatHook() error {
	if !isPluginEnabled("format") {
		fmt.Println("Format plugin disabled - skipping")
		return nil
	}

	runner := &cchooks.Runner{
		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			// Format code files after editing
			if event.ToolName == "Edit" || event.ToolName == "Write" {
				var filePath string

				if event.ToolName == "Edit" {
					edit, err := event.InputAsEdit()
					if err == nil {
						filePath = edit.FilePath
					}
				} else if event.ToolName == "Write" {
					write, err := event.InputAsWrite()
					if err == nil {
						filePath = write.FilePath
					}
				}

				if filePath != "" {
					ext := strings.ToLower(filepath.Ext(filePath))

					switch ext {
					case ".go":
						// Execute gofmt command
						cmd := exec.Command("gofmt", "-w", filePath)
						output, err := cmd.CombinedOutput()
						if err != nil {
							log.Printf("gofmt error on %s: %s", filePath, output)
						} else {
							fmt.Printf("Formatted Go file: %s\n", filePath)
						}
					case ".js", ".ts", ".jsx", ".tsx":
						// Execute prettier command
						cmd := exec.Command("prettier", "--write", filePath)
						output, err := cmd.CombinedOutput()
						if err != nil {
							log.Printf("prettier error on %s: %s", filePath, output)
						} else {
							fmt.Printf("Formatted JS/TS file: %s\n", filePath)
						}
					case ".py":
						// Execute ruff format via uvx (isolated environment)
						cmd := exec.Command("uvx", "ruff", "format", filePath)
						output, err := cmd.CombinedOutput()
						if err != nil {
							log.Printf("ruff format error on %s: %s", filePath, output)
						} else {
							fmt.Printf("Formatted Python file: %s\n", filePath)
						}
					}
				}
			}

			return cchooks.Allow()
		},
	}

	runner.Run()
	return nil
}

// runDebugHook implements debug logging logic
func runDebugHook() error {
	if !isPluginEnabled("debug") {
		fmt.Println("Debug plugin disabled - skipping")
		return nil
	}

	// Setup logging
	logFile, err := os.OpenFile("claude-hooks.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			logger.Printf("PRE-TOOL: %s", event.ToolName)

			// Log specific tool details
			switch event.ToolName {
			case "Bash":
				if bash, err := event.AsBash(); err == nil {
					logger.Printf("  Command: %s", bash.Command)
				}
			case "Edit":
				if edit, err := event.AsEdit(); err == nil {
					logger.Printf("  File: %s", edit.FilePath)
				}
			case "Write":
				if write, err := event.AsWrite(); err == nil {
					logger.Printf("  File: %s", write.FilePath)
				}
			case "Read":
				if read, err := event.AsRead(); err == nil {
					logger.Printf("  File: %s", read.FilePath)
				}
			}

			return cchooks.Approve()
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			logger.Printf("POST-TOOL: %s", event.ToolName)
			return cchooks.Allow()
		},
	}

	fmt.Println("Debug hook started - logging to claude-hooks.log")
	runner.Run()
	return nil
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp string                 `json:"timestamp"`
	Event     string                 `json:"event"`
	ToolName  string                 `json:"tool_name"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

func logAuditEntry(entry AuditEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit entry: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

// runAuditHook implements comprehensive audit logging
func runAuditHook() error {
	if !isPluginEnabled("audit") {
		fmt.Println("Audit plugin disabled - skipping")
		return nil
	}
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			entry := AuditEntry{
				Event:    "pre_tool_use",
				ToolName: event.ToolName,
				Details:  make(map[string]interface{}),
			}

			// Add tool-specific details
			switch event.ToolName {
			case "Bash":
				if bash, err := event.AsBash(); err == nil {
					entry.Details["command"] = bash.Command
					entry.Details["description"] = bash.Description
				}
			case "Edit":
				if edit, err := event.AsEdit(); err == nil {
					entry.Details["file_path"] = edit.FilePath
					entry.Details["old_string_length"] = len(edit.OldString)
					entry.Details["new_string_length"] = len(edit.NewString)
				}
			case "Write":
				if write, err := event.AsWrite(); err == nil {
					entry.Details["file_path"] = write.FilePath
					entry.Details["content_length"] = len(write.Content)
				}
			case "Read":
				if read, err := event.AsRead(); err == nil {
					entry.Details["file_path"] = read.FilePath
				}
			}

			logAuditEntry(entry)
			return cchooks.Approve()
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			entry := AuditEntry{
				Event:    "post_tool_use",
				ToolName: event.ToolName,
				Details:  make(map[string]interface{}),
			}

			logAuditEntry(entry)
			return cchooks.Allow()
		},
	}

	runner.Run()
	return nil
}
