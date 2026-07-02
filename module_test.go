package onlineconf

import (
	"path/filepath"
	"testing"
	"time"
)

// OpenModule should not leave a never-closed pending channel under the key
// in modCache, even if it fails.
func TestOpenModuleErrorDoesNotPoisonCache(t *testing.T) {
	name := filepath.Join(t.TempDir(), "does-not-exist", "mod")

	if _, err := OpenModule(name); err == nil {
		t.Fatalf("OpenModule(%q): expected an error for a nonexistent module", name)
	}

	type result struct {
		mod *Module
		err error
	}

	done := make(chan result, 1)

	go func() {
		mod, err := OpenModule(name)
		done <- result{mod, err}
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatalf("DEADLOCK: second OpenModule(%q) after a failed open blocked forever", name)
	case res := <-done:
		if res.err == nil {
			t.Fatalf("second OpenModule(%q): expected an error, got module %v", name, res.mod)
		}
	}
}
