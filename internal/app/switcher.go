package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"tagecode/pvm/internal/config"
	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
	"tagecode/pvm/pkg/win"
)

type UseOptions struct {
	DryRun bool
}

type UseResult struct {
	ID           string
	SetupApplied bool
}

type Switcher struct {
	Home string
}

func NewSwitcher(home string) *Switcher {
	return &Switcher{Home: home}
}

func (s *Switcher) Use(ctx context.Context, target string, opts UseOptions) (UseResult, error) {
	var result UseResult

	reg := NewRegistry(s.Home)
	id, err := reg.ResolveID(target)
	if err != nil {
		return result, err
	}

	versionDir := paths.VersionDir(s.Home, id)
	if _, err := os.Stat(filepath.Join(versionDir, "php.exe")); err != nil {
		return result, ErrNotInstalled
	}

	result.ID = id
	if opts.DryRun {
		return result, nil
	}

	if err := paths.EnsureLayout(s.Home); err != nil {
		return result, err
	}

	link := paths.CurrentLink(s.Home)
	mode := win.LinkMode(config.LinkMode())
	if err := win.CreateLink(link, versionDir, mode); err != nil {
		return result, SystemError("failed to switch version", err.Error())
	}
	if err := paths.WriteActiveVersion(s.Home, id); err != nil {
		return result, SystemError("failed to record active version", err.Error())
	}

	setup := NewSetupService(s.Home)
	applied, err := setup.EnsureUserEnvironment(ctx, SetupOptions{})
	if err != nil {
		return result, err
	}
	result.SetupApplied = applied

	win.RefreshSessionEnv(link)
	return result, nil
}

func (s *Switcher) Current() (*InstalledVersion, error) {
	reg := NewRegistry(s.Home)
	activeID, err := reg.activeID()
	if err != nil || activeID == "" {
		return nil, UserError("no active PHP version", "run pvm install && pvm use <version>")
	}
	return reg.Get(activeID)
}

func (s *Switcher) Which(target string) (string, error) {
	reg := NewRegistry(s.Home)
	id, err := reg.ResolveID(target)
	if err != nil {
		return "", err
	}
	p := paths.VersionDir(s.Home, id)
	if _, err := os.Stat(p); err != nil {
		return "", ErrNotInstalled
	}
	return p, nil
}

func (s *Switcher) UseMessage(result UseResult) string {
	v, err := php.ParseID(result.ID)
	if err != nil {
		return fmt.Sprintf("Now using %s", result.ID)
	}
	msg := fmt.Sprintf("Now using PHP %s", v.String())
	if result.SetupApplied {
		msg += " (environment configured)"
	}
	return msg
}
