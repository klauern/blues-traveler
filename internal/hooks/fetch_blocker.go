package hooks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brads3290/cchooks"
	"github.com/klauern/blues-traveler/internal/config"
	"github.com/klauern/blues-traveler/internal/core"
)

// FetchBlockerHook implements URL path prefix blocking logic for WebFetch calls
type FetchBlockerHook struct {
	*core.BaseHook
}

// NewFetchBlockerHook creates a new fetch blocker hook instance
func NewFetchBlockerHook(ctx *core.HookContext) core.Hook {
	base := core.NewBaseHook("fetch-blocker", "Fetch URL Blocker", "Blocks WebFetch calls to URLs that require authentication or alternative access methods", ctx)
	return &FetchBlockerHook{BaseHook: base}
}

// Run executes the fetch blocker hook.
func (h *FetchBlockerHook) Run() error {
	if !h.IsEnabled() {
		fmt.Println("Fetch blocker plugin disabled - skipping")
		return nil
	}

	runner := h.Context().RunnerFactory(h.preToolUseHandler, nil, h.CreateRawHandler())
	runner.Run()
	return nil
}

func (h *FetchBlockerHook) preToolUseHandler(_ context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	h.logEventDetails(event)

	// Only check WebFetch calls
	if event.ToolName != "WebFetch" {
		return cchooks.Approve()
	}

	webFetch, err := event.AsWebFetch()
	if err != nil {
		h.logError(event.ToolName, err)
		return cchooks.Block("failed to parse WebFetch command")
	}

	// Load blocked prefixes
	blockedPrefixes, err := h.loadAllBlockedPrefixes()
	if err != nil {
		h.logLoadError(event.ToolName, err)
		// If we can't load the file, allow the request (fail open)
		return cchooks.Approve()
	}

	// Check and handle blocked URLs
	return h.checkAndBlockURL(webFetch.URL, blockedPrefixes)
}

// logEventDetails logs the event details if logging is enabled
func (h *FetchBlockerHook) logEventDetails(event *cchooks.PreToolUseEvent) {
	if !h.Context().LoggingEnabled {
		return
	}

	rawData := map[string]interface{}{"tool_name": event.ToolName}
	details := make(map[string]interface{})

	if event.ToolName == "WebFetch" {
		if webFetch, err := event.AsWebFetch(); err == nil {
			details["url"] = webFetch.URL
			details["prompt"] = webFetch.Prompt
		}
	}

	h.LogHookEvent("pre_tool_use_fetch_check", event.ToolName, rawData, details)
}

// logError logs a parsing error
func (h *FetchBlockerHook) logError(toolName string, err error) {
	if h.Context().LoggingEnabled {
		h.LogHookEvent("fetch_blocker_error", toolName, map[string]interface{}{"error": err.Error()}, nil)
	}
}

// logLoadError logs an error loading blocked prefixes
func (h *FetchBlockerHook) logLoadError(toolName string, err error) {
	if h.Context().LoggingEnabled {
		h.LogHookEvent("fetch_blocker_error", toolName, map[string]interface{}{
			"error": fmt.Sprintf("failed to load blocked prefixes: %v", err),
		}, nil)
	}
}

// loadAllBlockedPrefixes loads blocked prefixes from config and files
func (h *FetchBlockerHook) loadAllBlockedPrefixes() ([]BlockedPrefix, error) {
	blockedPrefixes := h.loadBlockedFromConfig()
	if len(blockedPrefixes) == 0 {
		// Fallback to files if not configured in JSON
		return h.loadBlockedPrefixes()
	}
	return blockedPrefixes, nil
}

// checkAndBlockURL checks if a URL should be blocked and returns appropriate response
func (h *FetchBlockerHook) checkAndBlockURL(url string, blockedPrefixes []BlockedPrefix) cchooks.PreToolUseResponseInterface {
	blocked, matchedPrefix, suggestion := h.isURLBlocked(url, blockedPrefixes)
	if !blocked {
		// Log approval and return
		if h.Context().LoggingEnabled {
			h.LogHookEvent("fetch_blocker_approved", "WebFetch", map[string]interface{}{"url": url}, nil)
		}
		return cchooks.Approve()
	}

	// Log block event
	if h.Context().LoggingEnabled {
		h.LogHookEvent("fetch_blocker_block", "WebFetch", map[string]interface{}{
			"url":            url,
			"matched_prefix": matchedPrefix,
			"suggestion":     suggestion,
		}, nil)
	}

	// Build block message
	message := fmt.Sprintf("URL blocked: matches prefix '%s'", matchedPrefix)
	if suggestion != "" {
		message += fmt.Sprintf(". %s", suggestion)
	}
	return cchooks.Block(message)
}

// loadBlockedPrefixes loads URL prefixes from the blocked URLs file
func (h *FetchBlockerHook) loadBlockedPrefixes() ([]BlockedPrefix, error) {
	// Look for blocked URLs file in multiple locations:
	// 1. Project-local: ./.claude/blocked-urls.txt
	// 2. Global: ~/.claude/blocked-urls.txt

	var filePaths []string

	// Project-local file
	if cwd, err := os.Getwd(); err == nil {
		filePaths = append(filePaths, filepath.Join(cwd, ".claude", "blocked-urls.txt"))
	}

	// Global file
	if homeDir, err := os.UserHomeDir(); err == nil {
		filePaths = append(filePaths, filepath.Join(homeDir, ".claude", "blocked-urls.txt"))
	}

	var prefixes []BlockedPrefix

	// Try each file location
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		file, err := os.Open(filePath) // #nosec G304 - controlled config file paths
		if err != nil {
			continue
		}
		defer func() {
			_ = file.Close() // Ignore close error in defer
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse line format: "prefix|suggestion" or just "prefix"
			parts := strings.SplitN(line, "|", 2)
			blocked := BlockedPrefix{
				Prefix: parts[0],
			}
			if len(parts) > 1 {
				blocked.Suggestion = parts[1]
			}
			prefixes = append(prefixes, blocked)
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
		}

		// If we successfully loaded from this file, return the prefixes
		return prefixes, nil
	}

	// No file found, return empty list (allow all)
	return []BlockedPrefix{}, nil
}

func (h *FetchBlockerHook) loadBlockedFromConfig() []BlockedPrefix {
	// Project then global
	for _, global := range []bool{false, true} {
		cfgPath, err := config.GetLogConfigPath(global)
		if err != nil {
			continue
		}
		lc, err := config.LoadLogConfig(cfgPath)
		if err != nil || lc == nil {
			continue
		}
		if len(lc.BlockedURLs) == 0 {
			continue
		}
		out := make([]BlockedPrefix, 0, len(lc.BlockedURLs))
		for _, b := range lc.BlockedURLs {
			out = append(out, BlockedPrefix{Prefix: b.Prefix, Suggestion: b.Suggestion})
		}
		return out
	}
	return []BlockedPrefix{}
}

// BlockedPrefix represents a blocked URL prefix with optional suggestion
type BlockedPrefix struct {
	Prefix     string
	Suggestion string
}

// isURLBlocked checks if a URL should be blocked based on prefix/pattern matching
func (h *FetchBlockerHook) isURLBlocked(url string, blockedPrefixes []BlockedPrefix) (bool, string, string) {
	for _, blocked := range blockedPrefixes {
		pat := blocked.Prefix

		// Fast-path: no wildcard → prefix match
		if !strings.Contains(pat, "*") {
			if strings.HasPrefix(url, pat) {
				return true, pat, blocked.Suggestion
			}
			continue
		}

		// Handle wildcard patterns using simple glob matching
		if wildcardMatch(url, pat) {
			return true, pat, blocked.Suggestion
		}
	}

	return false, "", ""
}

// wildcardMatch matches s against a pattern where '*' means any sequence (including '/').
func wildcardMatch(s, pattern string) bool {
	// Escape regex meta, then restore ".*" for '*'
	rePat := regexp.QuoteMeta(pattern)
	rePat = strings.ReplaceAll(rePat, `\*`, `.*`)
	// Anchor to beginning (prefix match behavior)
	rePat = "^" + rePat
	rx, err := regexp.Compile(rePat)
	if err != nil {
		return false // Invalid pattern
	}
	return rx.MatchString(s)
}
