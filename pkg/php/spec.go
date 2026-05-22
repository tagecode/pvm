package php

import (
	"fmt"
	"strconv"
	"strings"
)

type SpecKind int

const (
	SpecExact SpecKind = iota
	SpecMinor
	SpecLatest
	SpecAlias
)

type VersionSpec struct {
	Kind  SpecKind
	Major int
	Minor int
	Patch int
	Alias string
	Raw   string
}

// ParseSpec parses user input like "8.3.10", "8.3", "latest", or alias name.
func ParseSpec(raw string) (VersionSpec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return VersionSpec{}, fmt.Errorf("empty version spec")
	}

	switch strings.ToLower(raw) {
	case "latest":
		return VersionSpec{Kind: SpecLatest, Raw: raw}, nil
	}

	parts := strings.Split(raw, ".")
	if len(parts) == 1 {
		if _, err := strconv.Atoi(parts[0]); err != nil {
			return VersionSpec{Kind: SpecAlias, Alias: raw, Raw: raw}, nil
		}
	}

	if len(parts) < 2 || len(parts) > 3 {
		return VersionSpec{}, fmt.Errorf("invalid version spec %q", raw)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return VersionSpec{Kind: SpecAlias, Alias: raw, Raw: raw}, nil
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return VersionSpec{}, fmt.Errorf("invalid minor in %q", raw)
	}

	spec := VersionSpec{Major: major, Minor: minor, Raw: raw}
	if len(parts) == 2 {
		spec.Kind = SpecMinor
		return spec, nil
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return VersionSpec{}, fmt.Errorf("invalid patch in %q", raw)
	}
	spec.Patch = patch
	spec.Kind = SpecExact
	return spec, nil
}
