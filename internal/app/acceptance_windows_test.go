//go:build windows

package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"tagecode/pvm/pkg/download"
	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/win"
)

func docsZip(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "docs", "zip", name))
}

func TestAcceptanceOfflineInstallUseSwitch(t *testing.T) {
	ntsZip := docsZip("php-7.4.33-nts-Win32-vc15-x86.zip")
	tsZip := docsZip("php-7.4.33-Win32-vc15-x86.zip")
	if _, err := os.Stat(ntsZip); err != nil {
		t.Skip("docs/zip fixtures not available")
	}

	home := t.TempDir()
	inst := &Installer{Home: home}
	sw := NewSwitcher(home)
	ctx := t.Context()

	installZip := func(spec string, zip string) *InstalledVersion {
		t.Helper()
		s, err := php.ParseSpec(spec)
		if err != nil {
			t.Fatal(err)
		}
		iv, err := inst.Install(ctx, s, InstallOptions{FromZip: zip})
		if err != nil {
			t.Fatal(err)
		}
		return iv
	}

	nts := installZip("7.4.33", ntsZip)
	ts := installZip("7.4.33", tsZip)

	reg := NewRegistry(home)
	list, err := reg.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}

	runPHP := func(wantNTS bool) {
		t.Helper()
		out, err := exec.Command(filepath.Join(paths.CurrentLink(home), "php.exe"), "-v").CombinedOutput()
		if err != nil {
			t.Fatalf("php -v: %v: %s", err, out)
		}
		text := string(out)
		if !strings.Contains(text, "7.4.33") {
			t.Fatalf("php -v = %q", text)
		}
		if wantNTS && strings.Contains(text, "ZTS") {
			t.Fatalf("expected NTS, got %q", text)
		}
		if !wantNTS && !strings.Contains(text, "ZTS") {
			t.Fatalf("expected TS, got %q", text)
		}
	}

	for _, step := range []struct {
		id      string
		wantNTS bool
	}{
		{nts.ID, true},
		{ts.ID, false},
		{nts.ID, true},
	} {
		if _, err := sw.Use(ctx, step.id, UseOptions{}); err != nil {
			t.Fatalf("use %s: %v", step.id, err)
		}
		runPHP(step.wantNTS)
	}
}

func TestAcceptanceAliasUseProd(t *testing.T) {
	zip := docsZip("php-7.4.33-nts-Win32-vc15-x86.zip")
	if _, err := os.Stat(zip); err != nil {
		t.Skip("docs/zip fixtures not available")
	}

	home := t.TempDir()
	inst := &Installer{Home: home}
	spec, err := php.ParseSpec("7.4.33")
	if err != nil {
		t.Fatal(err)
	}
	iv, err := inst.Install(t.Context(), spec, InstallOptions{FromZip: zip})
	if err != nil {
		t.Fatal(err)
	}

	store := NewAliasStore(home)
	if err := store.Set("prod", iv.ID); err != nil {
		t.Fatal(err)
	}

	sw := NewSwitcher(home)
	if _, err := sw.Use(t.Context(), "prod", UseOptions{}); err != nil {
		t.Fatal(err)
	}

	cur, err := sw.Current()
	if err != nil {
		t.Fatal(err)
	}
	if cur.ID != iv.ID {
		t.Fatalf("current = %q, want %q", cur.ID, iv.ID)
	}
}

func TestAcceptanceOnlineInstallMock(t *testing.T) {
	v := php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	zipName := v.ZipName()
	zipPath := buildFixtureZip(t, strings.TrimSuffix(zipName, ".zip"))

	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprintf(w, `<a href="%s">%s</a>`, zipName, zipName)
		case "/sha256sum.txt":
			fmt.Fprintf(w, "%s *%s\n", checksum, zipName)
		case "/" + zipName:
			http.ServeFile(w, r, zipPath)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	home := t.TempDir()
	viper.Set("security.verify_sha256", true)

	inst := &Installer{
		Home:       home,
		Remote:     php.NewRemoteIndex(srv.URL + "/"),
		Downloader: download.NewClient(),
	}

	spec, err := php.ParseSpec("8.3.31")
	if err != nil {
		t.Fatal(err)
	}
	iv, err := inst.Install(context.Background(), spec, InstallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if iv.ID != v.ID() {
		t.Fatalf("id = %q, want %q", iv.ID, v.ID())
	}

	reg := NewRegistry(home)
	list, err := reg.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list = %+v", list)
	}
}

func TestAcceptanceSetupUnsetupPreservesOtherPath(t *testing.T) {
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

	systemPHP := `C:\Windows\System32\fake-system-php`
	pathWithSystem := win.PrependPathValue(origPath, systemPHP)
	if err := win.SetUserEnv("Path", pathWithSystem); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	if err := paths.EnsureLayout(home); err != nil {
		t.Fatal(err)
	}
	current := paths.CurrentLink(home)
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}

	setup := NewSetupService(home)
	if err := setup.Run(t.Context(), SetupOptions{}); err != nil {
		t.Fatal(err)
	}

	status, err := setup.Status()
	if err != nil {
		t.Fatal(err)
	}
	if !status.InPath || !status.PathAtFront {
		t.Fatalf("status = %+v", status)
	}
	if !strings.Contains(strings.ToLower(status.Path), strings.ToLower(systemPHP)) {
		t.Fatalf("system PATH entry removed: %q", status.Path)
	}
	if !strings.EqualFold(status.PHPHome, current) {
		t.Fatalf("PHP_HOME = %q, want %q", status.PHPHome, current)
	}

	if err := setup.Unsetup(t.Context(), SetupOptions{}); err != nil {
		t.Fatal(err)
	}

	after, err := win.GetUserEnv("Path")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(after), strings.ToLower(current)) {
		t.Fatalf("PVM path still present after unsetup: %q", after)
	}
	if !strings.Contains(strings.ToLower(after), strings.ToLower(systemPHP)) {
		t.Fatalf("system PATH entry lost after unsetup: %q", after)
	}
}

func TestAcceptanceRefreshScript(t *testing.T) {
	home := t.TempDir()
	setup := NewSetupService(home)
	ps, cmd := setup.RefreshScript()
	current := paths.CurrentLink(home)
	if !strings.Contains(ps, current) || !strings.Contains(cmd, current) {
		t.Fatalf("refresh scripts missing current link: ps=%q cmd=%q", ps, cmd)
	}
}

func TestAcceptanceInstallArchX86FromZip(t *testing.T) {
	zip := docsZip("php-7.4.33-nts-Win32-vc15-x86.zip")
	if _, err := os.Stat(zip); err != nil {
		t.Skip("docs/zip fixtures not available")
	}

	home := t.TempDir()
	inst := &Installer{Home: home}
	spec, err := php.ParseSpec("7.4.33")
	if err != nil {
		t.Fatal(err)
	}
	iv, err := inst.Install(t.Context(), spec, InstallOptions{FromZip: zip, Arch: "x86"})
	if err != nil {
		t.Fatal(err)
	}
	if iv.ID != "7.4.33-x86-nts-vc15" {
		t.Fatalf("id = %q", iv.ID)
	}
}

func TestAcceptanceInstallEnsuresPhpIni(t *testing.T) {
	zip := docsZip("php-7.4.33-nts-Win32-vc15-x86.zip")
	if _, err := os.Stat(zip); err != nil {
		t.Skip("docs/zip fixtures not available")
	}

	home := t.TempDir()
	inst := &Installer{Home: home}
	spec, err := php.ParseSpec("7.4.33")
	if err != nil {
		t.Fatal(err)
	}
	iv, err := inst.Install(t.Context(), spec, InstallOptions{FromZip: zip})
	if err != nil {
		t.Fatal(err)
	}
	iniPath := filepath.Join(iv.Path, "php.ini")
	if _, err := os.Stat(iniPath); err != nil {
		t.Fatalf("php.ini missing after install: %v", err)
	}
}

func TestAcceptanceInstallFailsWithoutIniTemplate(t *testing.T) {
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

func TestParseDocsZip533(t *testing.T) {
	v, err := php.ParseZipName("php-5.3.3-nts-Win32-VC9-x86.zip")
	if err != nil {
		t.Fatal(err)
	}
	if v.VC != "vc9" || v.Arch != "x86" || v.Major != 5 {
		t.Fatalf("parsed = %+v", v)
	}
}
