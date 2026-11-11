# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.0.4] - 2026-11-11

### Fixed
- A `restart_condition_script` exit code other than `0` or `1` could never match `restart_condition_script_exit_code_for_reboot`, because the vendored `gohelper` turned every non-zero exit code into `1`; the restart hooks also had no timeout. Fixed in `gohelper` v0.0.2, which this version now uses
- The tests discarded the HTTP client created in `TestMain`, so every test talking to the fake goahead server crashed with a nil pointer in its subprocess
- `go vet` (run by `go test`) failed on the tests because of non-constant format strings, and `TestUptimeLow` expected a payload without `restart_reason`; both fixed so `go test ./...` can be the release gate

### Changed
- Updated `gohelper` to v0.0.2 and the other dependencies and vendor (`go` 1.26.0)
- Releases are built with the `go-build-release` git submodule (https://github.com/xorpaul/go-build-release) instead of the copied `build_release.sh`, which was removed
- `-version` prints the version injected at build time (`-X main.buildversion=...`), or `dev` for a plain `go build`, instead of a hard-coded `0.0.3`

### Added
- Test for a restart condition script that exits with a code other than `0` and `1` (`tests/exit-2.sh`)

## [v0.0.3] - 2025-04-17

### Changed
- Updated vendor dependencies

## [v0.0.2] - 2023-05-30

### Added
- `/etc/goahead/disabled`: if the file exists the goahead run is skipped and its content is logged as the reason
- Go module support (`go.mod`)

### Changed
- Use `gohelper` (`h`) for the helper functions

## [v0.0.1] - 2020-02-05

### Added
- Client that asks the goahead service whether the system may restart: restart condition script, restart hooks directory, workflow examples
- `LICENSE`, `.gitignore` and `build_release.sh`
