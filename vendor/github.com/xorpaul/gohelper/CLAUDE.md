# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Purpose

`github.com/xorpaul/gohelper` is a tiny library (single package `helper`, file `helper.go`) used by several other Go tools (goahead, goahead_client, g10k, gorpe, pkgproxy, ...). Every exported function is public API of all of them: do not rename or change behavior lightly, and check the call sites of the consumers (they vendor this module) before changing semantics.

## Layout

- `helper.go` — all helper functions and the package variables (`Debug`, `Verbose`, `Info`, `InfoTimestamp`, `WarnExit`, `FatalExit`, `Quiet`)
- `helper_test.go` — tests (package `helper`)
- `CHANGELOG.md` — Keep a Changelog format, keep `## [Unreleased]` current

## Conventions

- Public repository on GitHub: never put internal hostnames, IPs, URLs or other internal service names into code, tests, docs or commit messages
- Keep the module dependency-light; consumers vendor it
- Log through the package's own `Debugf`/`Infof`/`Warnf`/`Fatalf`, not `log` directly
- `ExecuteCommand` contract: real exit code in `ReturnCode`, `timeout` in seconds (`124` on timeout, `0` = no timeout), `1` if the command could not be started; `allowFail=false` makes failures fatal
- Add or update a test for every behavior change and a `CHANGELOG.md` entry

## Testing

```
go test -v ./...
```

Dependencies are vendored (`vendor/`); update them with `go get -u ./... && go mod tidy && go mod vendor`.

## Releasing

Releases are made with the `go-build-release` git submodule (`git submodule update --init`; read `go-build-release/README.md` first): rename `[Unreleased]` in `CHANGELOG.md` to the version (latest is `v0.0.2`), commit, then `./go-build-release/build_release.sh <version>` (runs the tests, tags, pushes the tag, creates the GitHub release with the changelog section). This is a library without a `main` package, so `.build.cfg` switches off all platform builds (they would only produce useless archives); keep it. Then bump the dependency in the consumers (`go get github.com/xorpaul/gohelper@vX.Y.Z && go mod vendor`).

## Git commits

Do not run `git commit`; write the commands to `commit.sh` at the repo root (see the global instructions).
