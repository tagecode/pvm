package win

import (
	"os"
	"strings"
	"testing"
)

func TestRefreshSessionEnv(t *testing.T) {
	current := `C:\pvm\current`
	_ = os.Setenv("Path", `C:\Windows\System32`)
	_ = os.Setenv("PHP_HOME", "")

	RefreshSessionEnv(current)

	if got := os.Getenv("PHP_HOME"); got != current {
		t.Fatalf("PHP_HOME = %q, want %q", got, current)
	}
	path := os.Getenv("Path")
	if !strings.HasPrefix(path, current+";") {
		t.Fatalf("Path = %q, want prefix %q", path, current+";")
	}
}
