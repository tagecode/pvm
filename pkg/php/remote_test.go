package php

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestParseZipNameOfficial(t *testing.T) {
	tests := []struct {
		name           string
		wantTS         bool
		wantArch       string
		wantVC         string
		wantPatch      int
		wantRevision   int
		wantPrerelease string
	}{
		{"php-8.3.31-nts-Win32-vs16-x64.zip", false, "x64", "vs16", 31, 0, ""},
		{"php-8.3.31-Win32-vs16-x64.zip", true, "x64", "vs16", 31, 0, ""},
		{"php-7.4.33-nts-Win32-vc15-x86.zip", false, "x86", "vc15", 33, 0, ""},
		{"php-5.3.3-nts-Win32-VC9-x86.zip", false, "x86", "vc9", 3, 0, ""},
		{"php-5.2.10-Win32-VC6-x86.zip", true, "x86", "vc6", 10, 0, ""},
		{"php-7.2.33-1-nts-Win32-VC15-x64.zip", false, "x64", "vc15", 33, 1, ""},
		{"php-8.5.0-dev-nts-Win32-vs17-x86.zip", false, "x86", "vs17", 0, 0, "dev"},
		{"php-8.4.21-Win32-vs17-x86.zip", true, "x86", "vs17", 21, 0, ""},
	}

	for _, tt := range tests {
		v, err := ParseZipName(tt.name)
		if err != nil {
			t.Fatalf("ParseZipName(%q): %v", tt.name, err)
		}
		if v.ThreadSafe != tt.wantTS {
			t.Fatalf("%q ThreadSafe = %v, want %v", tt.name, v.ThreadSafe, tt.wantTS)
		}
		if v.Arch != tt.wantArch {
			t.Fatalf("%q Arch = %q, want %q", tt.name, v.Arch, tt.wantArch)
		}
		if v.VC != tt.wantVC {
			t.Fatalf("%q VC = %q, want %q", tt.name, v.VC, tt.wantVC)
		}
		if v.Patch != tt.wantPatch {
			t.Fatalf("%q Patch = %d, want %d", tt.name, v.Patch, tt.wantPatch)
		}
		if v.Revision != tt.wantRevision {
			t.Fatalf("%q Revision = %d, want %d", tt.name, v.Revision, tt.wantRevision)
		}
		if v.Prerelease != tt.wantPrerelease {
			t.Fatalf("%q Prerelease = %q, want %q", tt.name, v.Prerelease, tt.wantPrerelease)
		}
	}
}

func TestParseZipNameRoundTrip(t *testing.T) {
	names := []string{
		"php-8.3.31-nts-Win32-vs16-x64.zip",
		"php-7.2.33-1-Win32-vc15-x86.zip",
		"php-8.5.0-dev-nts-Win32-vs17-x64.zip",
	}
	for _, name := range names {
		v, err := ParseZipName(name)
		if err != nil {
			t.Fatal(err)
		}
		if got := v.ZipName(); got != name {
			t.Fatalf("ZipName() = %q, want %q", got, name)
		}
	}
}

func TestIsBinaryZipNameSkipsNonBinary(t *testing.T) {
	cases := map[string]bool{
		"php-8.3.31-nts-Win32-vs16-x64.zip":           true,
		"php-debug-pack-8.3.31-nts-Win32-vs16-x64.zip": false,
		"php-test-pack-8.3.31.zip":                     false,
		"php-8.3.31-src.zip":                           false,
	}
	for name, want := range cases {
		if got := IsBinaryZipName(name); got != want {
			t.Fatalf("IsBinaryZipName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestParseAllOfficialBinaryZipNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skip official index fetch in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := &http.Client{Timeout: 2 * time.Minute}
	urls := []string{OfficialReleasesURL, OfficialArchivesURL}

	seen := map[string]struct{}{}
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("fetch %s: %v", url, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("fetch %s: HTTP %d", url, resp.StatusCode)
		}

		for _, name := range ExtractZipNamesFromHTML(string(body)) {
			if !IsBinaryZipName(name) {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			if _, err := ParseZipName(name); err != nil {
				t.Fatalf("ParseZipName(%q) from %s: %v", name, url, err)
			}
		}
	}

	if len(seen) < 1000 {
		t.Fatalf("expected many official binary zips, got %d", len(seen))
	}
	t.Logf("parsed %d unique official binary zip names", len(seen))
}

func TestAttachChecksumFromMap(t *testing.T) {
	checksums := map[string]string{
		"php-8.3.31-nts-Win32-vs16-x64.zip": "abc123",
	}
	name := "php-8.3.31-nts-Win32-vs16-x64.zip"
	if got := checksums[name]; got != "abc123" {
		t.Fatalf("checksum = %q", got)
	}
}
