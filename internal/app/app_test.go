package app

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
)

func init() {
	viper.SetDefault("security.verify_sha256", true)
}

func TestRegistryListAndResolve(t *testing.T) {
	home := t.TempDir()
	id := "8.3.31-x64-nts-vs16"
	versionDir := filepath.Join(home, "versions", id)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "php.exe"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	reg := NewRegistry(home)
	list, err := reg.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("list = %+v", list)
	}

	got, err := reg.ResolveID("8.3.31")
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("ResolveID(8.3.31) = %q", got)
	}

	got, err = reg.ResolveID("8.3")
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("ResolveID(8.3) = %q", got)
	}
}

func TestAliasStoreResolveLoop(t *testing.T) {
	home := t.TempDir()
	aliasesDir := filepath.Join(home, "aliases")
	if err := os.MkdirAll(aliasesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aliasesDir, "a"), []byte("b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aliasesDir, "b"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewAliasStore(home)
	if _, err := store.ResolveTarget("a"); err == nil {
		t.Fatal("expected alias loop error")
	}
}

func TestAliasStoreSet(t *testing.T) {
	home := t.TempDir()
	id := "8.3.31-x64-nts-vs16"
	versionDir := filepath.Join(home, "versions", id)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "php.exe"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	store := NewAliasStore(home)
	if err := store.Set("prod", "8.3.31"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("prod")
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("prod -> %q", got)
	}
}

func buildFixtureZip(t *testing.T, root string) string {
	t.Helper()
	return writeFixtureZip(t, root, map[string]string{
		root + "/php.exe":             "stub",
		root + "/php.ini-development": "development=1",
	})
}

func buildFixtureZipWithoutIniTemplate(t *testing.T, root string) string {
	t.Helper()
	return writeFixtureZip(t, root, map[string]string{
		root + "/php.exe": "stub",
	})
}

func writeFixtureZip(t *testing.T, root string, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "php-fixture.zip")
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUninstallActiveVersion(t *testing.T) {
	home := t.TempDir()
	id := "8.3.31-x64-nts-vs16"
	versionDir := filepath.Join(home, "versions", id)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "php.exe"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := paths.WriteActiveVersion(home, id); err != nil {
		t.Fatal(err)
	}

	inst := &Installer{Home: home}
	if err := inst.Uninstall(t.Context(), id); err == nil {
		t.Fatal("expected error uninstalling active version")
	}
}

func TestChecksumVerificationFails(t *testing.T) {
	home := t.TempDir()
	v := php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	zipPath := buildFixtureZip(t, "php-8.3.31-nts-Win32-vs16-x64")

	inst := &Installer{Home: home}
	release := php.RemoteRelease{
		Version:      v,
		ZipFile:      filepath.Base(zipPath),
		Checksum:     "deadbeef",
		ChecksumAlgo: "sha256",
	}
	_, err := inst.InstallRelease(t.Context(), release, InstallOptions{FromZip: zipPath})
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestInstallReleaseFromZip(t *testing.T) {
	home := t.TempDir()
	v := php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	id := v.ID()
	zipPath := buildFixtureZip(t, "php-8.3.31-nts-Win32-vs16-x64")

	inst := &Installer{Home: home}
	release := php.RemoteRelease{Version: v, ZipFile: filepath.Base(zipPath)}
	iv, err := inst.InstallRelease(t.Context(), release, InstallOptions{FromZip: zipPath})
	if err != nil {
		t.Fatal(err)
	}
	if iv.ID != id {
		t.Fatalf("id = %q", iv.ID)
	}
	if _, err := os.Stat(filepath.Join(home, "versions", id, "php.exe")); err != nil {
		t.Fatal(err)
	}
	iniData, err := os.ReadFile(filepath.Join(home, "versions", id, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(iniData), "development=1") {
		t.Fatalf("php.ini = %q", iniData)
	}
}

func TestInstallReleaseFailsWithoutIniTemplate(t *testing.T) {
	home := t.TempDir()
	v := php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	zipPath := buildFixtureZipWithoutIniTemplate(t, "php-8.3.31-nts-Win32-vs16-x64")

	inst := &Installer{Home: home}
	release := php.RemoteRelease{Version: v, ZipFile: filepath.Base(zipPath)}
	_, err := inst.InstallRelease(t.Context(), release, InstallOptions{FromZip: zipPath})
	if err == nil {
		t.Fatal("expected install to fail without ini template")
	}
}
