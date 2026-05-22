package php

import (
	"fmt"
	"testing"
)

func TestParseSpec(t *testing.T) {
	tests := []struct {
		raw  string
		kind SpecKind
	}{
		{"latest", SpecLatest},
		{"8.3", SpecMinor},
		{"8.3.10", SpecExact},
		{"prod", SpecAlias},
	}

	for _, tt := range tests {
		spec, err := ParseSpec(tt.raw)
		if err != nil {
			t.Fatalf("ParseSpec(%q): %v", tt.raw, err)
		}
		if spec.Kind != tt.kind {
			t.Fatalf("ParseSpec(%q) kind = %v, want %v", tt.raw, spec.Kind, tt.kind)
		}
	}
}

func TestVersionID(t *testing.T) {
	v := Version{Major: 8, Minor: 3, Patch: 10, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	if got, want := v.ID(), "8.3.10-x64-nts-vs16"; got != want {
		t.Fatalf("ID() = %q, want %q", got, want)
	}
}

func TestZipNameNTSx64(t *testing.T) {
	v := Version{Major: 8, Minor: 3, Patch: 10, Arch: "x64", ThreadSafe: false, VC: "vs16"}
	if got, want := v.ZipName(), "php-8.3.10-nts-Win32-vs16-x64.zip"; got != want {
		t.Fatalf("ZipName() = %q, want %q", got, want)
	}
}

func TestZipNameTSx64(t *testing.T) {
	v := Version{Major: 8, Minor: 3, Patch: 10, Arch: "x64", ThreadSafe: true, VC: "vs16"}
	if got, want := v.ZipName(), "php-8.3.10-Win32-vs16-x64.zip"; got != want {
		t.Fatalf("ZipName() = %q, want %q", got, want)
	}
}

func TestDefaultVC(t *testing.T) {
	cases := map[string]string{
		"5.2": "vc6",
		"5.3": "vc9",
		"5.6": "vc11",
		"7.1": "vc14",
		"7.4": "vc15",
		"8.3": "vs16",
		"8.4": "vs17",
	}
	for ver, want := range cases {
		var major, minor int
		fmt.Sscanf(ver, "%d.%d", &major, &minor)
		if got := DefaultVC(major, minor); got != want {
			t.Fatalf("DefaultVC(%d,%d) = %q, want %q", major, minor, got, want)
		}
	}
}

func TestParseIDWithRevisionAndDev(t *testing.T) {
	v, err := ParseID("7.2.33-1-x64-nts-vc15")
	if err != nil {
		t.Fatal(err)
	}
	if v.Revision != 1 || v.Arch != "x64" {
		t.Fatalf("unexpected: %+v", v)
	}

	v, err = ParseID("8.5.0-dev-x86-ts-vs17")
	if err != nil {
		t.Fatal(err)
	}
	if v.Prerelease != "dev" || !v.ThreadSafe {
		t.Fatalf("unexpected: %+v", v)
	}
}

func TestParseID(t *testing.T) {
	v, err := ParseID("8.3.10-x64-nts-vs16")
	if err != nil {
		t.Fatal(err)
	}
	if v.Major != 8 || v.Minor != 3 || v.Patch != 10 {
		t.Fatalf("unexpected version: %+v", v)
	}
}
