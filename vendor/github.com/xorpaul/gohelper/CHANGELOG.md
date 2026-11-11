# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.0.2] - 2026-11-11

### Fixed
- `ExecuteCommand` returned `1` for every non-zero exit code, so callers comparing against another exit code (e.g. `goahead_client`'s `restart_condition_script_exit_code_for_reboot: 2`) never matched. The real exit code is now kept in `ReturnCode` (`-1` if the command was killed by a signal); a command that can not be started still returns `1`
- `ExecuteCommand` ignored its `timeout` argument. A `timeout` larger than `0` now kills the command after that many seconds and returns `ReturnCode` `124` (like `timeout(1)`), `0` means no timeout

### Changed
- Updated dependencies and vendor (`go.mod`, `go.sum`, `vendor/`); `go test ./...` works again
- Releases are made with the `go-build-release` git submodule (https://github.com/xorpaul/go-build-release); `.build.cfg` disables all platform builds because this is a library without a binary

### Added
- Tests for `ExecuteCommand`: exit codes, timeout and a command that can not be started

## [v0.0.1] - 2025-07-22

### Added
- `Colorize()` (exported) for colored output
- `DetectBrowserOrCLI`

### Changed
- Updated vendor dependencies

The history before the first tag (2018 - 2023) added, among others: `FatalExit`, `StringInSlice`, `ParseNetworks`, `getRequestClientIp`, `BoolPointer`, `InBetween` and exiting on `Fatal` if `WarnExit` is set.
