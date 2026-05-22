package php

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureIniFromDevelopment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "php.ini-development"), []byte("dev=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureIni(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "dev=1\n" {
		t.Fatalf("php.ini = %q", data)
	}
}

func TestEnsureIniFromProductionFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "php.ini-production"), []byte("prod=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureIni(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "php.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "prod=1\n" {
		t.Fatalf("php.ini = %q", data)
	}
}

func TestEnsureIniSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "php.ini"), []byte("custom=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "php.ini-development"), []byte("dev=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureIni(dir); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "php.ini"))
	if string(data) != "custom=1\n" {
		t.Fatalf("php.ini was overwritten: %q", data)
	}
}

func TestEnsureIniNoTemplateFails(t *testing.T) {
	dir := t.TempDir()
	err := EnsureIni(dir)
	if err == nil {
		t.Fatal("expected error when no ini templates exist")
	}
}

func TestEnsureIniPrefersDevelopment(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "php.ini-development"), []byte("dev=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "php.ini-production"), []byte("prod=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureIni(dir); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "php.ini"))
	if string(data) != "dev=1\n" {
		t.Fatalf("php.ini = %q, want development template", data)
	}
}
