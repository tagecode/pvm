package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tagecode/pvm/internal/config"
	"tagecode/pvm/pkg/download"
	"tagecode/pvm/pkg/paths"
	"tagecode/pvm/pkg/php"
)

type InstallOptions struct {
	Arch       string
	ThreadSafe *bool
	VC         string
	FromZip    string
	Reinstall  bool
	DryRun     bool
}

type Installer struct {
	Home       string
	Remote     *php.RemoteIndex
	Downloader *download.Client
}

func NewInstaller(home string) *Installer {
	return &Installer{
		Home:       home,
		Remote:     php.NewRemoteIndex(config.MirrorURL()),
		Downloader: download.NewClient(),
	}
}

func (i *Installer) Install(ctx context.Context, spec php.VersionSpec, opts InstallOptions) (*InstalledVersion, error) {
	if err := paths.EnsureLayout(i.Home); err != nil {
		return nil, err
	}

	resolveOpts := php.ResolveOptions{
		Arch:       config.ResolveArch(opts.Arch),
		ThreadSafe: opts.ThreadSafe,
		VC:         opts.VC,
	}

	release, err := i.resolveRelease(ctx, spec, resolveOpts, opts)
	if err != nil {
		return nil, err
	}

	return i.InstallRelease(ctx, release, opts)
}

func (i *Installer) resolveRelease(ctx context.Context, spec php.VersionSpec, resolveOpts php.ResolveOptions, opts InstallOptions) (php.RemoteRelease, error) {
	if opts.FromZip != "" {
		return i.releaseFromLocal(spec, opts)
	}
	return i.Remote.Resolve(ctx, spec, resolveOpts)
}

func (i *Installer) releaseFromLocal(spec php.VersionSpec, opts InstallOptions) (php.RemoteRelease, error) {
	name := filepath.Base(opts.FromZip)
	v, err := php.ParseZipName(name)
	if err != nil {
		return php.RemoteRelease{}, UserError("cannot infer version from zip filename", "rename to official format or use pvm install 8.3.31 --from-zip <path>")
	}

	if opts.VC != "" {
		v.VC = opts.VC
	}
	// Prefer arch from zip filename; only override when user passes --arch explicitly.
	if opts.Arch != "" && opts.Arch != "auto" {
		v.Arch = opts.Arch
	}
	if opts.ThreadSafe != nil {
		v.ThreadSafe = *opts.ThreadSafe
	}
	v = php.ApplyDefaults(v)

	switch spec.Kind {
	case php.SpecExact:
		if v.Major != spec.Major || v.Minor != spec.Minor || v.Patch != spec.Patch {
			return php.RemoteRelease{}, UserError("version spec does not match zip file", name)
		}
	case php.SpecMinor:
		if v.Major != spec.Major || v.Minor != spec.Minor {
			return php.RemoteRelease{}, UserError("version spec does not match zip file", name)
		}
	}

	return php.RemoteRelease{
		Version: v,
		ZipFile: name,
	}, nil
}

func (i *Installer) InstallRelease(ctx context.Context, release php.RemoteRelease, opts InstallOptions) (*InstalledVersion, error) {
	id := release.Version.ID()
	dest := paths.VersionDir(i.Home, id)

	if _, err := os.Stat(dest); err == nil && !opts.Reinstall {
		return nil, ErrAlreadyInstalled
	}

	if opts.DryRun {
		return &InstalledVersion{Version: release.Version, ID: id, Path: dest}, nil
	}

	if err := i.extractRelease(ctx, release, dest, opts); err != nil {
		return nil, err
	}

	iv := &InstalledVersion{
		Version: release.Version,
		ID:      id,
		Path:    dest,
	}
	if config.GetBool("defaults.auto_use_after_install") {
		sw := NewSwitcher(i.Home)
		if _, err := sw.Use(ctx, id, UseOptions{}); err != nil {
			return iv, err
		}
		iv.Active = true
	}
	return iv, nil
}

func (i *Installer) extractRelease(ctx context.Context, release php.RemoteRelease, dest string, opts InstallOptions) error {
	cacheDir := config.CacheDir(i.Home)
	zipName := release.ZipFile
	if zipName == "" {
		zipName = release.Version.ZipName()
	}
	zipPath := filepath.Join(cacheDir, zipName)

	if opts.FromZip != "" {
		zipPath = opts.FromZip
	} else {
		if err := i.Downloader.Fetch(ctx, release.URL, zipPath); err != nil {
			return SystemError("download failed", err.Error())
		}
	}

	if config.VerifySHA256() && release.Checksum == "" && opts.FromZip == "" {
		return SystemError(
			"checksum not available for "+zipName,
			"official sha256sum.txt or archives sha1sum.txt may be missing this file",
		)
	}

	if config.VerifySHA256() && release.Checksum != "" {
		algo := release.ChecksumAlgo
		if algo == "" {
			algo = download.AlgoSHA256
		}
		if err := download.VerifyChecksum(zipPath, release.Checksum, algo); err != nil {
			if opts.FromZip == "" {
				os.Remove(zipPath)
			}
			return SystemError("checksum verification failed", err.Error())
		}
	}

	tmp := fmt.Sprintf("%s.tmp.%d", dest, os.Getpid())
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}

	cleanup := func() { os.RemoveAll(tmp) }

	if err := php.ExtractZip(zipPath, tmp); err != nil {
		cleanup()
		return SystemError("extract failed", err.Error())
	}
	if err := php.EnsureIni(tmp); err != nil {
		cleanup()
		return err
	}
	if _, err := os.Stat(filepath.Join(tmp, "php.exe")); err != nil {
		cleanup()
		return SystemError("invalid PHP package", "php.exe not found after extract")
	}

	if err := os.RemoveAll(dest); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		cleanup()
		return err
	}
	return nil
}

func (i *Installer) Uninstall(_ context.Context, target string) error {
	reg := NewRegistry(i.Home)
	id, err := reg.ResolveID(target)
	if err != nil {
		return err
	}

	activeID, _ := reg.activeID()
	if stringsEqualFold(id, activeID) {
		return ErrActiveVersion
	}

	dest := paths.VersionDir(i.Home, id)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return ErrNotInstalled
	}
	return os.RemoveAll(dest)
}

func stringsEqualFold(a, b string) bool {
	return strings.EqualFold(a, b)
}