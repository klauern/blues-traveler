package cursor

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("NewConfig() returned nil")
	}

	if config.Version != 1 {
		t.Errorf("NewConfig().Version = %d, want 1", config.Version)
	}

	if config.Hooks == nil {
		t.Error("NewConfig().Hooks should be initialized, got nil")
	}

	if len(config.Hooks) != 0 {
		t.Errorf("NewConfig().Hooks should be empty, got %d hooks", len(config.Hooks))
	}
}

func TestConfig_AddHook(t *testing.T) {
	t.Run("add to empty config", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/path/to/hook.sh")

		if len(config.Hooks) != 1 {
			t.Errorf("Expected 1 event, got %d", len(config.Hooks))
		}

		hooks := config.Hooks[BeforeShellExecution]
		if len(hooks) != 1 {
			t.Fatalf("Expected 1 hook, got %d", len(hooks))
		}

		if hooks[0].Command != "/path/to/hook.sh" {
			t.Errorf("Expected command '/path/to/hook.sh', got %q", hooks[0].Command)
		}
	})

	t.Run("add multiple hooks to same event", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/first/hook.sh")
		config.AddHook(BeforeShellExecution, "/second/hook.sh")

		hooks := config.Hooks[BeforeShellExecution]
		if len(hooks) != 2 {
			t.Fatalf("Expected 2 hooks, got %d", len(hooks))
		}

		if hooks[0].Command != "/first/hook.sh" {
			t.Errorf("Expected first command '/first/hook.sh', got %q", hooks[0].Command)
		}
		if hooks[1].Command != "/second/hook.sh" {
			t.Errorf("Expected second command '/second/hook.sh', got %q", hooks[1].Command)
		}
	})

	t.Run("add hooks to different events", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/shell/hook.sh")
		config.AddHook(AfterFileEdit, "/file/hook.sh")

		if len(config.Hooks) != 2 {
			t.Errorf("Expected 2 events, got %d", len(config.Hooks))
		}

		if len(config.Hooks[BeforeShellExecution]) != 1 {
			t.Errorf("Expected 1 hook for BeforeShellExecution")
		}
		if len(config.Hooks[AfterFileEdit]) != 1 {
			t.Errorf("Expected 1 hook for AfterFileEdit")
		}
	})

	t.Run("add hook to nil Hooks map", func(t *testing.T) {
		config := &Config{Version: 1}
		config.AddHook(BeforeShellExecution, "/path/to/hook.sh")

		if config.Hooks == nil {
			t.Error("Hooks map should be initialized")
		}

		if len(config.Hooks[BeforeShellExecution]) != 1 {
			t.Errorf("Expected 1 hook after adding to nil map")
		}
	})
}

func TestConfig_RemoveHook(t *testing.T) {
	t.Run("remove existing hook", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/path/to/hook.sh")

		removed := config.RemoveHook(BeforeShellExecution, "/path/to/hook.sh")
		if !removed {
			t.Error("RemoveHook() should return true when hook is removed")
		}

		if len(config.Hooks) != 0 {
			t.Errorf("Expected empty Hooks map after removing last hook, got %d events", len(config.Hooks))
		}
	})

	t.Run("remove from multiple hooks", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/first/hook.sh")
		config.AddHook(BeforeShellExecution, "/second/hook.sh")
		config.AddHook(BeforeShellExecution, "/third/hook.sh")

		removed := config.RemoveHook(BeforeShellExecution, "/second/hook.sh")
		if !removed {
			t.Error("RemoveHook() should return true")
		}

		hooks := config.Hooks[BeforeShellExecution]
		if len(hooks) != 2 {
			t.Fatalf("Expected 2 hooks remaining, got %d", len(hooks))
		}

		if hooks[0].Command != "/first/hook.sh" {
			t.Errorf("Expected first hook to remain")
		}
		if hooks[1].Command != "/third/hook.sh" {
			t.Errorf("Expected third hook to remain")
		}
	})

	t.Run("remove non-existent hook", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/path/to/hook.sh")

		removed := config.RemoveHook(BeforeShellExecution, "/different/hook.sh")
		if removed {
			t.Error("RemoveHook() should return false when hook doesn't exist")
		}

		if len(config.Hooks[BeforeShellExecution]) != 1 {
			t.Errorf("Hook should not be removed")
		}
	})

	t.Run("remove from non-existent event", func(t *testing.T) {
		config := NewConfig()

		removed := config.RemoveHook(BeforeShellExecution, "/path/to/hook.sh")
		if removed {
			t.Error("RemoveHook() should return false when event doesn't exist")
		}
	})

	t.Run("remove last hook cleans up event key", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/hook.sh")

		config.RemoveHook(BeforeShellExecution, "/hook.sh")

		if _, exists := config.Hooks[BeforeShellExecution]; exists {
			t.Error("Event key should be removed when last hook is deleted")
		}
	})
}

func TestConfig_HasHook(t *testing.T) {
	t.Run("has existing hook", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/path/to/hook.sh")

		if !config.HasHook(BeforeShellExecution, "/path/to/hook.sh") {
			t.Error("HasHook() should return true for existing hook")
		}
	})

	t.Run("does not have non-existent hook", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/path/to/hook.sh")

		if config.HasHook(BeforeShellExecution, "/different/hook.sh") {
			t.Error("HasHook() should return false for non-existent hook")
		}
	})

	t.Run("does not have hook for non-existent event", func(t *testing.T) {
		config := NewConfig()

		if config.HasHook(BeforeShellExecution, "/path/to/hook.sh") {
			t.Error("HasHook() should return false when event doesn't exist")
		}
	})

	t.Run("has hook among multiple", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/first/hook.sh")
		config.AddHook(BeforeShellExecution, "/second/hook.sh")
		config.AddHook(BeforeShellExecution, "/third/hook.sh")

		if !config.HasHook(BeforeShellExecution, "/second/hook.sh") {
			t.Error("HasHook() should find hook among multiple")
		}
	})

	t.Run("case sensitive command matching", func(t *testing.T) {
		config := NewConfig()
		config.AddHook(BeforeShellExecution, "/path/to/Hook.sh")

		if config.HasHook(BeforeShellExecution, "/path/to/hook.sh") {
			t.Error("HasHook() should be case-sensitive")
		}
	})
}

func TestConfig_AddRemoveHasHook_Integration(t *testing.T) {
	config := NewConfig()

	// Start with no hooks
	if config.HasHook(BeforeShellExecution, "/hook.sh") {
		t.Error("Should not have hook initially")
	}

	// Add hook
	config.AddHook(BeforeShellExecution, "/hook.sh")
	if !config.HasHook(BeforeShellExecution, "/hook.sh") {
		t.Error("Should have hook after adding")
	}

	// Add duplicate (should add another entry)
	config.AddHook(BeforeShellExecution, "/hook.sh")
	hooks := config.Hooks[BeforeShellExecution]
	if len(hooks) != 2 {
		t.Errorf("Should have 2 hooks after adding duplicate, got %d", len(hooks))
	}

	// Remove first occurrence
	removed := config.RemoveHook(BeforeShellExecution, "/hook.sh")
	if !removed {
		t.Error("Should have removed hook")
	}
	if len(config.Hooks[BeforeShellExecution]) != 1 {
		t.Error("Should still have one hook remaining")
	}

	// Remove second occurrence
	removed = config.RemoveHook(BeforeShellExecution, "/hook.sh")
	if !removed {
		t.Error("Should have removed second hook")
	}
	if config.HasHook(BeforeShellExecution, "/hook.sh") {
		t.Error("Should not have hook after removing all")
	}
	if _, exists := config.Hooks[BeforeShellExecution]; exists {
		t.Error("Event should be cleaned up after removing all hooks")
	}
}
