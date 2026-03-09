# tailer

`tailer` is a CLI for watching a directory of log files by pattern.

It behaves like `tail -f`, but it also keeps rescanning the directory and starts following new matching files as they appear. This is useful for log directories where new files are dropped over time.

## Install

Build and install from this repository:

```bash
go install ./cmd/tailer
```

The binary will be installed to `GOBIN` if it is set. Otherwise Go installs it to `$(go env GOPATH)/bin`.

For example:

```bash
$(go env GOPATH)/bin/tailer
```

## Usage

```bash
tailer --dir <directory> --pattern <glob>
```

Example:

```bash
tailer --dir /var/log/myapp --pattern "*.log"
```

This will:

- Find existing files in `/var/log/myapp` matching `*.log`
- Start following appended lines in those files
- Detect new `*.log` files in that directory
- Start following those new files automatically
- Prefix each output line with the source file path

## Flags

```text
-dir string
    Directory to watch for matching files. (default ".")
-from-start
    Start existing files from byte 0 instead of the end.
-pattern string
    Glob pattern for files inside the watched directory. (default "*.log")
-poll-interval duration
    How often to rescan the directory for new files. (default 1s)
```

## Output Format

Each emitted line is prefixed with the full path of the source file:

```text
/var/log/myapp/app.log: request complete
/var/log/myapp/worker.log: job queued
```

## Common Examples

Tail all log files in the current directory:

```bash
tailer --pattern "*.log"
```

Read existing files from the beginning instead of starting at EOF:

```bash
tailer --dir ./logs --pattern "*.log" --from-start
```

Use faster directory rescans for rapidly created files:

```bash
tailer --dir ./logs --pattern "*.log" --poll-interval 250ms
```

Watch rotated or recreated files:

```bash
tailer --dir ./logs --pattern "app-*.log"
```

`tailer` uses [`github.com/nxadm/tail`](https://pkg.go.dev/github.com/nxadm/tail) for low-level tail-follow behavior, including reopen support.

The CLI itself is built with Cobra.

## Current Scope

The current implementation:

- Watches one directory, not recursive subdirectories
- Uses glob matching against filenames in that directory
- Starts existing files at EOF by default
- Starts newly discovered files from the beginning
- Stops cleanly on `Ctrl-C`

## Development

Run tests:

```bash
go test ./...
```

Build the binary:

```bash
go build ./...
```
