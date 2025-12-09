package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/klauern/blues-traveler/internal/config"
)

// Scope constants
const (
	ScopeProject = "project"
	ScopeGlobal  = "global"
)

// listAvailableHooks lists all available hook plugins
func listAvailableHooks(
	getPlugin func(string) (PluginProvider, bool),
	pluginKeys func() []string,
) error {
	keys := pluginKeys()
	var builtin, custom []string
	groupSet := map[string]struct{}{}
	for _, k := range keys {
		if strings.HasPrefix(k, "config:") {
			custom = append(custom, k)
			parts := strings.SplitN(k, ":", 3)
			if len(parts) == 3 {
				groupSet[parts[1]] = struct{}{}
			}
		} else {
			builtin = append(builtin, k)
		}
	}
	sort.Strings(builtin)
	sort.Strings(custom)

	fmt.Println("Available hook plugins:")
	fmt.Println()
	if len(builtin) > 0 {
		fmt.Println("Built-in hooks:")
		for _, key := range builtin {
			p, _ := getPlugin(key)
			fmt.Printf("  %s - %s\n", key, p.Description())
		}
		fmt.Println()
	}
	if len(custom) > 0 {
		fmt.Println("Custom config hooks (from config):")
		for _, key := range custom {
			p, _ := getPlugin(key)
			fmt.Printf("  %s - %s\n", key, p.Description())
		}
		// Show groups summary
		groups := make([]string, 0, len(groupSet))
		for g := range groupSet {
			groups = append(groups, g)
		}
		sort.Strings(groups)
		if len(groups) > 0 {
			fmt.Printf("\nGroups: %s\n", strings.Join(groups, ", "))
		}
		fmt.Println()
	} else {
		// Suggest how to create custom hooks if none present
		fmt.Println("No custom hooks found. Create .claude/hooks.yml and use 'blues-traveler hooks custom install --list'.")
		fmt.Println()
	}

	fmt.Println("Use 'blues-traveler hooks run <key>' to run a hook.")
	fmt.Println("Use 'blues-traveler hooks install <key>' to install a built-in hook.")
	fmt.Println("Use 'blues-traveler hooks custom install <group>' to install a group from hooks.yml.")
	return nil
}

// listInstalledHooks lists hooks installed in settings
func listInstalledHooks(global bool) error {
	// Get settings path
	settingsPath, err := config.GetSettingsPath(global)
	if err != nil {
		scope := ScopeProject
		if global {
			scope = ScopeGlobal
		}
		return fmt.Errorf("failed to locate %s settings path: %w\n  Suggestion: Ensure you're in a project directory or use --global flag for global settings", scope, err)
	}

	// Load existing settings
	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load settings from %s: %w\n  Suggestion: Check if the settings file exists and is valid JSON", settingsPath, err)
	}

	scope := ScopeProject
	if global {
		scope = ScopeGlobal
	}

	fmt.Printf("Installed hooks (%s settings):\n", scope)
	fmt.Printf("Settings file: %s\n\n", settingsPath)

	if config.IsHooksConfigEmpty(settings.Hooks) {
		fmt.Println("No hooks are currently installed.")
	} else {
		printHookMatchers("PreToolUse", settings.Hooks.PreToolUse)
		printHookMatchers("PostToolUse", settings.Hooks.PostToolUse)
		printHookMatchers("UserPromptSubmit", settings.Hooks.UserPromptSubmit)
		printHookMatchers("Notification", settings.Hooks.Notification)
		printHookMatchers("Stop", settings.Hooks.Stop)
		printHookMatchers("SubagentStop", settings.Hooks.SubagentStop)
		printHookMatchers("PreCompact", settings.Hooks.PreCompact)
		printHookMatchers("SessionStart", settings.Hooks.SessionStart)
		printHookMatchers("SessionEnd", settings.Hooks.SessionEnd)
	}

	// Add examples section
	printUninstallExamples(global)
	return nil
}

// listEvents lists all available Claude Code hook events
func listEvents(allEvents func() []ClaudeCodeEvent) error {
	fmt.Println("Available Claude Code Hook Events:")
	fmt.Println()

	events := allEvents()
	ccHooksSupported := 0
	for _, event := range events {
		status := ""
		if event.SupportedByCCHooks {
			status = " ✓ (cchooks library)"
			ccHooksSupported++
		} else {
			status = " ⚠ (Claude Code only)"
		}

		fmt.Printf("  %s%s\n", event.Name, status)
		fmt.Printf("      %s\n", event.Description)
		fmt.Println()
	}

	fmt.Printf("Total: %d events available (%d supported by cchooks library)\n\n", len(events), ccHooksSupported)
	fmt.Println("✓ Events marked with checkmark can be handled by blues-traveler plugins")
	fmt.Println("⚠ Events marked with warning require custom hook implementations")
	fmt.Println()
	fmt.Println("Use 'blues-traveler hooks install <plugin-key> --event <event-name>' to install a hook for a specific event.")
	fmt.Println("Use 'blues-traveler hooks list --installed' to see currently configured hooks.")
	return nil
}

// printHookMatchers prints hook matchers for a specific event
func printHookMatchers(eventName string, matchers []config.HookMatcher) {
	if len(matchers) == 0 {
		return
	}

	fmt.Printf("%s:\n", eventName)
	for _, matcher := range matchers {
		matcherStr := matcher.Matcher
		if matcherStr == "" {
			matcherStr = "*"
		}
		fmt.Printf("  Matcher: %s\n", matcherStr)
		for _, hook := range matcher.Hooks {
			fmt.Printf("    - %s", hook.Command)
			if hook.Timeout != nil {
				fmt.Printf(" (timeout: %ds)", *hook.Timeout)
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

// printUninstallExamples prints examples of how to uninstall hooks
func printUninstallExamples(global bool) {
	scope := ScopeProject
	globalFlag := ""
	if global {
		scope = ScopeGlobal
		globalFlag = " --global"
	}

	fmt.Printf("📝 How to remove hooks:\n\n")

	fmt.Printf("Remove a specific hook type (removes ALL instances):\n")
	fmt.Printf("  blues-traveler hooks uninstall debug%s\n", globalFlag)
	fmt.Printf("  blues-traveler hooks uninstall security%s\n", globalFlag)
	fmt.Printf("  blues-traveler hooks uninstall audit%s\n\n", globalFlag)

	fmt.Printf("Remove ALL blues-traveler hooks (preserves other hooks):\n")
	fmt.Printf("  blues-traveler hooks uninstall all%s\n\n", globalFlag)

	fmt.Printf("Remove ALL hooks from %s settings:\n", scope)
	if global {
		fmt.Printf("  rm ~/.claude/settings.json\n")
		fmt.Printf("  # (or edit the file manually to remove specific events)\n\n")
	} else {
		fmt.Printf("  rm .claude/settings.json\n")
		fmt.Printf("  # (or edit the file manually to remove specific events)\n\n")
	}

	fmt.Printf("View this list again:\n")
	fmt.Printf("  %s hooks list --installed%s\n\n", "blues-traveler", globalFlag)

	fmt.Printf("💡 Note: The 'uninstall' command removes ALL instances of a hook type\n")
	fmt.Printf("   from ALL events (PreToolUse, PostToolUse, etc.)\n")
}

// syncOptions holds parameters for the sync command
type syncOptions struct {
	useGlobal       bool
	dryRun          bool
	eventFilter     string
	groupFilter     string
	defaultMatcher  string
	postMatcher     string
	timeoutOverride int
	execPath        string
}

// pickMatcherForEvent returns the appropriate matcher based on event type
func pickMatcherForEvent(eventName, postMatcher, defaultMatcher string) string {
	if eventName == "PostToolUse" {
		return postMatcher
	}
	return defaultMatcher
}

// buildConfigGroupsMap creates a map of group names from hooks config
func buildConfigGroupsMap(hooksCfg *config.CustomHooksConfig) map[string]bool {
	configGroups := make(map[string]bool)
	if hooksCfg != nil {
		for groupName := range *hooksCfg {
			configGroups[groupName] = true
		}
	}
	return configGroups
}

// cleanupStaleGroups removes groups that exist in settings but not in config
func cleanupStaleGroups(settings *config.Settings, existingGroups map[string]bool, configGroups map[string]bool, opts syncOptions) int {
	changed := 0
	for existingGroup := range existingGroups {
		if shouldSkipGroup(existingGroup, opts.groupFilter) {
			continue
		}
		if !configGroups[existingGroup] {
			removed := config.RemoveConfigGroupFromSettings(settings, existingGroup, opts.eventFilter)
			if removed > 0 {
				printCleanupMessage(removed, existingGroup, opts.eventFilter)
				changed += removed
			}
		}
	}
	return changed
}

// shouldSkipGroup returns true if the group should be skipped based on filter
func shouldSkipGroup(groupName, groupFilter string) bool {
	return groupFilter != "" && groupName != groupFilter
}

// printCleanupMessage prints a message about cleaned up entries
func printCleanupMessage(removed int, groupName, eventFilter string) {
	suffix := ""
	if eventFilter != "" {
		suffix = " (event: " + eventFilter + ")"
	}
	fmt.Printf("Cleaned up %d stale entries for removed group '%s'%s\n", removed, groupName, suffix)
}

// printPrunedMessage prints a message about pruned entries
func printPrunedMessage(removed int, groupName, eventFilter string) {
	suffix := ""
	if eventFilter != "" {
		suffix = " (event: " + eventFilter + ")"
	}
	fmt.Printf("Pruned %d entries for group '%s'%s\n", removed, groupName, suffix)
}

// syncGroupToSettings syncs a single group's events and jobs to settings
func syncGroupToSettings(settings *config.Settings, groupName string, group config.HookGroup, opts syncOptions) int {
	changed := 0
	for eventName, ev := range group {
		if shouldSkipEvent(eventName, opts.eventFilter) {
			continue
		}
		changed += syncJobsForEvent(settings, groupName, eventName, ev, opts)
	}
	return changed
}

// shouldSkipEvent returns true if the event should be skipped based on filter
func shouldSkipEvent(eventName, eventFilter string) bool {
	return eventFilter != "" && eventFilter != eventName
}

// syncJobsForEvent syncs jobs for a specific event to settings
func syncJobsForEvent(settings *config.Settings, groupName, eventName string, ev *config.EventConfig, opts syncOptions) int {
	changed := 0
	for _, job := range ev.Jobs {
		if job.Name == "" {
			continue
		}
		hookCommand := buildHookCommand(opts.execPath, groupName, job.Name)
		timeout := selectTimeout(opts.timeoutOverride, job.Timeout)
		matcher := pickMatcherForEvent(eventName, opts.postMatcher, opts.defaultMatcher)

		result := config.AddHookToSettings(settings, eventName, matcher, hookCommand, timeout)
		if !result.WasDuplicate {
			changed++
		}

		if opts.dryRun {
			fmt.Printf("Would add: [%s] matcher=%q command=%q\n", eventName, matcher, hookCommand)
		}
	}
	return changed
}

// buildHookCommand constructs the hook command string
func buildHookCommand(execPath, groupName, jobName string) string {
	return fmt.Sprintf("%s hooks run config:%s:%s", execPath, groupName, jobName)
}

// selectTimeout returns the appropriate timeout value
func selectTimeout(override, jobTimeout int) *int {
	if override > 0 {
		return &override
	}
	if jobTimeout > 0 {
		return &jobTimeout
	}
	return nil
}

// installOptions holds parameters for the install command
type installOptions struct {
	groupName       string
	useGlobal       bool
	defaultMatcher  string
	postMatcher     string
	eventFilter     string
	timeoutOverride int
	prune           bool
	init            bool
}

// listCustomHookGroups lists all custom hook groups from config
func listCustomHookGroups(cfg *config.CustomHooksConfig) error {
	groups := config.ListHookGroups(cfg)
	if len(groups) == 0 {
		fmt.Println("No custom hook groups found. Create .claude/hooks.yml to define groups.")
		return nil
	}
	fmt.Println("Available custom hook groups:")
	for _, g := range groups {
		fmt.Printf("- %s\n", g)
	}
	return nil
}

// loadOrCreateGroup loads a group from config, optionally creating a stub if --init is used
func loadOrCreateGroup(cfg *config.CustomHooksConfig, groupName string, initFlag, useGlobal bool) (*config.CustomHooksConfig, error) {
	if cfg != nil && (*cfg)[groupName] != nil {
		return cfg, nil
	}

	if !initFlag {
		return nil, fmt.Errorf("group '%s' not found in hooks config (use --init to stub one)", groupName)
	}

	// Create stub group
	sample := createSampleGroupYAML(groupName)
	if _, err := config.WriteSampleHooksConfig(useGlobal, sample, false); err != nil {
		return nil, fmt.Errorf("write hooks sample: %w", err)
	}

	// Reload after creating stub
	reloadedCfg, err := config.LoadHooksConfig()
	if err != nil {
		return nil, fmt.Errorf("reload hooks config: %w", err)
	}

	if reloadedCfg == nil || (*reloadedCfg)[groupName] == nil {
		return nil, fmt.Errorf("failed to create group '%s' in hooks.yml", groupName)
	}

	return reloadedCfg, nil
}

// createSampleGroupYAML creates a sample YAML stub for a new group
func createSampleGroupYAML(groupName string) string {
	return fmt.Sprintf(`%s:
  PreToolUse:
    jobs:
      - name: example-check
        run: echo "TOOL=${TOOL_NAME} FILES=${FILES_CHANGED}"
        glob: ["*"]
  PostToolUse:
    jobs:
      - name: example-post
        run: echo "Post ${EVENT_NAME} on ${TOOL_NAME}"
        glob: ["*"]
`, groupName)
}

// installGroupHooks installs all hooks from a group into settings
func installGroupHooks(settings *config.Settings, group config.HookGroup, opts installOptions) int {
	installed := 0
	for eventName, ev := range group {
		if shouldSkipEvent(eventName, opts.eventFilter) {
			continue
		}
		installed += installJobsForEvent(settings, eventName, ev, opts)
	}
	return installed
}

// installJobsForEvent installs jobs for a specific event
func installJobsForEvent(settings *config.Settings, eventName string, ev *config.EventConfig, opts installOptions) int {
	installed := 0
	for _, job := range ev.Jobs {
		if job.Name == "" {
			continue
		}

		hookCommand, err := buildInstallCommand(opts.groupName, job.Name)
		if err != nil {
			continue // skip on error
		}

		timeout := selectTimeout(opts.timeoutOverride, job.Timeout)
		matcher := pickMatcherForEvent(eventName, opts.postMatcher, opts.defaultMatcher)

		config.AddHookToSettings(settings, eventName, matcher, hookCommand, timeout)
		installed++
	}
	return installed
}

// buildInstallCommand builds the hook command for installation
func buildInstallCommand(groupName, jobName string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Quote the path when it contains spaces to handle paths like "/Program Files/app"
	// Commands are executed via bash -lc, so unquoted paths with spaces break
	if strings.ContainsRune(execPath, ' ') {
		execPath = `"` + execPath + `"`
	}

	return fmt.Sprintf("%s hooks run config:%s:%s", execPath, groupName, jobName), nil
}

// printInstallSuccess prints success message for hook installation
func printInstallSuccess(groupName, scope string, installed int, settingsPath string) {
	fmt.Printf("✅ Installed custom group '%s' to %s settings (%d entries)\n", groupName, scope, installed)
	fmt.Printf("   Settings: %s\n", settingsPath)
}

// getScopeName returns the scope name for display
func getScopeName(useGlobal bool) string {
	if useGlobal {
		return ScopeGlobal
	}
	return ScopeProject
}

// blockedURLHelpers contains helper functions for blocked URL management

// loadLogConfigForBlockedURLs loads log config for blocked URL operations
func loadLogConfigForBlockedURLs(useGlobal bool) (string, *config.LogConfig, error) {
	path, err := config.GetLogConfigPath(useGlobal)
	if err != nil {
		return "", nil, err
	}

	lc, err := config.LoadLogConfig(path)
	if err != nil {
		return "", nil, err
	}

	return path, lc, nil
}

// displayBlockedURLs prints the blocked URLs list
func displayBlockedURLs(lc *config.LogConfig, path string, useGlobal bool) {
	scope := getScopeName(useGlobal)
	fmt.Printf("Blocked URLs (%s config: %s):\n", scope, path)

	if len(lc.BlockedURLs) == 0 {
		fmt.Println("(none)")
		return
	}

	for _, b := range lc.BlockedURLs {
		if b.Suggestion != "" {
			fmt.Printf("- %s | %s\n", b.Prefix, b.Suggestion)
		} else {
			fmt.Printf("- %s\n", b.Prefix)
		}
	}
}

// addBlockedURL adds a new blocked URL prefix
func addBlockedURL(lc *config.LogConfig, prefix, suggestion string) bool {
	// Check duplicate
	for _, b := range lc.BlockedURLs {
		if b.Prefix == prefix {
			return false
		}
	}

	lc.BlockedURLs = append(lc.BlockedURLs, config.BlockedURL{
		Prefix:     prefix,
		Suggestion: suggestion,
	})
	return true
}

// removeBlockedURL removes a blocked URL prefix
func removeBlockedURL(lc *config.LogConfig, prefix string) bool {
	filtered := lc.BlockedURLs[:0]
	removed := false

	for _, b := range lc.BlockedURLs {
		if b.Prefix == prefix {
			removed = true
			continue
		}
		filtered = append(filtered, b)
	}

	if removed {
		lc.BlockedURLs = filtered
	}
	return removed
}
