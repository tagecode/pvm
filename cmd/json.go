package cmd

import (
	"fmt"

	"tagecode/pvm/internal/app"
	"tagecode/pvm/pkg/paths"
)

type listVersionJSON struct {
	ID         string `json:"id"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	ThreadSafe bool   `json:"thread_safe"`
	VC         string `json:"vc"`
	Path       string `json:"path"`
	Active     bool   `json:"active"`
	Default    bool   `json:"default"`
}

func toListJSON(items []app.InstalledVersion) map[string]any {
	out := make([]listVersionJSON, 0, len(items))
	for _, v := range items {
		out = append(out, listVersionJSON{
			ID:         v.ID,
			Version:    fmt.Sprintf("%d.%d.%d", v.Version.Major, v.Version.Minor, v.Version.Patch),
			Arch:       v.Version.Arch,
			ThreadSafe: v.Version.ThreadSafe,
			VC:         v.Version.VC,
			Path:       v.Path,
			Active:     v.Active,
			Default:    v.Default,
		})
	}
	return map[string]any{"versions": out}
}

func toCurrentJSON(home string, v *app.InstalledVersion) map[string]any {
	return map[string]any{
		"id":      v.ID,
		"version": fmt.Sprintf("%d.%d.%d", v.Version.Major, v.Version.Minor, v.Version.Patch),
		"path":    paths.CurrentLink(home),
		"php_exe": paths.CurrentLink(home) + `\php.exe`,
	}
}
