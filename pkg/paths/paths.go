package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultHomeDir = ".pvm"

// Home returns PVM data root. Uses PVM_HOME env or %USERPROFILE%\.pvm.
func Home() (string, error) {
	if v := os.Getenv("PVM_HOME"); v != "" {
		return filepath.Clean(v), nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(userHome, defaultHomeDir), nil
}

func ConfigFile(home string) string  { return filepath.Join(home, "config.toml") }
func VersionsDir(home string) string { return filepath.Join(home, "versions") }
func CurrentLink(home string) string { return filepath.Join(home, "current") }
func AliasesDir(home string) string  { return filepath.Join(home, "aliases") }
func CacheDir(home string) string    { return filepath.Join(home, "cache") }
func LogsDir(home string) string     { return filepath.Join(home, "logs") }
func StateDir(home string) string    { return filepath.Join(home, "state") }
func LogFile(home string) string     { return filepath.Join(LogsDir(home), "pvm.log") }
func ActiveVersionFile(home string) string {
	return filepath.Join(StateDir(home), "active")
}

func ReadActiveVersion(home string) (string, error) {
	data, err := os.ReadFile(ActiveVersionFile(home))
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", fmt.Errorf("empty active version")
	}
	return id, nil
}

func WriteActiveVersion(home, id string) error {
	if err := os.MkdirAll(StateDir(home), 0o755); err != nil {
		return err
	}
	return os.WriteFile(ActiveVersionFile(home), []byte(id+"\n"), 0o644)
}

func VersionDir(home, id string) string {
	return filepath.Join(VersionsDir(home), id)
}

func AliasFile(home, name string) string {
	return filepath.Join(AliasesDir(home), name)
}

// EnsureLayout creates required PVM_HOME subdirectories.
func EnsureLayout(home string) error {
	dirs := []string{
		home,
		VersionsDir(home),
		AliasesDir(home),
		CacheDir(home),
		LogsDir(home),
		StateDir(home),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}
	return nil
}
