//go:build windows

package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/win"
)

func TestUseAutoSetupAndSessionRefresh(t *testing.T) {
	origPath, err := win.GetUserEnv("Path")
	if err != nil {
		t.Fatal(err)
	}
	origPHPHome, err := win.GetUserEnv("PHP_HOME")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = win.SetUserEnv("Path", origPath)
		if origPHPHome == "" {
			_ = win.DeleteUserEnv("PHP_HOME")
		} else {
			_ = win.SetUserEnv("PHP_HOME", origPHPHome)
		}
	})

	home := t.TempDir()
	id := "8.3.31-x64-nts-vs16"
	versionDir := filepath.Join(home, "versions", id)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "php.exe"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	current := paths.CurrentLink(home)
	_ = win.RemoveUserPathEntry(current)
	_ = win.DeleteUserEnv("PHP_HOME")

	sw := NewSwitcher(home)
	result, err := sw.Use(t.Context(), id, UseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.SetupApplied {
		t.Fatal("expected auto setup on first use")
	}

	st, err := win.GetEnvStatus(home, current)
	if err != nil {
		t.Fatal(err)
	}
	if !st.InPath || !st.PathAtFront {
		t.Fatalf("user PATH not configured: %+v", st)
	}
	if !pathsEqualFold(st.PHPHome, current) {
		t.Fatalf("PHP_HOME = %q, want %q", st.PHPHome, current)
	}
	if got := os.Getenv("PHP_HOME"); !pathsEqualFold(got, current) {
		t.Fatalf("session PHP_HOME = %q, want %q", got, current)
	}
}

func TestAutoUseAfterInstallDefault(t *testing.T) {
	viper.Set("defaults.auto_use_after_install", true)
	t.Cleanup(func() { viper.Set("defaults.auto_use_after_install", true) })

	home := t.TempDir()
	v := php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	zipPath := buildFixtureZip(t, "php-8.3.31-nts-Win32-vs16-x64")

	inst := &Installer{Home: home}
	release := php.RemoteRelease{Version: v, ZipFile: filepath.Base(zipPath)}
	iv, err := inst.InstallRelease(t.Context(), release, InstallOptions{FromZip: zipPath})
	if err != nil {
		t.Fatal(err)
	}
	if !iv.Active {
		t.Fatal("expected auto use after install")
	}
}

func pathsEqualFold(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}
