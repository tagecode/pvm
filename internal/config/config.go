package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"tagecode/pvm/pkg/paths"
)

const envPrefix = "PVM"

// Load initializes Viper with defaults and reads config.toml from PVM_HOME.
func Load() (home string, err error) {
	home, err = paths.Home()
	if err != nil {
		return "", err
	}

	v := viper.GetViper()
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(home)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return "", fmt.Errorf("read config: %w", err)
		}
	}

	return home, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("defaults.arch", "auto")
	v.SetDefault("defaults.thread_safe", false)
	v.SetDefault("defaults.link_mode", "junction")
	v.SetDefault("defaults.auto_use_after_install", true)

	v.SetDefault("mirror.url", "https://downloads.php.net/~windows/releases/")
	v.SetDefault("mirror.preset", "official")

	v.SetDefault("proxy.enabled", false)
	v.SetDefault("proxy.http", "")
	v.SetDefault("proxy.https", "")

	v.SetDefault("download.cache_dir", "")
	v.SetDefault("download.concurrency", 4)
	v.SetDefault("download.retry", 3)

	v.SetDefault("security.verify_sha256", true)
	v.SetDefault("security.verify_gpg", false)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.file", "")

	v.SetDefault("project.auto_switch", false)
}

func GetString(key string) string { return viper.GetString(key) }
func GetBool(key string) bool     { return viper.GetBool(key) }
func GetInt(key string) int       { return viper.GetInt(key) }

func MirrorURL() string {
	return viper.GetString("mirror.url")
}

func LinkMode() string {
	return viper.GetString("defaults.link_mode")
}

func VerifySHA256() bool {
	return viper.GetBool("security.verify_sha256")
}

func CacheDir(home string) string {
	if d := viper.GetString("download.cache_dir"); d != "" {
		return d
	}
	return paths.CacheDir(home)
}
