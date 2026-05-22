package config

import "tagecode/pvm/pkg/php"

// ResolveArch returns install architecture from CLI flag or config default.
func ResolveArch(flag string) string {
	if flag != "" && flag != "auto" {
		return flag
	}
	cfg := GetString("defaults.arch")
	if cfg == "" || cfg == "auto" {
		return php.DefaultArch()
	}
	return cfg
}
