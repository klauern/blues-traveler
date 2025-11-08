package cursor

import (
	"testing"

	"github.com/klauern/blues-traveler/internal/core"
)

func TestEventAliasesIncludesBeforeReadFile(t *testing.T) {
	t.Parallel()

	aliases := EventAliases(core.PreToolUseEvent)
	if len(aliases) == 0 {
		t.Fatalf("expected aliases for PreToolUseEvent, got none")
	}

	found := false
	for _, alias := range aliases {
		if alias == "BeforeReadFile" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected BeforeReadFile alias for PreToolUseEvent, got %v", aliases)
	}
}

func TestResolveCursorEventRecognizesBeforeReadFile(t *testing.T) {
	t.Parallel()

	resolved, ok := ResolveCursorEvent("BeforeReadFile")
	if !ok {
		t.Fatalf("expected BeforeReadFile to resolve to a core event")
	}

	if resolved != core.PreToolUseEvent {
		t.Fatalf("expected BeforeReadFile to resolve to PreToolUseEvent, got %q", resolved)
	}
}
