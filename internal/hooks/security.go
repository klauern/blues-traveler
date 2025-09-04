package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/core"
)

// SecurityHook implements security blocking logic for dangerous commands
type SecurityHook struct {
	*core.BaseHook
}

// NewSecurityHook creates a new security hook instance
func NewSecurityHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("security", "Security Hook", "Blocks dangerous commands and provides security controls", ctx)
	return &SecurityHook{BaseHook: base}
}

// Run executes the security hook.
func (h *SecurityHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Security plugin disabled - skipping")
		return nil
	}

	runner := h.Context().RunnerFactory(h.preToolUseHandler, nil, h.CreateRawHandler())
	runner.Run()
	return nil
}

func (h *SecurityHook) preToolUseHandler(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	// Log detailed event data if logging is enabled
	if h.Context().LoggingEnabled {
		details := make(map[string]interface{})
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName

		if event.ToolName == "Bash" {
			if bash, err := event.AsBash(); err == nil {
				details["command"] = bash.Command
				details["description"] = bash.Description
			}
		}

		h.LogHookEvent("pre_tool_use_security_check", event.ToolName, rawData, details)
	}

	if event.ToolName != "Bash" {
		return cchooks.Approve()
	}

	bash, err := event.AsBash()
	if err != nil {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_error", event.ToolName, map[string]interface{}{"error": err.Error()}, nil)
		}
		return cchooks.Block("failed to parse bash command")
	}

	cmdLower := strings.ToLower(bash.Command)
	tokens := strings.Fields(cmdLower)

	// Check various security patterns
	if blocked, reason := h.checkStaticPatterns(cmdLower); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_block", "Bash", map[string]interface{}{
				"command":    bash.Command,
				"reason":     reason,
				"check_type": "static_patterns",
			}, nil)
		}
		return cchooks.Block(reason)
	}

	if blocked, reason := h.checkMacOSPatterns(cmdLower); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_block", "Bash", map[string]interface{}{
				"command":    bash.Command,
				"reason":     reason,
				"check_type": "macos_patterns",
			}, nil)
		}
		return cchooks.Block(reason)
	}

	if blocked, reason := h.detectDangerousRm(tokens); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_block", "Bash", map[string]interface{}{
				"command":    bash.Command,
				"reason":     reason,
				"check_type": "dangerous_rm",
			}, nil)
		}
		return cchooks.Block(reason)
	}

	if blocked, reason := h.detectVolumeWipe(tokens); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_block", "Bash", map[string]interface{}{
				"command":    bash.Command,
				"reason":     reason,
				"check_type": "volume_wipe",
			}, nil)
		}
		return cchooks.Block(reason)
	}

	if blocked, reason := h.detectRecursiveOwnershipOrPerm(tokens); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_block", "Bash", map[string]interface{}{
				"command":    bash.Command,
				"reason":     reason,
				"check_type": "recursive_ownership_perm",
			}, nil)
		}
		return cchooks.Block(reason)
	}

	if blocked, reason := h.detectPotentialExfil(tokens, cmdLower); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("security_block", "Bash", map[string]interface{}{
				"command":    bash.Command,
				"reason":     reason,
				"check_type": "potential_exfil",
			}, nil)
		}
		return cchooks.Block(reason)
	}

	// Log approved commands if logging is enabled
	if h.Context().LoggingEnabled {
		h.LogHookEvent("security_approved", "Bash", map[string]interface{}{
			"command": bash.Command,
		}, nil)
	}

	return cchooks.Approve()
}

// checkStaticPatterns checks for high-risk pattern list (simple substring)
func (h *SecurityHook) checkStaticPatterns(cmdLower string) (bool, string) {
	staticSubstrings := []string{
		"dd if=",          // raw disk writing
		"mkfs",            // filesystem creation
		"> /dev/",         // redirect into device nodes
		"sudo rm",         // elevated deletion
		"shutdown -h now", // immediate shutdown
		"shutdown -r now",
		"nvram -c", // clearing NVRAM (EFI vars)
	}

	for _, s := range staticSubstrings {
		if strings.Contains(cmdLower, s) {
			return true, fmt.Sprintf("blocked dangerous command pattern: %s", s)
		}
	}
	return false, ""
}

// checkMacOSPatterns checks macOS specific critical command regexes
func (h *SecurityHook) checkMacOSPatterns(cmdLower string) (bool, string) {
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
			return true, "blocked high-risk macOS command: " + label
		}
	}
	return false, ""
}

// detectDangerousRm blocks destructive rm invocations aimed at root / system paths
func (h *SecurityHook) detectDangerousRm(tokens []string) (bool, string) {
	if len(tokens) < 2 || tokens[0] != "rm" {
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
	if !strings.Contains(flagStr, "r") && !strings.Contains(flagStr, "R") {
		return false, ""
	}

	// Dangerous root/system targets
	dangerousPrefixes := []string{"/system", "/library", "/applications", "/users", "/private", "/usr", "/bin", "/sbin", "/etc", "/var", "/Volumes"}
	for _, tgt := range targets {
		lt := strings.ToLower(tgt)
		if lt == "/" {
			return true, "blocked rm: targets filesystem root"
		}
		// Wildcards at root level (check before exact match for /*)
		if strings.HasPrefix(lt, "/*") {
			return true, "blocked rm: wildcard at root " + tgt
		}
		for _, p := range dangerousPrefixes {
			if strings.HasPrefix(lt, p) {
				return true, "blocked rm: targets critical path " + tgt
			}
		}
	}
	return false, ""
}

// detectVolumeWipe identifies diskutil / asr style full volume operations
func (h *SecurityHook) detectVolumeWipe(tokens []string) (bool, string) {
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
func (h *SecurityHook) detectRecursiveOwnershipOrPerm(tokens []string) (bool, string) {
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

	// Check paths in tokens
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
func (h *SecurityHook) detectPotentialExfil(tokens []string, cmdLower string) (bool, string) {
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
