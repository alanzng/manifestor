# manifestor

[![Go Reference](https://pkg.go.dev/badge/github.com/alanzng/manifestor.svg)](https://pkg.go.dev/github.com/alanzng/manifestor)
[![CI](https://github.com/alanzng/manifestor/actions/workflows/ci.yml/badge.svg)](https://github.com/alanzng/manifestor/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alanzng/manifestor)](https://goreportcard.com/report/github.com/alanzng/manifestor)
[![Coverage](https://codecov.io/gh/alannguyen/manifestor/branch/main/graph/badge.svg)](https://codecov.io/gh/alannguyen/manifestor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Parse, filter, build, and transform HLS & DASH manifests in Go. Zero dependencies. Ships as a library, HTTP proxy server, and CLI tool.

---

## Features

- **Parse** HLS Master Playlists (`.m3u8`) and DASH MPDs (`.mpd`) from string, file, or URL
- **Filter** variants/representations by codec, resolution, bandwidth, frame rate, audio language, MIME type
- **Transform** URIs — CDN rewrites, absolute URI resolution, auth token injection
- **Build** complete HLS or DASH manifests from scratch with a fluent builder API
- **Serve** as an HTTP proxy that filters manifests on the fly
- **CLI** tool for scripting and local use
- Zero non-stdlib dependencies
- Thread-safe — `Filter()` is safe for concurrent use
- HLS versions 3–7; DASH profiles `isoff-on-demand` and `isoff-live`

---

## Installation

### Library

```bash
go get github.com/alanzng/manifestor
```

### CLI

```bash
go install github.com/alanzng/manifestor/cmd/manifestor@latest
```

### Docker

```bash
docker pull ghcr.io/alannguyen/manifestor:latest
```

---

## Quick Start

### Filter an HLS manifest

```go
import "github.com/alanzng/manifestor/manifest"

filtered, err := manifest.Filter(content,
    manifest.WithCodec("h264"),
    manifest.WithMaxResolution(1920, 1080),
    manifest.WithMaxBandwidth(5_000_000),
    manifest.WithCDNBaseURL("https://cdn.example.com"),
)
```

### Filter from URL

```go
filtered, err := manifest.FilterFromURL("https://example.com/master.m3u8",
    manifest.WithCodec("h264"),
    manifest.WithAuthToken("token=abc123"),
)
```

### Build an HLS Master Playlist

```go
import "github.com/alanzng/manifestor/hls"

b := hls.NewMasterBuilder()
b.SetVersion(6).
    AddAudioTrack(hls.AudioTrackParams{
        GroupID:    "audio-en",
        Name:       "English",
        Language:   "en",
        URI:        "https://cdn.example.com/audio/en/index.m3u8",
        Default:    true,
        AutoSelect: true,
    }).
    AddVariant(hls.VariantParams{
        URI:          "https://cdn.example.com/1080p/index.m3u8",
        Bandwidth:    5_000_000,
        Codecs:       "avc1.640028,mp4a.40.2",
        Width:        1920,
        Height:       1080,
        FrameRate:    29.97,
        AudioGroupID: "audio-en",
    }).
    AddVariant(hls.VariantParams{
        URI:          "https://cdn.example.com/720p/index.m3u8",
        Bandwidth:    2_800_000,
        Codecs:       "avc1.4d401f,mp4a.40.2",
        Width:        1280,
        Height:       720,
        FrameRate:    29.97,
        AudioGroupID: "audio-en",
    })

playlist, err := b.Build()
```

### Build a DASH MPD

```go
import "github.com/alanzng/manifestor/dash"

b := dash.NewMPDBuilder(dash.MPDConfig{
    Profile:       "isoff-on-demand",
    Duration:      "PT4M0.00S",
    MinBufferTime: "PT1.5S",
})
b.AddAdaptationSet(dash.AdaptationSetParams{
    MimeType: "video/mp4",
    SegmentTemplate: &dash.SegmentTemplateParams{
        Initialization: "$RepresentationID$/init.mp4",
        Media:          "$RepresentationID$/$Number$.m4s",
        Timescale:      90000,
        Duration:       270000,
    },
    Representations: []dash.RepresentationParams{
        {ID: "v1", Bandwidth: 5_000_000, Codecs: "avc1.640028", Width: 1920, Height: 1080},
        {ID: "v2", Bandwidth: 2_000_000, Codecs: "avc1.4d401f", Width: 1280, Height: 720},
    },
})

mpd, err := b.Build()
```

### HTTP Server

```bash
# Start the proxy server
manifestor serve --port 8080

# Filter a live manifest via HTTP
curl "http://localhost:8080/filter?url=https://example.com/master.m3u8&codec=h264&max_res=1920x1080"
```

### CLI

```bash
# Filter a local file
manifestor filter --input master.m3u8 --codec h264 --max-res 1920x1080

# Filter from URL and write to file
manifestor filter --url https://example.com/master.m3u8 --codec h264 --output filtered.m3u8

# Build from a JSON spec
manifestor build --format hls --variants spec.json --output master.m3u8
```

---

## API Reference

### Filter Options

| Option | Description |
|---|---|
| `WithCodec(codec)` | Keep only variants matching codec: `h264`, `h265`, `vp9`, `av1` |
| `WithMaxResolution(w, h)` | Exclude variants wider or taller than `w×h` |
| `WithMinResolution(w, h)` | Exclude variants smaller than `w×h` |
| `WithExactResolution(w, h)` | Keep only variants with exactly `w×h` |
| `WithMaxBandwidth(bps)` | Exclude variants above `bps` bits/s |
| `WithMinBandwidth(bps)` | Exclude variants below `bps` bits/s |
| `WithMaxFrameRate(fps)` | Exclude variants with frame rate above `fps` |
| `WithAudioLanguage(lang)` | Keep only audio tracks matching BCP-47 `lang` |
| `WithMimeType(mime)` | Keep only representations matching MIME type |
| `WithCDNBaseURL(base)` | Rewrite all URIs to use `base` as the CDN origin |
| `WithAbsoluteURIs(origin)` | Resolve relative URIs to absolute using `origin` |
| `WithAuthToken(token)` | Append `token` as a query string to all URIs |
| `WithCustomFilter(fn)` | User-defined filter function `func(v *Variant) bool` |
| `WithCustomTransformer(fn)` | User-defined transformer `func(v *Variant)` |

Filters compose with **AND** logic — a variant must pass all filters to survive.

### Errors

| Error | Condition |
|---|---|
| `ErrInvalidFormat` | Content is neither valid HLS nor DASH |
| `ErrNotMasterPlaylist` | HLS content is a media playlist, not a master |
| `ErrNoVariantsRemain` | All variants were filtered out |
| `ErrFetchFailed` | Upstream URL fetch failed |
| `ErrParseFailure` | Manifest could not be parsed |
| `ErrEmptyVariantList` | `Build()` called with no variants added |
| `ErrInvalidVariant` | A variant is missing `URI` or `Bandwidth` |
| `ErrOrphanedGroupID` | `AudioGroupID` references a non-existent `#EXT-X-MEDIA` group |
| `ErrInvalidLanguageTag` | DASH `lang` is not a valid BCP-47 tag |

---

## HTTP API

### `GET /filter`

Fetches and filters an upstream manifest.

| Parameter | Required | Description |
|---|---|---|
| `url` | yes | Upstream manifest URL |
| `codec` | no | `h264` \| `h265` \| `vp9` \| `av1` |
| `max_res` | no | e.g. `1920x1080` |
| `min_res` | no | e.g. `854x480` |
| `max_bw` | no | bits/s e.g. `5000000` |
| `min_bw` | no | bits/s e.g. `500000` |
| `fps` | no | max frame rate e.g. `30` |
| `cdn` | no | CDN base URL |
| `token` | no | Auth token string |
| `lang` | no | BCP-47 audio language |

**Responses:** `200 OK`, `400 Bad Request`, `422 Unprocessable Entity`, `502 Bad Gateway`

### `POST /build`

Builds a manifest from a JSON payload. See [HTTP API docs](docs/http-api.md) for full schema.

---

## Performance

| Operation | Target | Typical |
|---|---|---|
| Parse + filter + serialize 50 KB manifest | < 5 ms | ~1–2 ms |
| Build 100-variant manifest | < 2 ms | ~0.5 ms |

---

## Supported Manifest Sources

Tested against real-world output from:

- [Bento4](https://www.bento4.com/) `mp4-dash.py`
- [Shaka Packager](https://github.com/shaka-project/shaka-packager)
- AWS MediaConvert
- Azure Media Services

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

---

## License

MIT — see [LICENSE](LICENSE).
