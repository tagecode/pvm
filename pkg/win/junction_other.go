//go:build !windows

package win

import "fmt"

type LinkMode string

const (
	LinkModeJunction LinkMode = "junction"
	LinkModeSymlink  LinkMode = "symlink"
	LinkModeCopy     LinkMode = "copy"
)

func CreateLink(link, target string, mode LinkMode) error {
	return fmt.Errorf("link operations are only supported on Windows")
}

func RemoveLink(link string) error {
	return fmt.Errorf("link operations are only supported on Windows")
}

func ReadLinkTarget(link string) (string, error) {
	return "", fmt.Errorf("link operations are only supported on Windows")
}
