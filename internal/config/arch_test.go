package config

import "testing"

func TestResolveArchAuto(t *testing.T) {
	got := ResolveArch("")
	if got != "x64" && got != "x86" {
		t.Fatalf("ResolveArch() = %q", got)
	}
}

func TestResolveArchExplicit(t *testing.T) {
	if got := ResolveArch("x86"); got != "x86" {
		t.Fatalf("ResolveArch(x86) = %q", got)
	}
}
