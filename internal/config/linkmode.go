package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"tagecode/pvm/pkg/paths"
)

const (
	LinkModeJunction = "junction"
	LinkModeSymlink  = "symlink"
	LinkModeCopy     = "copy"
)

var validLinkModes = map[string]struct{}{
	LinkModeJunction: {},
	LinkModeSymlink:  {},
	LinkModeCopy:     {},
}

// ValidateLinkMode returns an error if mode is not supported.
func ValidateLinkMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if _, ok := validLinkModes[mode]; !ok {
		return fmt.Errorf("unsupported link mode %q (use junction, symlink, or copy)", mode)
	}
	return nil
}

// SetLinkMode updates defaults.link_mode and persists config.toml.
func SetLinkMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if err := ValidateLinkMode(mode); err != nil {
		return err
	}
	viper.Set("defaults.link_mode", mode)
	return Save()
}

// Save writes the current Viper state to %PVM_HOME%\config.toml.
func Save() error {
	home, err := paths.Home()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	path := paths.ConfigFile(home)
	if used := viper.ConfigFileUsed(); used != "" {
		return viper.WriteConfig()
	}
	return viper.SafeWriteConfigAs(path)
}
