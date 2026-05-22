//go:build windows

package win

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func openUserEnv(write bool) (registry.Key, error) {
	if write {
		return registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	}
	return registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
}

func GetUserEnv(key string) (string, error) {
	k, err := openUserEnv(false)
	if err != nil {
		return "", err
	}
	defer k.Close()

	val, _, err := k.GetStringValue(key)
	if err == registry.ErrNotExist {
		return "", nil
	}
	return val, err
}

func DeleteUserEnv(key string) error {
	k, err := openUserEnv(true)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.DeleteValue(key)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

func pathContainsEntry(path, entry string) (bool, int) {
	entry = strings.TrimSpace(entry)
	for i, p := range strings.Split(path, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.EqualFold(p, entry) {
			return true, i
		}
	}
	return false, -1
}

// PrependUserPath puts entry at the front of the user PATH (moves it if already present).
func PrependUserPath(entry string) error {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return fmt.Errorf("empty PATH entry")
	}

	k, err := openUserEnv(true)
	if err != nil {
		return fmt.Errorf("open user environment: %w", err)
	}
	defer k.Close()

	path, _, err := k.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("read Path: %w", err)
	}

	path = PrependPathValue(path, entry)
	return k.SetStringValue("Path", path)
}

func SetUserEnv(key, value string) error {
	k, err := openUserEnv(true)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(key, value)
}

func RemoveUserPathEntry(entry string) error {
	k, err := openUserEnv(true)
	if err != nil {
		return err
	}
	defer k.Close()

	path, _, err := k.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}

	var kept []string
	for _, p := range strings.Split(path, ";") {
		if strings.EqualFold(strings.TrimSpace(p), strings.TrimSpace(entry)) {
			continue
		}
		if strings.TrimSpace(p) != "" {
			kept = append(kept, p)
		}
	}
	return k.SetStringValue("Path", strings.Join(kept, ";"))
}

func GetEnvStatus(pvmHome, currentLink string) (*EnvStatus, error) {
	userPath, _ := GetUserEnv("Path")
	phpHome, _ := GetUserEnv("PHP_HOME")

	inPath, idx := pathContainsEntry(userPath, currentLink)

	return &EnvStatus{
		PVMHome:     pvmHome,
		PHPHome:     phpHome,
		PHPRC:       os.Getenv("PHPRC"),
		Path:        userPath,
		InPath:      inPath,
		PathIndex:   idx,
		PathAtFront: inPath && idx == 0,
	}, nil
}

// NeedsUserEnvironmentSetup reports whether user PATH/PHP_HOME still need setup for current.
func NeedsUserEnvironmentSetup(currentLink string) (bool, error) {
	currentLink = strings.TrimSpace(currentLink)
	if currentLink == "" {
		return false, fmt.Errorf("empty current link")
	}
	userPath, err := GetUserEnv("Path")
	if err != nil {
		return false, err
	}
	phpHome, err := GetUserEnv("PHP_HOME")
	if err != nil {
		return false, err
	}
	inPath, idx := pathContainsEntry(userPath, currentLink)
	if !inPath || idx != 0 {
		return true, nil
	}
	if !strings.EqualFold(strings.TrimSpace(phpHome), currentLink) {
		return true, nil
	}
	return false, nil
}

// RefreshSessionEnv updates the current process PATH and PHP_HOME for immediate php usage.
func RefreshSessionEnv(currentLink string) {
	currentLink = strings.TrimSpace(currentLink)
	if currentLink == "" {
		return
	}
	_ = os.Setenv("PHP_HOME", currentLink)
	path := os.Getenv("Path")
	_ = os.Setenv("Path", PrependPathValue(path, currentLink))
}
