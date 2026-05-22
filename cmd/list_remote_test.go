package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/ui"
)

func TestToListRemoteTable(t *testing.T) {
	table := toListRemoteTable([]php.RemoteRelease{
		{
			Version:  php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"},
			ZipFile:  "php-8.3.31-nts-Win32-vs16-x64.zip",
			Archived: false,
		},
		{
			Version:  php.Version{Major: 7, Minor: 4, Patch: 33, Arch: "x86", ThreadSafe: true, VC: "vc15"},
			ZipFile:  "php-7.4.33-Win32-vc15-x86.zip",
			Archived: true,
		},
	})

	var buf bytes.Buffer
	ui.SetOutput(&buf, &buf)
	ui.PrintTable(table)

	got := buf.String()
	if !bytes.Contains([]byte(got), []byte("VERSION")) {
		t.Fatalf("table missing header: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("8.3.31")) {
		t.Fatalf("table missing release row: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("archives")) {
		t.Fatalf("table missing archived source: %q", got)
	}
}

func TestToListRemoteJSON(t *testing.T) {
	payload := toListRemoteJSON([]php.RemoteRelease{
		{
			Version: php.Version{Major: 8, Minor: 3, Patch: 31, Arch: "x64", ThreadSafe: false, VC: "vs16"},
			ZipFile: "php-8.3.31-nts-Win32-vs16-x64.zip",
			URL:     "https://example.test/php-8.3.31-nts-Win32-vs16-x64.zip",
		},
	})

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Releases []struct {
			ID      string `json:"id"`
			Version string `json:"version"`
			Source  string `json:"source"`
			URL     string `json:"url"`
		} `json:"releases"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Releases) != 1 {
		t.Fatalf("releases len = %d", len(decoded.Releases))
	}
	if decoded.Releases[0].ID != "8.3.31-x64-nts-vs16" {
		t.Fatalf("id = %q", decoded.Releases[0].ID)
	}
	if decoded.Releases[0].Source != "releases" {
		t.Fatalf("source = %q", decoded.Releases[0].Source)
	}
}
