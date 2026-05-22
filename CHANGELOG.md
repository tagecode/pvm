# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-05-22

### Added

- MVP CLI: `install`, `uninstall`, `use`, `list`, `list-remote`, `current`, `which`, `info`
- Alias and default version: `alias`, `default`
- Environment setup: `setup`, `unsetup`, `refresh`, `path`
- Configuration: `link-mode`, global flags (`--json`, `--dry-run`, `-y`)
- Offline install via `--from-zip`
- Windows junction-based version switching
- SHA256 verification for official PHP downloads
- **UX:** first `use` auto-setup; session PATH refresh after `use`/`install`; MSI/ZIP preconfigure `current` PATH; default `auto_use_after_install = true`
- GitHub Actions CI and release pipeline (exe / zip / msi for x64 and x86)

### Known limitations

- Verbose logging (`-v`) and HTTP proxy config not yet wired
- No `config` / `ext` / `.pvmrc` commands (planned for v1.0)

[0.1.0]: https://github.com/tagecode/pvm/releases/tag/v0.1.0
