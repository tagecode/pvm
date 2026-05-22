package php

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tagecode/pvm/pkg/download"
)

type RemoteRelease struct {
	Version      Version
	URL          string
	ZipFile      string
	Checksum     string
	ChecksumAlgo string
	Archived     bool

	// SHA256 is deprecated; use Checksum with ChecksumAlgo.
	SHA256 string
}

type ResolveOptions struct {
	Arch       string
	ThreadSafe *bool
	VC         string
}

type RemoteIndex struct {
	BaseURL    string
	HTTP       *http.Client
	ArchivesURL string
}

func NewRemoteIndex(baseURL string) *RemoteIndex {
	base := strings.TrimSuffix(baseURL, "/") + "/"
	return &RemoteIndex{
		BaseURL:     base,
		ArchivesURL: strings.Replace(base, "/releases/", "/releases/archives/", 1),
		HTTP:        &http.Client{Timeout: 2 * time.Minute},
	}
}

func (r *RemoteIndex) List(ctx context.Context, major, minor *int) ([]RemoteRelease, error) {
	sha256sums, err := r.fetchChecksumFile(ctx, r.BaseURL+"sha256sum.txt")
	if err != nil {
		return nil, fmt.Errorf("fetch sha256sum.txt: %w", err)
	}

	releases, err := r.fetchPage(ctx, r.BaseURL, sha256sums, download.AlgoSHA256, false)
	if err != nil {
		return nil, err
	}

	sha1sums, err := r.fetchChecksumFile(ctx, r.ArchivesURL+"sha1sum.txt")
	if err != nil {
		sha1sums = map[string]string{}
	}
	archived, err := r.fetchPage(ctx, r.ArchivesURL, sha1sums, download.AlgoSHA1, true)
	if err != nil {
		archived = nil
	}

	releases = mergeReleases(releases, archived)
	return filterReleases(releases, major, minor), nil
}

func (r *RemoteIndex) Resolve(ctx context.Context, spec VersionSpec, opts ResolveOptions) (RemoteRelease, error) {
	list, err := r.List(ctx, intPtr(spec.Major), intPtrIf(spec.Kind == SpecMinor || spec.Kind == SpecExact, spec.Minor))
	if err != nil {
		return RemoteRelease{}, err
	}
	if len(list) == 0 {
		return RemoteRelease{}, fmt.Errorf("no releases found for %q", spec.Raw)
	}

	threadSafe := false
	if opts.ThreadSafe != nil {
		threadSafe = *opts.ThreadSafe
	}
	arch := opts.Arch
	if arch == "" || arch == "auto" {
		arch = DefaultArch()
	}
	vc := opts.VC
	if vc == "" {
		vc = DefaultVC(spec.Major, spec.Minor)
	}

	switch spec.Kind {
	case SpecLatest:
		return list[0], nil
	case SpecMinor:
		for _, rel := range list {
			v := ApplyDefaults(rel.Version)
			if v.Major == spec.Major && v.Minor == spec.Minor && v.Arch == arch && v.ThreadSafe == threadSafe && v.VC == vc {
				return rel, nil
			}
		}
	case SpecExact:
		for _, rel := range list {
			v := ApplyDefaults(rel.Version)
			if v.Major == spec.Major && v.Minor == spec.Minor && v.Patch == spec.Patch && v.Arch == arch && v.ThreadSafe == threadSafe && v.VC == vc {
				return rel, nil
			}
		}
	}

	return RemoteRelease{}, fmt.Errorf("no matching release for %q (%s %s %s)", spec.Raw, arch, tsLabel(threadSafe), vc)
}

func (r *RemoteIndex) fetchChecksumFile(ctx context.Context, url string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return download.ParseChecksumSum(string(body))
}

func (r *RemoteIndex) fetchPage(ctx context.Context, base string, checksums map[string]string, algo string, archived bool) ([]RemoteRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch releases: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out []RemoteRelease
	for _, name := range ExtractZipNamesFromHTML(string(body)) {
		if !IsBinaryZipName(name) {
			continue
		}
		v, err := ParseZipName(name)
		if err != nil {
			continue
		}
		sum := checksums[name]
		out = append(out, RemoteRelease{
			Version:      v,
			URL:          base + name,
			ZipFile:      name,
			Checksum:     sum,
			ChecksumAlgo: algo,
			Archived:     archived,
			SHA256:       sumIfSHA256(algo, sum),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Version, out[j].Version
		if a.Major != b.Major {
			return a.Major > b.Major
		}
		if a.Minor != b.Minor {
			return a.Minor > b.Minor
		}
		return a.Patch > b.Patch
	})
	return out, nil
}

func sumIfSHA256(algo, sum string) string {
	if algo == download.AlgoSHA256 {
		return sum
	}
	return ""
}

func mergeReleases(a, b []RemoteRelease) []RemoteRelease {
	seen := map[string]RemoteRelease{}
	for _, r := range a {
		seen[r.Version.ID()] = r
	}
	for _, r := range b {
		if _, ok := seen[r.Version.ID()]; !ok {
			seen[r.Version.ID()] = r
		}
	}
	out := make([]RemoteRelease, 0, len(seen))
	for _, r := range seen {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		x, y := out[i].Version, out[j].Version
		if x.Major != y.Major {
			return x.Major > y.Major
		}
		if x.Minor != y.Minor {
			return x.Minor > y.Minor
		}
		return x.Patch > y.Patch
	})
	return out
}

func filterReleases(list []RemoteRelease, major, minor *int) []RemoteRelease {
	var out []RemoteRelease
	for _, r := range list {
		if major != nil && r.Version.Major != *major {
			continue
		}
		if minor != nil && r.Version.Minor != *minor {
			continue
		}
		out = append(out, r)
	}
	return out
}

func intPtr(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

func intPtrIf(cond bool, v int) *int {
	if !cond {
		return nil
	}
	return &v
}

func tsLabel(ts bool) string {
	if ts {
		return "ts"
	}
	return "nts"
}

func ExtractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	rootPrefix := detectSingleRootPrefix(r.File)
	for _, f := range r.File {
		name := f.Name
		if rootPrefix != "" {
			name = strings.TrimPrefix(name, rootPrefix)
			if name == "" {
				continue
			}
		}
		target := filepath.Join(dest, name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path in zip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func EnsureIni(root string) error {
	ini := filepath.Join(root, "php.ini")
	if _, err := os.Stat(ini); err == nil {
		return nil
	}
	source, err := iniTemplatePath(root)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(source), err)
	}
	return os.WriteFile(ini, data, 0o644)
}

func iniTemplatePath(root string) (string, error) {
	dev := filepath.Join(root, "php.ini-development")
	if _, err := os.Stat(dev); err == nil {
		return dev, nil
	}
	prod := filepath.Join(root, "php.ini-production")
	if _, err := os.Stat(prod); err == nil {
		return prod, nil
	}
	return "", fmt.Errorf("no php.ini template in %s (need php.ini-development or php.ini-production)", root)
}

func detectSingleRootPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}
	root := ""
	for _, f := range files {
		parts := strings.SplitN(filepath.ToSlash(f.Name), "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			return ""
		}
		if root == "" {
			root = parts[0] + "/"
			continue
		}
		if parts[0]+"/" != root {
			return ""
		}
	}
	return root
}
