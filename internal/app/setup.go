package app

import (
	"context"
	"fmt"
	"strings"

	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/win"
)

type SetupOptions struct {
	System bool
	DryRun bool
}

type SetupService struct {
	Home string
}

func NewSetupService(home string) *SetupService {
	return &SetupService{Home: home}
}

func (s *SetupService) Run(_ context.Context, opts SetupOptions) error {
	if opts.System {
		return SystemError("system-level setup is not implemented yet", "use pvm setup without --system for now")
	}
	if err := paths.EnsureLayout(s.Home); err != nil {
		return err
	}
	current := paths.CurrentLink(s.Home)
	if opts.DryRun {
		return nil
	}
	return configureUserEnvironment(current)
}

// EnsureUserEnvironment writes user PATH/PHP_HOME when current is missing or misordered.
func (s *SetupService) EnsureUserEnvironment(_ context.Context, opts SetupOptions) (applied bool, err error) {
	if opts.DryRun {
		return false, nil
	}
	current := paths.CurrentLink(s.Home)
	needed, err := win.NeedsUserEnvironmentSetup(current)
	if err != nil {
		return false, err
	}
	if !needed {
		return false, nil
	}
	if err := configureUserEnvironment(current); err != nil {
		return false, err
	}
	return true, nil
}

func configureUserEnvironment(current string) error {
	if err := win.PrependUserPath(current); err != nil {
		return SystemError("failed to update user PATH", err.Error())
	}
	if err := win.SetUserEnv("PHP_HOME", current); err != nil {
		return SystemError("failed to set PHP_HOME", err.Error())
	}
	return nil
}

func (s *SetupService) Unsetup(_ context.Context, opts SetupOptions) error {
	if opts.System {
		return SystemError("system-level unsetup is not implemented yet", "use pvm unsetup without --system for now")
	}
	current := paths.CurrentLink(s.Home)
	if opts.DryRun {
		return nil
	}

	if err := win.RemoveUserPathEntry(current); err != nil {
		return SystemError("failed to update user PATH", err.Error())
	}

	phpHome, _ := win.GetUserEnv("PHP_HOME")
	if strings.EqualFold(strings.TrimSpace(phpHome), current) {
		if err := win.DeleteUserEnv("PHP_HOME"); err != nil {
			return SystemError("failed to remove PHP_HOME", err.Error())
		}
	}
	return nil
}

func (s *SetupService) Status() (*win.EnvStatus, error) {
	return win.GetEnvStatus(s.Home, paths.CurrentLink(s.Home))
}

func (s *SetupService) RefreshScript() (powershell, cmd string) {
	current := paths.CurrentLink(s.Home)
	ps := fmt.Sprintf(`$env:Path = "%s;" + $env:Path; $env:PHP_HOME = "%s"`, current, current)
	c := fmt.Sprintf(`set PATH=%s;%%PATH%% && set PHP_HOME=%s`, current, current)
	return ps, c
}
