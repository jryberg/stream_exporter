# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Prometheus exporter that reads lines from a streaming input (file, socket, syslog, stdin, named pipe) and turns regex matches into metrics. Maintained fork of the archived `carlpett/stream_exporter`; module path is `github.com/jryberg/stream_exporter`.

## Commands

```sh
make build                     # go build with version ldflags (VERSION from git describe --tags)
make test                      # go test -v -cover ./...
go test ./linemetrics -run TestUnmarshalHistogramConfig   # single test
go test ./input -bench .       # syslog message handler benchmark
go vet ./...                   # CI runs this
golangci-lint run ./...        # default linters, no config file; CI runs this
GOOS=windows golangci-lint run ./...   # needed to lint service_windows.go and logging_windows.go
make crossbuild                # CGO_ENABLED=0 binaries for linux/windows/darwin into binaries/
make docker
```

CI (`.github/workflows/ci.yaml`) runs vet, `make test`, `make crossbuild`, and golangci-lint v2.15.0 for both GOOS=linux and GOOS=windows. Crossbuild matters because `service_windows.go` and `input/namedpipe_linux.go` only compile on their OS, so a change can build locally and still break another platform. Go version comes from `go.mod`; the version is pinned in `go.mod`, `.promu.yml` and the `Dockerfile` (digest-pinned image), so bump all three together.

Try a config end to end without a server: `./stream_exporter -config c.yaml -input.type file -input.file.path some.log -input.file.mode dryrun` processes the file, prints the metrics output and exits.

## Architecture

`exporter.go` (package `main`) parses flags, builds one `LineMetric` per configured pattern, starts the chosen input in a goroutine, and loops reading lines from a `chan string`, calling `MatchLine` on every metric for every line. The loop ends when the input closes the channel or a quit signal arrives. Inputs that finish (file `dryrun`, stdin with `-input.stdin.quit-on-eof`) close the channel to make the process exit. The channel is unbuffered, so a completed send only means the main loop received the line, not that the metrics are updated. An input that reads the metrics itself (`writeMetrics` at EOF) must call `waitForProcessing()` first, which sends on `input.LinesProcessed`, a channel the main loop only receives from between lines.

**Self-registration via `init()`.** Most extension points are global registries filled from `init()` functions, so adding a file is enough to wire it in:
- `input/`: each input file calls `registerInput(name, factory)` and declares its own package-level `flag` vars prefixed `input.<name>.`. The factory validates flags and returns a `StreamInput` (`StartStream(ch chan<- string)`). Inputs treat runtime errors as fatal (`log.Fatal`).
- `main`: OS-specific files add entries to `startupTasks`, which run after `flag.Parse()`. `service_windows.go` uses this to install, uninstall or run as a Windows service, and stops the process by sending on `quitSig`.
- Platform restriction uses filename suffixes (`_linux.go`, `_windows.go`) where a single GOOS fits, and a `//go:build` tag only where it doesn't (`logging_syslog.go` is `unix`, i.e. everything but Windows).

**`linemetrics/`.** `ReadPatternConfig` unmarshals YAML (`gopkg.in/yaml.v2`) into `MetricsConfig`; `HistogramConfig` and `SummaryConfig` are embedded `inline`, so their keys sit at the same level as `name`/`kind`/`pattern`. `NewLineMetric` compiles the regex and uses its named capture groups as label names. Each kind has a plain and a `*Vec` implementation, picked by whether the pattern has capture groups. For gauge, histogram and summary, the capture group named `value` is required: it is parsed as a float and removed from the labels (`getValueCaptureIndex` / `getLabelsAndValue`). Metric help text is the metric name.

`input/util.go:writeMetrics` dumps the default Prometheus registry as text format. Dryrun and stdin `write-on-eof` use it.

The README documents every flag and config option. Update it when you add or change one.
