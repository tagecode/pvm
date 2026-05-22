package php

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	zipHrefPattern = regexp.MustCompile(`(?i)href="(php-[^"]+\.zip)"`)
	parseZipNameRE = regexp.MustCompile(`(?i)^php-(\d+)\.(\d+)\.(\d+)((?:-\d+|-dev)?)(-nts)?-Win32-(vc\d+|vs\d+)-(x64|x86)\.zip$`)
)

const (
	OfficialReleasesURL = "https://downloads.php.net/~windows/releases/"
	OfficialArchivesURL = "https://downloads.php.net/~windows/releases/archives/"
)

// ZipName returns the official Windows PHP zip filename used on windows.php.net.
// Examples:
//
//	php-8.3.31-Win32-vs16-x64.zip          (TS)
//	php-8.3.31-nts-Win32-vs16-x64.zip      (NTS)
//	php-7.2.33-1-nts-Win32-vc15-x64.zip    (Windows rebuild)
func (v Version) ZipName() string {
	v = ApplyDefaults(v)
	prefix := fmt.Sprintf("php-%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Revision > 0 {
		prefix += fmt.Sprintf("-%d", v.Revision)
	} else if v.Prerelease != "" {
		prefix += "-" + v.Prerelease
	}
	if !v.ThreadSafe {
		prefix += "-nts"
	}
	arch := v.Arch
	if arch == "" {
		arch = DefaultArch()
	}
	return fmt.Sprintf("%s-Win32-%s-%s.zip", prefix, v.VC, arch)
}

func (v Version) DownloadURL(base string) string {
	return base + v.ZipName()
}

// ExtractZipNamesFromHTML returns php-*.zip href targets from an Apache index page.
func ExtractZipNamesFromHTML(html string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range zipHrefPattern.FindAllStringSubmatch(html, -1) {
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// IsBinaryZipName reports whether name looks like an installable PHP Windows binary zip.
func IsBinaryZipName(name string) bool {
	lower := strings.ToLower(name)
	if !strings.HasPrefix(lower, "php-") || !strings.Contains(lower, "-win32-") {
		return false
	}
	if strings.HasPrefix(lower, "php-debug-pack-") ||
		strings.HasPrefix(lower, "php-test-pack-") ||
		strings.HasSuffix(lower, "-src.zip") {
		return false
	}
	_, err := ParseZipName(name)
	return err == nil
}

// ParseZipName parses official Windows PHP zip filenames from releases/archives pages.
func ParseZipName(name string) (Version, error) {
	m := parseZipNameRE.FindStringSubmatch(name)
	if m == nil {
		return Version{}, fmt.Errorf("unrecognized zip name %q", name)
	}
	var major, minor, patch int
	fmt.Sscanf(m[1]+"."+m[2]+"."+m[3], "%d.%d.%d", &major, &minor, &patch)

	threadSafe := m[5] == ""
	revision := 0
	prerelease := ""
	switch suffix := m[4]; suffix {
	case "":
	case "-dev":
		prerelease = "dev"
	default:
		fmt.Sscanf(suffix, "-%d", &revision)
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Revision:   revision,
		Prerelease: prerelease,
		Arch:       strings.ToLower(m[7]),
		ThreadSafe: threadSafe,
		VC:         strings.ToLower(m[6]),
	}, nil
}
