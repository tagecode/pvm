package php

import (
	"fmt"
	"runtime"
	"strings"
)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Revision   int    // Windows rebuild suffix, e.g. 7.2.33-1
	Prerelease string // e.g. "dev" for 8.5.0-dev
	Arch       string
	ThreadSafe bool
	VC         string
}

func (v Version) versionLabel() string {
	label := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Revision > 0 {
		label += fmt.Sprintf("-%d", v.Revision)
	} else if v.Prerelease != "" {
		label += "-" + v.Prerelease
	}
	return label
}

func (v Version) ID() string {
	ts := "nts"
	if v.ThreadSafe {
		ts = "ts"
	}
	return fmt.Sprintf("%s-%s-%s-%s", v.versionLabel(), v.Arch, ts, v.VC)
}

func (v Version) String() string {
	ts := "NTS"
	if v.ThreadSafe {
		ts = "TS"
	}
	return fmt.Sprintf("%s (%s %s %s)", v.versionLabel(), strings.ToUpper(v.Arch), ts, v.VC)
}

// ParseID parses directory id like "8.3.10-x64-nts-vs16" or "7.2.33-1-x64-nts-vc15".
func ParseID(id string) (Version, error) {
	parts := strings.Split(id, "-")
	if len(parts) < 4 {
		return Version{}, fmt.Errorf("invalid version id %q", id)
	}

	vc := parts[len(parts)-1]
	tsToken := parts[len(parts)-2]
	arch := parts[len(parts)-3]
	verParts := parts[:len(parts)-3]

	threadSafe := tsToken == "ts"
	if tsToken != "ts" && tsToken != "nts" {
		return Version{}, fmt.Errorf("invalid thread token in %q", id)
	}

	major, minor, patch, revision, prerelease, err := parseVersionParts(verParts)
	if err != nil {
		return Version{}, fmt.Errorf("parse version segment in %q: %w", id, err)
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Revision:   revision,
		Prerelease: prerelease,
		Arch:       arch,
		ThreadSafe: threadSafe,
		VC:         vc,
	}, nil
}

func parseVersionParts(parts []string) (major, minor, patch, revision int, prerelease string, err error) {
	if len(parts) == 0 {
		return 0, 0, 0, 0, "", fmt.Errorf("empty version")
	}
	if _, err = fmt.Sscanf(parts[0], "%d.%d.%d", &major, &minor, &patch); err != nil {
		return 0, 0, 0, 0, "", err
	}
	if len(parts) == 1 {
		return major, minor, patch, 0, "", nil
	}
	switch parts[1] {
	case "dev":
		return major, minor, patch, 0, "dev", nil
	default:
		if _, err = fmt.Sscanf(parts[1], "%d", &revision); err != nil {
			return 0, 0, 0, 0, "", fmt.Errorf("invalid suffix %q", parts[1])
		}
		return major, minor, patch, revision, "", nil
	}
}

func DefaultArch() string {
	if runtime.GOARCH == "386" {
		return "x86"
	}
	return "x64"
}

func DefaultVC(major, minor int) string {
	switch {
	case major < 5 || (major == 5 && minor <= 2):
		return "vc6"
	case major == 5 && minor == 3:
		return "vc9"
	case major == 5:
		return "vc11"
	case major == 7 && minor <= 1:
		return "vc14"
	case major == 7:
		return "vc15"
	case major == 8 && minor <= 3:
		return "vs16"
	case major >= 8:
		return "vs17"
	default:
		return "vs16"
	}
}

func ApplyDefaults(v Version) Version {
	if v.Arch == "" {
		v.Arch = DefaultArch()
	}
	if v.VC == "" {
		v.VC = DefaultVC(v.Major, v.Minor)
	}
	return v
}
