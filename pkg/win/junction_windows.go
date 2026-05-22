//go:build windows

package win

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LinkMode string

const (
	LinkModeJunction LinkMode = "junction"
	LinkModeSymlink  LinkMode = "symlink"
	LinkModeCopy     LinkMode = "copy"
)

func CreateLink(link, target string, mode LinkMode) error {
	link = filepath.Clean(link)
	target = filepath.Clean(target)

	if err := os.RemoveAll(link); err != nil {
		return fmt.Errorf("remove existing link %s: %w", link, err)
	}

	switch mode {
	case LinkModeJunction, "":
		return createJunction(link, target)
	case LinkModeSymlink:
		return createSymlink(link, target)
	case LinkModeCopy:
		return copyTree(link, target)
	default:
		return fmt.Errorf("unsupported link mode %q", mode)
	}
}

func RemoveLink(link string) error {
	return os.RemoveAll(link)
}

func createJunction(link, target string) error {
	cmd := exec.Command("cmd", "/c", "mklink", "/J", link, target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink /J: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func createSymlink(link, target string) error {
	cmd := exec.Command("cmd", "/c", "mklink", "/D", link, target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink /D: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func copyTree(link, target string) error {
	if err := os.MkdirAll(link, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("robocopy", target, link, "/MIR", "/NFL", "/NDL", "/NJH", "/NJS", "/NC", "/NS")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code := exit.ExitCode()
			if code >= 0 && code < 8 {
				return nil
			}
		}
		return fmt.Errorf("robocopy: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func ReadLinkTarget(link string) (string, error) {
	if target, err := filepath.EvalSymlinks(link); err == nil && target != "" {
		return filepath.Clean(target), nil
	}

	fi, err := os.Lstat(link)
	if err != nil {
		return "", err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(link)
		if err != nil {
			return "", err
		}
		return filepath.Clean(target), nil
	}

	// Junction fallback via fsutil.
	cmd := exec.Command("cmd", "/c", "fsutil", "reparsepoint", "query", link)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve link target: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Substitute Name:") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "Substitute Name:"))
			raw = strings.TrimPrefix(raw, `\??\`)
			return filepath.Clean(raw), nil
		}
	}
	return "", fmt.Errorf("could not resolve link target for %s", link)
}
