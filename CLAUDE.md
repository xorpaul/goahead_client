# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Purpose

`goahead_client` is the client for the [goahead](https://github.com/xorpaul/goahead) restart management service. It runs on each managed host (cron/timer): it checks a local `restart_condition_script`, asks the goahead service whether the host may restart now, and if the service answers with `go_ahead`, runs the restart hooks in `os_restart_hooks_dir` (which finally reboot the host). See `README.md` for the workflow and the config file.

## Layout

- `client.go` — `main()`, `doMain()` (local restart condition, then inquire), `doRestart()`, `executeRestartHooks()`, HTTP client setup, request/response handling
- `config.go` — `configSettings` and `readConfigfile()` (YAML, default `/etc/goahead/client.yml`)
- `payload.go` — request payload (fqdn, uptime, restart reason)
- `client_test.go` — tests with an `httptest` fake goahead server; most tests re-run the test binary as a subprocess (`TEST_FOR_CRASH_<name>=1`) because the code calls `os.Exit` via `gohelper`
- `tests/` — fake server responses (`*.json`), test scripts (`always-true.sh`, `always-false.sh`, `exit-2.sh`) and hook directories
- `example/` — example restart condition script and hooks
- `vendor/` — vendored dependencies, including `github.com/xorpaul/gohelper`
- `go-build-release/` — git submodule (https://github.com/xorpaul/go-build-release) with the release scripts; `./go-build-release/build_release.sh <version>` runs the tests, cross-compiles (injecting `main.buildversion` and `main.buildtime`), tags, pushes and creates the GitHub release with the `CHANGELOG.md` section as text. Read its README first. After a fresh clone: `git submodule update --init`

## Conventions

- Public repository on GitHub: never put internal hostnames, IPs, URLs or other internal service names into code, tests, docs, examples or commit messages
- Command execution goes through `gohelper`'s `ExecuteCommand(command, timeout, allowFail)`. The exit code of `restart_condition_script` is compared with `restart_condition_script_exit_code_for_reboot`, so exit codes other than 0 and 1 only work with a `gohelper` that keeps the real exit code (see `CHANGELOG.md`). The restart condition script has a 5s timeout, restart hooks use `os_restart_hooks_timeout` (default 0 = no limit)
- A change of `gohelper` behavior needs `go get github.com/xorpaul/gohelper@<version> && go mod vendor` here
- Keep `CHANGELOG.md` current (`## [Unreleased]`)

## Testing

```
go test -v ./...
```

`go test -v ./...` is also the gate of the release script. `TestRestartConditionScriptExitCode2` needs `gohelper` >= v0.0.2, which keeps the real exit code of commands.

## Git commits

Do not run `git commit`; write the commands to `commit.sh` at the repo root (see the global instructions).
