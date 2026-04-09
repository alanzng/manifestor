# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

_No changes yet._

---

## [0.5.0] - 2026-04-09

### Changed

- **Typed media identifiers (root `manifestor` package).** Codec is now a named type with constants `H264`, `H265`, `VP9`, and `AV1`, plus `ParseCodec` and `Codec.MatchesCodec` for RFC 6381 `CODECS` matching. Unknown codec strings return `*InvalidCodecError`.
- **Resolution** is a `struct { Width, Height int }` with presets `Res360p` through `Res4K`, `ParseResolution` for `WxH` strings, and `String()` for stable serialization.
- **MimeType** is a string-based type with constants such as `MimeVideoMP4`, `MimeAudioMP4`, and `MimeTextVTT` for DASH representation filtering.
- **`manifest`, `hls`, and `dash` options** that previously took raw strings for codec, resolution, or MIME now take these types (for example `manifest.WithCodec`, `WithMaxResolution`, `WithMimeType`, and the corresponding package-local `With*` functions).
- **CLI and HTTP server** continue to accept human-readable strings (`h264`, `1920x1080`, etc.) and parse them with `ParseCodec` / `ParseResolution` before building options.
- README quick-start examples, feature list, and the unified API table were rewritten to match the new types.

---

## [0.4.0] - 2026-04-08

### Added

- **`manifestor` CLI** (`cmd/manifestor`):
  - `filter` — load a manifest from `--url` or `--input`, apply options, write to stdout or `--output`.
  - `build` — `--format` (`hls`|`dash`), `--variants` (path to JSON spec), optional `--output`, HLS-only `--version`, DASH-only `--duration` / `--profile` (`ondemand`|`live`), optional post-build `--cdn` / `--token` (also applied when set in JSON).
  - `serve` — run the HTTP proxy locally (`--port`, default 8080; `--timeout` for upstream fetch, default 10s).
- **Filter flags:** `--codec` (`h264|h265|vp9|av1`), `--max-res` / `--min-res` (`WxH`), `--max-bw` / `--min-bw` (bits/s), `--fps` (max frame rate), `--cdn`, `--token`, `--lang` (BCP-47 audio), `--origin` (base for resolving relative URIs).
- **`server` package:** `GET /filter` with required `url` query parameter and optional `codec`, `max_res`, `min_res`, `max_bw`, `min_bw`, `fps`, `cdn`, `token`, `lang`, and `origin` (same semantics as the library options). Responses use `application/vnd.apple.mpegurl` or `application/dash+xml` based on content.
- **`POST /build`** with a JSON body: `format` (`hls` or `dash`), structured variants/audio/subtitles (HLS) or periods/adaptation sets (DASH), plus optional post-build `cdn` and `token` fields that run a second `manifest.Filter` pass.
- **HTTP errors:** distinct status codes for bad input (400), empty/invalid build state (422), upstream fetch failures (`ErrFetchFailed` → 502), and related edge cases.
- **Test coverage:** expanded tests for `cmd/manifestor`, `server`, and `manifest` error paths (`ErrInvalidFormat`, `ErrFetchFailed`, body read failures), resolution parsing, post-build token/CDN flows, and filter branches (height “exact” logic, I-frames, DASH MIME inheritance, HDCP-related attributes, etc.); tests use `runtime.GOROOT` for the `go` binary and context-aware HTTP where appropriate.

### Changed

- **Codecov:** added `codecov.yml` with patch coverage target **70%** (5% threshold), project auto target with 1% threshold, and **`cmd/manifestor/main.go` excluded** from coverage so the thin CLI entrypoint does not dominate patch metrics.

### Fixed

- When resolving relative manifest URLs, ensure the `origin` URL used with `ResolveReference` ends with `/` so path resolution matches browser-style base URLs.

---

## [0.3.0] - 2026-04-07

### Changed

- **Performance:** less allocation and redundant work across HLS and DASH code paths (including shared URI rewrite logic), without changing public behavior.
- **Documentation layout:** `REQUIREMENT.md` moved to [`docs/REQUIREMENT.md`](docs/REQUIREMENT.md); images relocated under `docs/images/`.
- **README “Who is using manifestor”:** expanded Vieon entry (including docker image path corrections and richer context).
- **Tests:** broader `rewriteURI` error-path coverage; removed stray debug output from HLS tests.

### Fixed

- Removed dead helpers left over after consolidating `rewriteURI` implementations.

---

## [0.2.0] - 2026-04-07

### Fixed

- **HLS `IFrameStream`:** when serializing `EXT-X-I-FRAME-STREAM-INF`, emit `AVERAGE-BANDWIDTH` when it was present on the parsed structure (previously dropped).
- **HLS `Raw` tag lines:** only actual HLS tags belong in `Variant.Raw` / related raw slices; lines that are plain comments (e.g. `# comment`) are no longer stored as if they were tag lines.

---

## [0.1.0] - 2026-04-07

### Added

- **HLS master playlists:** parse and serialize master playlists (tags and attributes), including mixed-codec real-world fixtures under `testdata/hls/`.
- **HLS filtering:** keep variants by codec family, min/max/exact resolution, bandwidth, max frame rate, audio language (BCP-47), plus URI transformers (`WithCDNBaseURL`, `WithAbsoluteURIs`, `WithAuthToken`) applied to variants, audio, subtitles, and I-frame streams.
- **HLS `MasterBuilder`:** fluent build API with validation (`ErrEmptyVariantList`, `ErrInvalidVariant`, `ErrOrphanedGroupID`, etc.).
- **HLS inject API:** `WithInjectVariant`, `WithInjectAudioTrack`, `WithInjectSubtitle` to append synthetic tracks after filtering.
- **DASH MPD:** XML parser and writer for ISOFF-style manifests (`testdata/dash/` fixtures: on-demand, live with `SegmentTemplate`, Azure Media Services).
- **DASH filtering:** codec, resolution, bandwidth, language, and **MIME type** (representation / adaptation-set level, including inherited `mimeType`).
- **DASH URI transforms:** `BaseURL` elements participate in rewrite and resolution; `WithInjectAdaptationSet` can append adaptation sets after filtering.
- **DASH model extensions:** `BaseURL`, `AudioChannelConfiguration`, `Role`, `Name`, and related fields for richer round-trips.
- **`manifest` package:** `Detect`, `Filter`, `FilterFromURL`, `FilterFromFile`, `Build`, and a single `manifest.Option` surface wired through to both HLS and DASH (including inject options and `WithHLSVariantSubtitleGroup` for subtitle group wiring on the unified API).
- **Integration tests** against Vieon-style VOD HLS and DASH manifests (URI rewriting, codec filters, subtitles/audio).
- **Project docs:** `AGENTS.md`, requirements, CI (tests, format, lint, bench), Codecov/GitHub Actions wiring, and initial “Who is using” README content.

[Unreleased]: https://github.com/alanzng/manifestor/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/alanzng/manifestor/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/alanzng/manifestor/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/alanzng/manifestor/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/alanzng/manifestor/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/alanzng/manifestor/compare/24af02ef2c6d0c2d667bf56df934ff92dc247761...v0.1.0
