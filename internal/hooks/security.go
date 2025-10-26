package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/constants"
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

// securityCheck represents a single security check
type securityCheck struct {
	checkType string
	check     func([]string, string) (bool, string)
}

// logSecurityEvent logs a security event with standard formatting
func (h *SecurityHook) logSecurityEvent(eventType, command, reason, checkType string) {
	if !h.Context().LoggingEnabled {
		return
	}

	h.LogHookEvent(eventType, constants.ToolBash, map[string]interface{}{
		"command":    command,
		"reason":     reason,
		"check_type": checkType,
	}, nil)
}

// logPreToolUseCheck logs the initial pre-tool-use check
func (h *SecurityHook) logPreToolUseCheck(event *cchooks.PreToolUseEvent) {
	if !h.Context().LoggingEnabled {
		return
	}

	details := make(map[string]interface{})
	rawData := map[string]interface{}{"tool_name": event.ToolName}

	if event.ToolName == constants.ToolBash {
		if bash, err := event.AsBash(); err == nil {
			details["command"] = bash.Command
			details["description"] = bash.Description
		}
	}

	h.LogHookEvent("pre_tool_use_security_check", event.ToolName, rawData, details)
}

// runSecurityChecks executes all security checks and returns the first match
func (h *SecurityHook) runSecurityChecks(_ string, tokens []string, cmdLower string) (bool, string, string) {
	checks := []securityCheck{
		{"static_patterns", func(_ []string, c string) (bool, string) { return h.checkStaticPatterns(c) }},
		{"macos_patterns", func(_ []string, c string) (bool, string) { return h.checkMacOSPatterns(c) }},
		{"dangerous_rm", func(t []string, _ string) (bool, string) { return h.detectDangerousRm(t) }},
		{"volume_wipe", func(t []string, _ string) (bool, string) { return h.detectVolumeWipe(t) }},
		{"recursive_ownership_perm", func(t []string, _ string) (bool, string) { return h.detectRecursiveOwnershipOrPerm(t) }},
		{"potential_exfil", func(t []string, c string) (bool, string) { return h.detectPotentialExfil(t, c) }},
	}

	for _, check := range checks {
		if blocked, reason := check.check(tokens, cmdLower); blocked {
			return true, reason, check.checkType
		}
	}

	return false, "", ""
}

func (h *SecurityHook) preToolUseHandler(_ context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	h.logPreToolUseCheck(event)

	if event.ToolName != constants.ToolBash {
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

	// Run all security checks
	if blocked, reason, checkType := h.runSecurityChecks(bash.Command, tokens, cmdLower); blocked {
		h.logSecurityEvent("security_block", bash.Command, reason, checkType)
		return cchooks.Block(reason)
	}

	// Log approved commands if logging is enabled
	if h.Context().LoggingEnabled {
		h.LogHookEvent("security_approved", constants.ToolBash, map[string]interface{}{
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
