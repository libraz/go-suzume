package suzume

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteIfChanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dict.bin")

	// New file: written from scratch.
	if err := writeIfChanged(path, []byte("abc")); err != nil {
		t.Fatalf("write new: %v", err)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "abc" {
		t.Fatalf("after new write got %q, %v", b, err)
	}

	// Unchanged content: no-op, file preserved.
	if err := writeIfChanged(path, []byte("abc")); err != nil {
		t.Fatalf("write unchanged: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "abc" {
		t.Errorf("after unchanged write got %q", b)
	}

	// Changed content: overwritten.
	if err := writeIfChanged(path, []byte("xyz")); err != nil {
		t.Fatalf("write changed: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "xyz" {
		t.Errorf("after changed write got %q", b)
	}
}

func TestMaterializeDicts(t *testing.T) {
	dir, err := materializeDicts()
	if err != nil {
		t.Fatalf("materializeDicts: %v", err)
	}

	for _, name := range embeddedDicts {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("expected %s in %s: %v", name, dir, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s was written empty", name)
		}
	}
}
