//go:build !windows

package win

import "fmt"

func GetUserEnv(string) (string, error) {
	return "", fmt.Errorf("environment operations are only supported on Windows")
}

func DeleteUserEnv(string) error {
	return fmt.Errorf("environment operations are only supported on Windows")
}

func SetUserEnv(string, string) error {
	return fmt.Errorf("environment operations are only supported on Windows")
}

func PrependUserPath(string) error {
	return fmt.Errorf("environment operations are only supported on Windows")
}

func RemoveUserPathEntry(string) error {
	return fmt.Errorf("environment operations are only supported on Windows")
}

func GetEnvStatus(string, string) (*EnvStatus, error) {
	return nil, fmt.Errorf("environment operations are only supported on Windows")
}

func NeedsUserEnvironmentSetup(string) (bool, error) {
	return false, fmt.Errorf("environment operations are only supported on Windows")
}

func RefreshSessionEnv(string) {}
