# manifestor

[![Go Reference](https://pkg.go.dev/badge/github.com/alanzng/manifestor.svg)](https://pkg.go.dev/github.com/alanzng/manifestor)
[![CI](https://github.com/alanzng/manifestor/actions/workflows/ci.yml/badge.svg)](https://github.com/alanzng/manifestor/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alanzng/manifestor)](https://goreportcard.com/report/github.com/alanzng/manifestor)
[![Coverage](https://codecov.io/gh/alanzng/manifestor/branch/main/graph/badge.svg)](https://codecov.io/gh/alanzng/manifestor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Parse, filter, build, and transform HLS & DASH manifests in Go. Zero dependencies. Ships as a library, HTTP proxy server, and CLI tool.

---

## Features

- **Parse** HLS Master Playlists (`.m3u8`) and DASH MPDs (`.mpd`) from string, file, or URL
- **Filter** variants/representations by codec, resolution, bandwidth, frame rate, audio language, MIME type
- **Transform** URIs — CDN rewrites, absolute URI resolution, auth token injection (variants, audio tracks, I-frame streams, and DASH `<BaseURL>` elements)
- **Inject** extra tracks — append subtitle tracks, alternate audio, or additional representations after filtering
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
docker pull ghcr.io/alanng/manifestor:latest
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

### Real-world example — Vieon VOD pipeline

Take a Bento4-generated master playlist with mixed AVC1/HVC1 video and a single audio track, and produce a delivery manifest with H.265 only, max 720p, absolute CDN URLs, a dubbed audio track, and subtitles.

**HLS:**

```go
import (
    "github.com/alanzng/manifestor/hls"
    "github.com/alanzng/manifestor/manifest"
)

const cdnBase = "https://vod-bp.vieon.vn/abc123/.../vod/2026/03/12/uuid/"
const dubbedBase = "https://vod-bp.vieon.vn/def456/.../vod/2026/03/24/uuid2/"

out, err := manifest.Filter(content,
    manifest.WithCodec("h265"),
    manifest.WithMaxResolution(1280, 720),
    manifest.WithAbsoluteURIs(cdnBase),
    manifest.WithHLSVariantSubtitleGroup("subs"),
    manifest.WithHLSInjectSubtitle(hls.SubtitleTrackParams{
        GroupID:  "subs",
        Name:     "Tiếng Việt",
        Language: "vi",
        URI:      "https://static.vieon.vn/subtitle/vi.m3u8",
        Default:  true,
    }),
    manifest.WithHLSInjectAudioTrack(hls.AudioTrackParams{
        GroupID:  "audio/mp4a",
        Name:     "Thuyết Minh",
        Language: "tm",
        URI:      dubbedBase + "audio-tg-mp4a.m3u8",
    }),
)
```

**DASH:**

```go
import (
    "github.com/alanzng/manifestor/dash"
    "github.com/alanzng/manifestor/manifest"
)

out, err := manifest.Filter(content,
    manifest.WithCodec("h265"),
    manifest.WithMaxResolution(1280, 720),
    manifest.WithAbsoluteURIs(cdnBase),
    manifest.WithDASHInjectAdaptationSet(dash.AdaptationSetParams{
        MimeType: "audio/mp4",
        Lang:     "tm",
        Name:     "Thuyết Minh",
        Representations: []dash.RepresentationParams{
            {ID: "tm-audio", Bandwidth: 196728, Codecs: "mp4a.40.2",
                BaseURL: dubbedBase + "media-audio-tg-mp4a.mp4"},
        },
    }),
    manifest.WithDASHInjectAdaptationSet(dash.AdaptationSetParams{
        ContentType: "text",
        MimeType:    "text/vtt",
        Lang:        "vi",
        Roles:       []dash.Role{{SchemeIDURI: "urn:mpeg:dash:role:2011", Value: "subtitle"}},
        Representations: []dash.RepresentationParams{
            {ID: "subtitles/vi", Bandwidth: 16,
                BaseURL: "https://static.vieon.vn/subtitle/vi.vtt"},
        },
    }),
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
b.AddAdaptationSet(dash.AdaptationSetParams{
    MimeType: "audio/mp4",
    Lang:     "en",
    Name:     "English",
    Representations: []dash.RepresentationParams{
        {
            ID:        "a1",
            Bandwidth: 128000,
            Codecs:    "mp4a.40.2",
            BaseURL:   "https://cdn.example.com/audio-en.mp4",
            AudioChannelConfiguration: &dash.AudioChannelConfiguration{
                SchemeIDURI: "urn:mpeg:dash:23003:3:audio_channel_configuration:2011",
                Value:       "2",
            },
        },
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

### Filter Options (unified — work on both HLS and DASH)

| Option | Description |
|---|---|
| `WithCodec(codec)` | Keep only **video** variants matching codec: `h264`, `h265`, `vp9`, `av1`. Audio tracks are always preserved. |
| `WithMaxResolution(w, h)` | Exclude variants wider or taller than `w×h` |
| `WithMinResolution(w, h)` | Exclude variants smaller than `w×h` |
| `WithExactResolution(w, h)` | Keep only variants with exactly `w×h` |
| `WithMaxBandwidth(bps)` | Exclude variants above `bps` bits/s |
| `WithMinBandwidth(bps)` | Exclude variants below `bps` bits/s |
| `WithMaxFrameRate(fps)` | Exclude variants with frame rate above `fps` |
| `WithAudioLanguage(lang)` | Keep only audio tracks matching BCP-47 `lang` |
| `WithMimeType(mime)` | Keep only representations matching MIME type (DASH only) |
| `WithCDNBaseURL(base)` | Rewrite all URIs to use `base` as CDN origin |
| `WithAbsoluteURIs(origin)` | Resolve relative URIs to absolute using `origin` |
| `WithAuthToken(token)` | Append `token=` query parameter to all URIs |

URI rewriting covers: HLS variant URIs, audio track URIs, I-frame stream URIs; DASH `<BaseURL>` elements.

### HLS-only filter options

| Option | Description |
|---|---|
| `WithHLSInjectVariant(p)` | Append a variant stream after filtering |
| `WithHLSInjectAudioTrack(p)` | Append an `#EXT-X-MEDIA AUDIO` track after filtering |
| `WithHLSInjectSubtitle(p)` | Append an `#EXT-X-MEDIA SUBTITLES` track after filtering |
| `WithHLSVariantSubtitleGroup(id)` | Set `SUBTITLES="id"` on all surviving variants |

### DASH-only filter options

| Option | Description |
|---|---|
| `WithDASHInjectAdaptationSet(p)` | Append an `<AdaptationSet>` to every Period after filtering |

### Custom callbacks (package-level only)

| Option | Package | Description |
|---|---|---|
| `WithCustomFilter(fn)` | `hls`, `dash` | User-defined filter: `func(*Variant) bool` / `func(*Representation) bool` |
| `WithCustomTransformer(fn)` | `hls`, `dash` | User-defined transformer applied to each surviving variant/representation |

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

## DASH Data Model

Key fields parsed and round-tripped through `Parse` → `Filter` → `Serialize`:

| Element | Fields |
|---|---|
| `<MPD>` | `Profile`, `Duration`, `MinBufferTime`, `MinUpdatePeriod` |
| `<AdaptationSet>` | `ID`, `ContentType`, `MimeType`, `Lang`, `Name` (label attr), `Roles`, `SegmentTemplate`, `SegmentBase` |
| `<Representation>` | `ID`, `Bandwidth`, `Codecs`, `Width`, `Height`, `FrameRate`, `MimeType`, `StartWithSAP`, `BaseURL`, `AudioChannelConfiguration` |
| `<Role>` | `SchemeIDURI`, `Value` |
| `<AudioChannelConfiguration>` | `SchemeIDURI`, `Value` |

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
- Vieon VOD platform

---

## Who is using manifestor

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Logo</th>
            <th>Website</th>
            <th>Description</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td><strong>Vieon</strong></td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="docs/images/vieon-logo-dark.svg" />
                    <img height="32px" src="docs/images/vieon-logo-light.svg" alt="Vieon" />
                </picture>
            </td>
            <td><a href="https://vieon.vn/">vieon.vn</a></td>
            <td>Vietnam's leading OTT streaming platform. Uses manifestor to filter and transform HLS &amp; DASH manifests for multi-codec VOD delivery (AVC1 + HVC1) with CDN rewriting and per-request auth tokens.</td>
        </tr>
    </tbody>
</table>

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

---

## License

MIT — see [LICENSE](LICENSE).
