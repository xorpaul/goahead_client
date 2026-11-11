# gohelper

Small collection of helper functions shared by the Go tools of this GitHub account (e.g. [goahead](https://github.com/xorpaul/goahead), [goahead_client](https://github.com/xorpaul/goahead_client), [g10k](https://github.com/xorpaul/g10k)). Package name `helper`, usually imported as `h`:

```go
import h "github.com/xorpaul/gohelper"
```

## What is in there

- **Logging**: `Debugf`, `Verbosef`, `Infof`, `Warnf`, `Fatalf`, switched with the package variables `Debug`, `Verbose`, `Info`, `InfoTimestamp`, `Quiet`. `Warnf` and `Fatalf` exit the program if `WarnExit` / `FatalExit` is set
- **Files and directories**: `FileExists`, `IsDir`, `NormalizeDir`, `CheckDirAndCreate`, `CreateOrPurgeDir`, `PurgeDir`, `GetSha256sumFile`, `WriteStructJSONFile`
- **Commands**: `ExecuteCommand(command, timeout, allowFail)`, see below
- **HTTP**: `GetRequestClientIp` (honors `X-Forwarded-For` from trusted proxy networks), `ParseNetworks`, `DetectBrowserOrCLI`
- **Misc**: `StringSliceContains`, `KeysString`, `InBetween`, `BoolPointer`, `RandSeq`, `FuncName`, `TimeTrack`, `Colorize`

## ExecuteCommand

```go
er := h.ExecuteCommand("/usr/lib/nagios/plugins/check_tcp -H localhost -p 22", 5, true)
fmt.Println(er.ReturnCode, er.Output)
```

The command is split with shell quoting rules and run without a shell, stdout and stderr are returned together in `Output`.

- `ReturnCode` is the real exit code of the command (`-1` if it was killed by a signal)
- `timeout` is in seconds: the command is killed after that and `ReturnCode` is `124` (like `timeout(1)`); `0` or less means no timeout
- a command that can not be started at all (e.g. it does not exist) returns `ReturnCode` `1`
- with `allowFail` set to `false` a failing command is fatal (`Fatalf`)

Versions before v0.0.2 replaced every non-zero exit code with `1` and ignored `timeout`, see the [changelog](CHANGELOG.md).

## Development

```
go test -v ./...
```

See [CHANGELOG.md](CHANGELOG.md) for the changes of each version.
