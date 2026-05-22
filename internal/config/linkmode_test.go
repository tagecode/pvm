package config

import "testing"

func TestValidateLinkMode(t *testing.T) {
	if err := ValidateLinkMode("junction"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLinkMode("invalid"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetLinkMode(t *testing.T) {
	t.Setenv("PVM_HOME", t.TempDir())
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	if err := SetLinkMode("junction"); err != nil {
		t.Fatal(err)
	}
	if LinkMode() != "junction" {
		t.Fatalf("link mode = %q", LinkMode())
	}
}
