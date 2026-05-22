//go:build windows

package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUseSwitchWindows(t *testing.T) {
	home := t.TempDir()
	id := "8.3.31-x64-nts-vs16"
	versionDir := filepath.Join(home, "versions", id)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "php.exe"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	sw := NewSwitcher(home)
	if _, err := sw.Use(t.Context(), "8.3.31", UseOptions{}); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry(home)
	cur, err := reg.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if !cur.Active {
		t.Fatal("expected active version")
	}
}
