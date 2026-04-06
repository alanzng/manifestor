# Agent Guide — manifestor

This file helps AI coding agents understand the project structure, conventions, and workflows for `manifestor`.

---

## Project Overview

`manifestor` is a **zero-dependency Go library** for parsing, filtering, transforming, and building HLS (`.m3u8`) and DASH (`.mpd`) streaming manifests. It also ships as an HTTP proxy server and a CLI tool.

**Module:** `github.com/alanzng/manifestor`
**Go version:** 1.22+
**Dependencies:** stdlib only (no external packages in core packages)

---

## Repository Layout

```
manifestor/
├── hls/                  # HLS-specific: types, parser, writer, filter, builder, options
├── dash/                 # DASH-specific: types, parser, writer, filter, builder, options
├── manifest/             # Unified auto-detect API (wraps hls/ and dash/)
├── server/               # HTTP proxy server (/filter and /build endpoints)
├── cmd/manifestor/       # CLI entry point (filter / build / serve subcommands)
├── testdata/
│   ├── hls/              # Real-world HLS fixture manifests (.m3u8)
│   └── dash/             # Real-world DASH fixture manifests (.mpd)
├── .github/workflows/    # CI: test, format, lint, bench
├── REQUIREMENT.md        # Full functional and non-functional requirements
└── AGENTS.md             # This file
```

---

## Essential Commands

```bash
# Build everything
go build ./...

# Run all tests (always use -race)
go test -race ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...

# Check formatting (must pass before commit)
gofmt -l .

# Auto-format all files
gofmt -w .

# Run linter
golangci-lint run

# Run benchmarks
go test -bench=. -benchmem -run='^$' ./...
```

**Always run `gofmt -w .` before committing.** CI has a dedicated format check that will fail on unformatted files.

---

## Architecture & Key Patterns

### Option pattern

Both `hls` and `dash` packages use a functional options pattern for `Filter()`:

```go
type Option func(*filterConfig)

func WithCodec(codec string) Option {
    return func(c *filterConfig) { c.codec = codec }
}
```

The `manifest` package defines a unified `Option` interface that both packages satisfy, so callers can use a single import for the auto-detect API.

### Builder pattern

Builders use a fluent/chaining API and return errors only from `Build()`, not from `Add*` methods:

```go
b := hls.NewMasterBuilder()
b.SetVersion(6).AddVariant(p).AddAudioTrack(a)
result, err := b.Build()
```

### Error handling

All sentinel errors are **package-level vars**, not inline `errors.New()`:

```go
var ErrNoVariantsRemain = errors.New("hls: no variants remain after filtering")
```

This allows callers to use `errors.Is()`. Never replace these with inline errors.

### Thread safety

`Filter()` must be safe for concurrent use — no global or shared mutable state. All state flows through the `filterConfig` struct created per call.

---

## Package Responsibilities

| Package | Responsibility |
|---|---|
| `hls` | Parse, filter, build, serialize HLS master playlists |
| `dash` | Parse, filter, build, serialize DASH MPDs |
| `manifest` | Auto-detect format; unified `Filter`, `FilterFromURL`, `FilterFromFile`, `Build` |
| `server` | HTTP handlers for `GET /filter` and `POST /build` |
| `cmd/manifestor` | CLI subcommands: `filter`, `build`, `serve` |

---

## Coding Rules

1. **No external dependencies** in `hls/`, `dash/`, `manifest/`, or `server/`.
2. All exported symbols must have **Go doc comments** ending with a period.
3. Filters compose with **AND logic** — a variant must pass every active filter.
4. Transformers apply **after** filters (only to surviving variants/representations).
5. `Build()` must validate all required fields before emitting any output.
6. No global state — all config via options or builder methods.
7. Test coverage must stay **≥ 80%** on core packages.
8. Every filter and transformer needs a unit test using a fixture from `testdata/`.

---

## Sentinel Errors (must not be changed or removed)

| Error | Package | Condition |
|---|---|---|
| `ErrNotMasterPlaylist` | `hls` | Content is a media playlist, not a master |
| `ErrNoVariantsRemain` | `hls`, `dash` | All variants filtered out |
| `ErrParseFailure` | `hls`, `dash` | Cannot parse manifest |
| `ErrEmptyVariantList` | `hls`, `dash` | `Build()` called with no variants |
| `ErrInvalidVariant` | `hls`, `dash` | Variant missing `URI` or `Bandwidth` |
| `ErrOrphanedGroupID` | `hls` | `AudioGroupID` has no matching `EXT-X-MEDIA` group |
| `ErrInvalidLanguageTag` | `dash` | `lang` is not a valid BCP-47 tag |
| `ErrInvalidFormat` | `manifest` | Content is neither HLS nor DASH |
| `ErrFetchFailed` | `manifest` | Upstream URL fetch failed |

---

## Test Fixtures

Real-world manifests live in `testdata/`. Use them for unit tests — do not generate synthetic manifests inline when a fixture covers the case.

| File | Source |
|---|---|
| `testdata/hls/bento4_master.m3u8` | Bento4 `mp4-dash.py` output (H.264 + H.265) |
| `testdata/hls/shaka_master.m3u8` | Shaka Packager HLS output |
| `testdata/hls/aws_mediaconvert_master.m3u8` | AWS MediaConvert HLS output |
| `testdata/dash/isoff_ondemand.mpd` | MPEG-DASH ISOFF on-demand profile |
| `testdata/dash/isoff_live.mpd` | MPEG-DASH ISOFF live profile with `SegmentTemplate` |
| `testdata/dash/azure_media_services.mpd` | Azure Media Services DASH output |

---

## Implementation Status

All source files are **skeletons** — types, errors, options, and function signatures are defined but bodies contain `panic("not implemented")`. Implementation follows the phased plan in `REQUIREMENT.md`:

- **Phase 1 (MVP):** HLS + DASH parsers, writers, filters, builders
- **Phase 2:** Advanced filters, HTTP server, subtitle/I-Frame support, BCP-47 validation
- **Phase 3:** CLI tool, benchmarks, CI polish

When implementing a function, remove the `panic("not implemented")` and add a corresponding test using a fixture from `testdata/`.

---

## CI Checks (must all pass)

| Job | Command |
|---|---|
| Test | `go test -race ./...` (Go 1.22 and 1.23) |
| Format | `gofmt -l .` (zero output required) |
| Lint | `golangci-lint run` |
| Bench | `go test -bench=. -benchmem -run='^$' ./...` |

---

## Pull Request Flow

1. Branch from `main`: `git checkout -b feat/<name>` or `fix/<name>`
2. Implement with tests
3. `gofmt -w .` — format all files
4. `go test -race ./...` — all tests pass
5. `golangci-lint run` — no issues
6. Update `CHANGELOG.md` under `[Unreleased]`
7. Push and open PR against `main`
