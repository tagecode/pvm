package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
)

type AliasStore struct {
	Home string
}

func NewAliasStore(home string) *AliasStore {
	return &AliasStore{Home: home}
}

func (s *AliasStore) Set(name, versionID string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return UserError("alias name is required", "")
	}

	reg := NewRegistry(s.Home)
	resolved, err := reg.ResolveID(versionID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(paths.AliasesDir(s.Home), 0o755); err != nil {
		return err
	}
	return os.WriteFile(paths.AliasFile(s.Home, name), []byte(resolved+"\n"), 0o644)
}

func (s *AliasStore) Get(name string) (string, error) {
	data, err := os.ReadFile(paths.AliasFile(s.Home, name))
	if err != nil {
		if os.IsNotExist(err) {
			return "", UserError(fmt.Sprintf("alias %q not found", name), "run pvm alias to list aliases")
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (s *AliasStore) Delete(name string) error {
	if name == "default" {
		return UserError("cannot delete reserved alias default", "use pvm default <version>")
	}
	err := os.Remove(paths.AliasFile(s.Home, name))
	if os.IsNotExist(err) {
		return UserError(fmt.Sprintf("alias %q not found", name), "")
	}
	return err
}

func (s *AliasStore) List() (map[string]string, error) {
	dir := paths.AliasesDir(s.Home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	out := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		out[e.Name()] = strings.TrimSpace(string(data))
	}
	return out, nil
}

// ResolveTarget resolves alias chain to a version ID.
func (s *AliasStore) ResolveTarget(name string) (string, error) {
	seen := map[string]struct{}{}
	current := strings.TrimSpace(name)
	for {
		if _, ok := seen[current]; ok {
			return "", UserError("alias loop detected", "check aliases under "+paths.AliasesDir(s.Home))
		}
		seen[current] = struct{}{}

		if _, err := php.ParseID(current); err == nil {
			return current, nil
		}

		next, err := s.Get(current)
		if err != nil {
			return "", err
		}
		current = next
	}
}
