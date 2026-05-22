package app

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/win"
)

type InstalledVersion struct {
	Version   php.Version `json:"version"`
	ID        string      `json:"id"`
	Path      string      `json:"path"`
	Active    bool        `json:"active"`
	Default   bool        `json:"default"`
	Installed time.Time   `json:"installed_at,omitempty"`
}

type Registry struct {
	Home string
}

func NewRegistry(home string) *Registry {
	return &Registry{Home: home}
}

func (r *Registry) List() ([]InstalledVersion, error) {
	dir := paths.VersionsDir(r.Home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	activeID, _ := r.activeID()
	defaultID, _ := r.defaultID()

	var out []InstalledVersion
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".tmp") {
			continue
		}
		v, err := php.ParseID(name)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(paths.VersionDir(r.Home, name), "php.exe")); err != nil {
			continue
		}
		p := paths.VersionDir(r.Home, name)
		info, _ := e.Info()
		item := InstalledVersion{
			Version: v,
			ID:      name,
			Path:    p,
			Active:  strings.EqualFold(name, activeID),
			Default: strings.EqualFold(name, defaultID),
		}
		if info != nil {
			item.Installed = info.ModTime()
		}
		out = append(out, item)
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Version, out[j].Version
		if a.Major != b.Major {
			return a.Major > b.Major
		}
		if a.Minor != b.Minor {
			return a.Minor > b.Minor
		}
		return a.Patch > b.Patch
	})
	return out, nil
}

func (r *Registry) Get(id string) (*InstalledVersion, error) {
	list, err := r.List()
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if strings.EqualFold(v.ID, id) {
			return &v, nil
		}
	}
	return nil, ErrNotInstalled
}

// ResolveID resolves a version id, alias, or partial version (8.3 / 8.3.31) to an installed ID.
func (r *Registry) ResolveID(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", UserError("version is required", "")
	}

	if _, err := php.ParseID(target); err == nil {
		if _, statErr := os.Stat(paths.VersionDir(r.Home, target)); statErr == nil {
			return target, nil
		}
	}

	if spec, err := php.ParseSpec(target); err == nil && spec.Kind != php.SpecAlias {
		if id, findErr := r.resolveInstalledSpec(spec); findErr == nil {
			return id, nil
		}
	}

	store := NewAliasStore(r.Home)
	if id, err := store.ResolveTarget(target); err == nil {
		if _, statErr := os.Stat(paths.VersionDir(r.Home, id)); statErr == nil {
			return id, nil
		}
	}

	return "", ErrNotInstalled
}

func (r *Registry) resolveInstalledSpec(spec php.VersionSpec) (string, error) {
	installed, err := r.List()
	if err != nil {
		return "", err
	}

	switch spec.Kind {
	case php.SpecExact, php.SpecMinor:
		for _, iv := range installed {
			v := iv.Version
			if v.Major != spec.Major || v.Minor != spec.Minor {
				continue
			}
			if spec.Kind == php.SpecMinor || (v.Patch == spec.Patch) {
				return iv.ID, nil
			}
		}
	case php.SpecLatest:
		if len(installed) > 0 {
			return installed[0].ID, nil
		}
	}
	return "", ErrNotInstalled
}

func (r *Registry) activeID() (string, error) {
	if id, err := paths.ReadActiveVersion(r.Home); err == nil {
		return id, nil
	}

	link := paths.CurrentLink(r.Home)
	target, err := win.ReadLinkTarget(link)
	if err != nil {
		return "", nil
	}
	return filepath.Base(filepath.Clean(target)), nil
}

func (r *Registry) defaultID() (string, error) {
	store := NewAliasStore(r.Home)
	id, err := store.Get("default")
	if err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			return "", nil
		}
		return "", err
	}
	return id, nil
}
