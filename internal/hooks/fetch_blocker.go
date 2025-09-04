package hooks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
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

func (h *FetchBlockerHook) preToolUseHandler(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
	// Log detailed event data if logging is enabled
	if h.Context().LoggingEnabled {
		details := make(map[string]interface{})
		rawData := make(map[string]interface{})
		rawData["tool_name"] = event.ToolName

		if event.ToolName == "WebFetch" {
			if webFetch, err := event.AsWebFetch(); err == nil {
				details["url"] = webFetch.URL
				details["prompt"] = webFetch.Prompt
			}
		}

		h.LogHookEvent("pre_tool_use_fetch_check", event.ToolName, rawData, details)
	}

	// Only check WebFetch calls
	if event.ToolName != "WebFetch" {
		return cchooks.Approve()
	}

	webFetch, err := event.AsWebFetch()
	if err != nil {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("fetch_blocker_error", event.ToolName, map[string]interface{}{"error": err.Error()}, nil)
		}
		return cchooks.Block("failed to parse WebFetch command")
	}

	// Load blocked URL prefixes from embedded config first, then files
	blockedPrefixes, err := h.loadBlockedFromConfig()
	if err == nil && len(blockedPrefixes) == 0 {
		// Fallback to files if not configured in JSON
		blockedPrefixes, err = h.loadBlockedPrefixes()
	}
	if err != nil {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("fetch_blocker_error", event.ToolName, map[string]interface{}{
				"error": fmt.Sprintf("failed to load blocked prefixes: %v", err),
			}, nil)
		}
		// If we can't load the file, allow the request (fail open)
		return cchooks.Approve()
	}

	// Check if URL matches any blocked prefix
	if blocked, matchedPrefix, suggestion := h.isURLBlocked(webFetch.URL, blockedPrefixes); blocked {
		if h.Context().LoggingEnabled {
			h.LogHookEvent("fetch_blocker_block", "WebFetch", map[string]interface{}{
				"url":            webFetch.URL,
				"matched_prefix": matchedPrefix,
				"suggestion":     suggestion,
			}, nil)
		}

		message := fmt.Sprintf("URL blocked: matches prefix '%s'", matchedPrefix)
		if suggestion != "" {
			message += fmt.Sprintf(". %s", suggestion)
		}
		return cchooks.Block(message)
	}

	// Log approved fetch if logging is enabled
	if h.Context().LoggingEnabled {
		h.LogHookEvent("fetch_blocker_approved", "WebFetch", map[string]interface{}{
			"url": webFetch.URL,
		}, nil)
	}

	return cchooks.Approve()
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
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		// If we successfully loaded from this file, return the prefixes
		return prefixes, nil
	}

	// No file found, return empty list (allow all)
	return []BlockedPrefix{}, nil
}

func (h *FetchBlockerHook) loadBlockedFromConfig() ([]BlockedPrefix, error) {
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
		return out, nil
	}
	return []BlockedPrefix{}, nil
}

// BlockedPrefix represents a blocked URL prefix with optional suggestion
type BlockedPrefix struct {
	Prefix     string
	Suggestion string
}

// isURLBlocked checks if a URL should be blocked based on prefix matching
func (h *FetchBlockerHook) isURLBlocked(url string, blockedPrefixes []BlockedPrefix) (bool, string, string) {
	for _, blocked := range blockedPrefixes {
		prefix := blocked.Prefix

		// Handle wildcard patterns (e.g., "https://example.com/path*")
		if strings.HasSuffix(prefix, "*") {
			prefixWithoutWildcard := prefix[:len(prefix)-1]
			if strings.HasPrefix(url, prefixWithoutWildcard) {
				return true, prefix, blocked.Suggestion
			}
		} else {
			// Exact prefix match
			if strings.HasPrefix(url, prefix) {
				return true, prefix, blocked.Suggestion
			}
		}
	}

	return false, "", ""
}
