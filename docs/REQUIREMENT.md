# 📋 Requirements Document
## Library: `manifestor` — HLS & DASH Manifest Manipulation Library

---

## 1. Overview

### Problem Statement

There is no actively maintained, production-ready, server-side library that supports both HLS (`.m3u8`) and DASH (`.mpd`) manifest parsing, filtering, and rewriting in a unified API. Existing libraries either:

- Cover only one format
- Are archived / unmaintained
- Are browser-based players (not server-side tools)
- Have heavy dependencies

### Goal

Build a server-side library that allows developers to parse, filter, transform, and **build from scratch** HLS and DASH manifests programmatically — usable as a Go library, HTTP proxy server, or CLI tool.

---

## 2. Scope

### In Scope

- Parse HLS Master Playlist (`.m3u8`)
- Parse MPEG-DASH Media Presentation Description (`.mpd`)
- Filter variants/representations by codec, resolution, bandwidth, frame rate
- Transform URIs (CDN rewriting, absolute URI resolution, token injection)
- Serialize filtered manifest back to valid HLS or DASH format
- **Build a new HLS Master Playlist from scratch** (no existing manifest required)
- **Build a new DASH MPD from scratch** (no existing manifest required)
- HTTP proxy server mode
- CLI tool mode

### Out of Scope

- Segment downloading or streaming
- DRM / encryption handling
- Media transcoding
- Live stream clock management
- Browser-side playback

---

## 3. Functional Requirements

### 3.1 Parsing

| ID | Requirement |
|---|---|
| P-01 | Parse HLS Master Playlist from string, file, or URL |
| P-02 | Parse HLS Media Playlist from string, file, or URL |
| P-03 | Parse DASH MPD from string, file, or URL |
| P-04 | Auto-detect format (HLS vs DASH) from content or Content-Type header |
| P-05 | Preserve unrecognized tags/attributes (pass-through unknown lines) |
| P-06 | Support HLS version 3–7 |
| P-07 | Support DASH profile: `isoff-on-demand`, `isoff-live` |

### 3.2 Filtering

| ID | Requirement |
|---|---|
| F-01 | Filter by video codec: `h264` (avc1), `h265` (hvc1/hev1), `vp9`, `av1` |
| F-02 | Filter by maximum resolution (width × height) |
| F-03 | Filter by minimum resolution (width × height) |
| F-04 | Filter by exact resolution |
| F-05 | Filter by maximum bandwidth |
| F-06 | Filter by minimum bandwidth |
| F-07 | Filter by maximum frame rate |
| F-08 | Filter by audio language (DASH `AdaptationSet@lang`, HLS `EXT-X-MEDIA` LANGUAGE) |
| F-09 | Filter by MIME type (`video/mp4`, `video/webm`) |
| F-10 | Multiple filters compose with AND logic (all must pass) |
| F-11 | Custom filter escape hatch: accept user-defined filter function |
| F-12 | If all variants are filtered out, return error `ErrNoVariantsRemain` |
| F-13 | Audio tracks are preserved by default unless explicitly filtered |
| F-14 | I-Frame playlists (HLS) follow the same codec/resolution filters |

### 3.3 Transformation (URI Rewriting)

| ID | Requirement |
|---|---|
| T-01 | Rewrite all relative URIs to absolute using origin base URL |
| T-02 | Replace origin base URL with CDN base URL |
| T-03 | Append query string token to all URIs (e.g. `?token=xxx`) |
| T-04 | Custom transformer escape hatch: accept user-defined transform function |
| T-05 | Transformers apply after filters (only to surviving variants) |
| T-06 | Transformers are composable (multiple can be chained) |
| T-07 | DASH `<BaseURL>` element per `Representation` must be parsed, rewritten by T-01/T-02/T-03, and re-serialized |

### 3.4 DASH Manifest Injection & Merging

| ID | Requirement |
|---|---|
| M-01 | Inject one or more additional `AdaptationSet` entries into a filtered DASH MPD (e.g. add a dubbed audio track from a different source) |
| M-02 | Injected `AdaptationSet` entries support all fields: `lang`, `mimeType`, `contentType`, `name`, `Representation`, `BaseURL`, `SegmentBase`, `AudioChannelConfiguration` |
| M-03 | `AdaptationSet` must support a human-readable `name` attribute (e.g. `name="Tiếng Gốc"`, `name="Thuyết Minh"`) |
| M-04 | Support subtitle `AdaptationSet` with `contentType="text"`, arbitrary `mimeType` (e.g. `text/vtt`), and a `<Role>` child element (`schemeIdUri`, `value`) |
| M-05 | `Representation` inside a text/subtitle `AdaptationSet` requires only `id`, `bandwidth`, and `<BaseURL>`; all other fields are optional |
| M-06 | `<AudioChannelConfiguration>` child element on `Representation` must be parsed and re-serialized intact |

### 3.5 Serialization

| ID | Requirement |
|---|---|
| S-01 | Serialize filtered HLS back to valid `.m3u8` string |
| S-02 | Serialize filtered DASH back to valid `.mpd` XML string |
| S-03 | Preserve `#EXT-X-VERSION` and other header tags in HLS output |
| S-04 | Preserve MPD-level attributes (duration, profiles, etc.) in DASH output |
| S-05 | Output must be parseable by standard players (hls.js, dash.js, shaka-player) |

### 3.6 Building (Composing from Scratch)

#### 3.5.1 Shared / Unified

| ID | Requirement |
|---|---|
| B-01 | Provide a unified `manifest.Build(format, ...)` entry point accepting a `Format` enum (`FormatHLS` or `FormatDASH`) |
| B-02 | Validate all required fields before building; return `ErrInvalidVariant` on the first invalid entry |
| B-03 | Return `ErrEmptyVariantList` if no variants or representations are added before calling `Build()` |
| B-04 | All existing transformers (T-01 – T-06) must be applicable to a built manifest after construction |
| B-05 | Built output must satisfy S-05 (parseable by hls.js, dash.js, shaka-player) |
| B-06 | Built manifest must be valid UTF-8 with no BOM |

#### 3.5.2 HLS Master Playlist Builder

| ID | Requirement |
|---|---|
| BH-01 | Build a `#EXTM3U` master playlist from a list of `VariantParams` (no input manifest required) |
| BH-02 | Required fields per video variant: `URI` (string), `Bandwidth` (int) |
| BH-03 | Optional fields per video variant: `AverageBandwidth`, `Codecs`, `Width`, `Height`, `FrameRate`, `AudioGroupID`, `SubtitleGroupID`, `HDCPLevel` |
| BH-04 | Support adding `#EXT-X-MEDIA` entries for `TYPE=AUDIO`, `TYPE=SUBTITLES`, and `TYPE=CLOSED-CAPTIONS` |
| BH-05 | Required fields per `#EXT-X-MEDIA` entry: `GroupID`, `Name`, `Type` |
| BH-06 | Optional fields per `#EXT-X-MEDIA` entry: `Language`, `URI`, `Default`, `AutoSelect`, `Forced` |
| BH-07 | Support adding `#EXT-X-I-FRAME-STREAM-INF` entries (URI + Bandwidth required; Codecs, Resolution optional) |
| BH-08 | Allow caller to set HLS version (`#EXT-X-VERSION`); default to `3` if not specified |
| BH-09 | Emit variants in the order they were added (caller controls quality ordering) |
| BH-10 | If `AudioGroupID` is set on a variant, a matching `#EXT-X-MEDIA` `GroupID` must exist; return `ErrOrphanedGroupID` otherwise |

#### 3.5.3 DASH MPD Builder

| ID | Requirement |
|---|---|
| BD-01 | Build a complete MPD document from `MPDConfig` + one or more `AdaptationSetParams` (no input manifest required) |
| BD-02 | Required `MPDConfig` fields: `Profile` (`isoff-on-demand` or `isoff-live`) |
| BD-03 | Optional `MPDConfig` fields: `Duration` (ISO 8601, e.g. `PT4M0.00S`), `MinBufferTime` (default `PT1.5S`), `MinUpdatePeriod` (live only) |
| BD-04 | Support multiple `AdaptationSet` entries per `Period` (video, audio, text) |
| BD-05 | Required fields per `Representation`: `ID` (string, unique within AdaptationSet), `Bandwidth` (int) |
| BD-06 | Optional fields per `Representation`: `Codecs`, `Width`, `Height`, `FrameRate`, `MimeType`, `StartWithSAP` |
| BD-07 | Support `SegmentTemplate` addressing: `initialization`, `media` template, `timescale`, `duration` |
| BD-08 | Support `SegmentBase` addressing: `indexRange`, `Initialization` element (for single-file on-demand) |
| BD-09 | Emit correct MPD XML namespace (`urn:mpeg:dash:schema:mpd:2011`) and `xmlns:xsi` / `xsi:schemaLocation` |
| BD-10 | `AdaptationSet` `contentType` is inferred from `MimeType` if not explicitly set (`video/…` → `video`, `audio/…` → `audio`, `text/…` → `text`) |
| BD-11 | Representations within an AdaptationSet are emitted in ascending `Bandwidth` order by default |
| BD-12 | `lang` attribute on `AdaptationSet` must be a valid BCP-47 language tag; return `ErrInvalidLanguageTag` if malformed |
| BD-13 | `AdaptationSet` supports optional `name` attribute (human-readable label, e.g. `"Tiếng Gốc"`, `"Thuyết Minh"`) |
| BD-14 | `Representation` supports `<BaseURL>` child element (required for Bento4 / single-file on-demand assets) |
| BD-15 | `Representation` supports `<AudioChannelConfiguration>` child element (`schemeIdUri`, `value`) |
| BD-16 | `AdaptationSet` supports `<Role>` child element (`schemeIdUri`, `value`) for subtitle tracks |
| BD-17 | Support injecting additional `AdaptationSet` entries into a filtered MPD via `WithInjectAdaptationSet` option |

---

## 4. Non-Functional Requirements

| ID | Requirement |
|---|---|
| NF-01 | **Zero non-stdlib dependencies** for core parse/filter/serialize |
| NF-02 | Single binary — no runtime dependencies |
| NF-03 | Parse a 50KB manifest in under 5ms |
| NF-04 | Thread-safe: `Filter()` must be safe to call concurrently |
| NF-05 | All public APIs must have Go doc comments |
| NF-06 | Test coverage ≥ 80% on core packages |
| NF-07 | Each filter/transformer must have a unit test with real-world sample manifests |
| NF-08 | No global state |
| NF-09 | `Build()` must validate and serialize a 100-variant manifest in under 2ms |

---

## 5. API Requirements

### 5.1 Library API (Go)

```go
// Unified — auto-detect format
manifest.Filter(content string, opts ...Option) (string, error)
manifest.FilterFromURL(url string, opts ...Option) (string, error)
manifest.FilterFromFile(path string, opts ...Option) (string, error)

// Format-specific
hls.Filter(content string, opts ...Option) (string, error)
dash.Filter(content string, opts ...Option) (string, error)

// Built-in options
manifest.WithCodec(codec string) Option                    // "h264", "h265", "vp9", "av1"
manifest.WithMaxResolution(w, h int) Option
manifest.WithMinResolution(w, h int) Option
manifest.WithExactResolution(w, h int) Option
manifest.WithMaxBandwidth(bps int) Option
manifest.WithMinBandwidth(bps int) Option
manifest.WithMaxFrameRate(fps float64) Option
manifest.WithAudioLanguage(lang string) Option
manifest.WithMimeType(mime string) Option
manifest.WithCDNBaseURL(base string) Option
manifest.WithAbsoluteURIs(origin string) Option
manifest.WithAuthToken(token string) Option
manifest.WithCustomFilter(fn func(v *Variant) bool) Option
manifest.WithCustomTransformer(fn func(v *Variant)) Option
```

#### Builder API

```go
// --- Shared types ---

type Format int
const (
    FormatHLS  Format = iota
    FormatDASH
)

// Unified entry point — caller picks format
manifest.Build(format Format, opts ...BuildOption) (string, error)

// --- HLS builder ---

type VariantParams struct {
    URI              string   // required
    Bandwidth        int      // required (bits/s)
    AverageBandwidth int      // optional
    Codecs           string   // optional e.g. "avc1.640028,mp4a.40.2"
    Width, Height    int      // optional
    FrameRate        float64  // optional
    AudioGroupID     string   // optional — must match an AudioTrackParams.GroupID
    SubtitleGroupID  string   // optional — must match a SubtitleTrackParams.GroupID
    HDCPLevel        string   // optional e.g. "TYPE-0"
}

type AudioTrackParams struct {
    GroupID    string // required
    Name       string // required
    Type       string // "AUDIO" (constant, inferred)
    Language   string // optional BCP-47
    URI        string // optional (omit for muxed audio)
    Default    bool
    AutoSelect bool
    Forced     bool
}

type SubtitleTrackParams struct {
    GroupID  string
    Name     string
    Language string
    URI      string // required for subtitles
    Default  bool
    Forced   bool
}

type IFrameParams struct {
    URI       string // required
    Bandwidth int    // required
    Codecs    string
    Width, Height int
}

b := hls.NewMasterBuilder()
b.SetVersion(n int) *MasterBuilder
b.AddVariant(p VariantParams) *MasterBuilder
b.AddAudioTrack(p AudioTrackParams) *MasterBuilder
b.AddSubtitleTrack(p SubtitleTrackParams) *MasterBuilder
b.AddIFrameStream(p IFrameParams) *MasterBuilder
b.Build() (string, error)

// --- DASH builder ---

type MPDConfig struct {
    Profile         string  // required: "isoff-on-demand" | "isoff-live"
    Duration        string  // optional ISO 8601 e.g. "PT4M0.00S"
    MinBufferTime   string  // optional, default "PT1.5S"
    MinUpdatePeriod string  // optional, live only
}

type RepresentationParams struct {
    ID           string  // required, unique within AdaptationSet
    Bandwidth    int     // required
    Codecs       string
    Width        int
    Height       int
    FrameRate    string  // e.g. "30000/1001"
    MimeType     string
    StartWithSAP int
}

type SegmentTemplateParams struct {
    Initialization string // e.g. "$RepresentationID$/init.mp4"
    Media          string // e.g. "$RepresentationID$/$Number$.m4s"
    Timescale      int
    Duration       int
    StartNumber    int
}

type SegmentBaseParams struct {
    IndexRange     string // e.g. "0-819"
    Initialization string // e.g. "0-499"
}

type AdaptationSetParams struct {
    ContentType     string                 // "video" | "audio" | "text" (inferred if blank)
    MimeType        string
    Lang            string                 // BCP-47
    SegmentTemplate *SegmentTemplateParams // one of these two
    SegmentBase     *SegmentBaseParams
    Representations []RepresentationParams
}

b := dash.NewMPDBuilder(cfg MPDConfig)
b.AddAdaptationSet(p AdaptationSetParams) *MPDBuilder
b.Build() (string, error)
```

### 5.2 HTTP Server API

```
GET /filter
  ?url=     required  — manifest URL to fetch and filter
  &codec=   optional  — h264 | h265 | vp9 | av1
  &max_res= optional  — e.g. 1920x1080
  &min_res= optional  — e.g. 854x480
  &max_bw=  optional  — e.g. 5000000
  &min_bw=  optional  — e.g. 500000
  &fps=     optional  — max frame rate e.g. 30
  &cdn=     optional  — CDN base URL to rewrite URIs to
  &token=   optional  — auth token to append to URIs
  &lang=    optional  — audio language filter

Response:
  200 OK  — filtered manifest with correct Content-Type
  400     — invalid parameters
  502     — failed to fetch upstream manifest
  422     — no variants remain after filtering (ErrNoVariantsRemain)

POST /build
  Content-Type: application/json
  Body (HLS example):
    {
      "format":  "hls",                    // required: "hls" | "dash"
      "version": 6,                        // HLS only, optional (default 3)
      "variants": [
        {
          "uri":       "https://cdn.example.com/1080p/index.m3u8",
          "bandwidth": 5000000,
          "codecs":    "avc1.640028,mp4a.40.2",
          "width": 1920, "height": 1080,
          "frame_rate": 29.97,
          "audio_group_id": "audio-en"
        }
      ],
      "audio_tracks": [
        {
          "group_id": "audio-en", "name": "English",
          "language": "en", "default": true, "auto_select": true,
          "uri": "https://cdn.example.com/audio/en/index.m3u8"
        }
      ]
    }
  Body (DASH example):
    {
      "format": "dash",
      "profile": "isoff-on-demand",
      "duration": "PT4M0.00S",
      "adaptation_sets": [
        {
          "content_type": "video",
          "mime_type": "video/mp4",
          "segment_template": {
            "initialization": "$RepresentationID$/init.mp4",
            "media": "$RepresentationID$/$Number$.m4s",
            "timescale": 90000, "duration": 270000
          },
          "representations": [
            { "id": "v1", "bandwidth": 5000000, "codecs": "avc1.640028", "width": 1920, "height": 1080 },
            { "id": "v2", "bandwidth": 2000000, "codecs": "avc1.4d401f", "width": 1280, "height": 720 }
          ]
        },
        {
          "content_type": "audio", "mime_type": "audio/mp4", "lang": "en",
          "representations": [
            { "id": "a1", "bandwidth": 128000, "codecs": "mp4a.40.2" }
          ]
        }
      ]
    }

Response:
  200 OK  — built manifest with correct Content-Type
  400     — invalid or missing parameters
  422     — no variants provided (ErrEmptyVariantList) or invalid variant (ErrInvalidVariant)
```

### 5.3 CLI API

```bash
manifestor filter [flags]
  --url         upstream manifest URL
  --input, -i   local file path
  --output, -o  output file (default: stdout)
  --codec       h264 | h265 | vp9 | av1
  --max-res     WxH e.g. 1920x1080
  --min-res     WxH
  --max-bw      int (bits per second)
  --min-bw      int
  --fps         float (max frame rate)
  --cdn         CDN base URL
  --token       auth token

manifestor serve [flags]
  --port        HTTP port (default: 8080)
  --timeout     upstream fetch timeout (default: 10s)

manifestor build [flags]
  --format      hls | dash  (required)
  --output, -o  output file (default: stdout)
  --variants    path to JSON file describing variants and tracks (see POST /build body schema)
  --version     HLS version number (default: 3; HLS only)
  --duration    DASH presentation duration ISO 8601 e.g. PT4M0S (DASH only)
  --profile     DASH profile: ondemand | live (DASH only)
  --cdn         CDN base URL applied to all URIs after building
  --token       auth token appended to all URIs after building
```

---

## 6. Error Handling Requirements

| Error | Condition |
|---|---|
| `ErrInvalidFormat` | Content is neither valid HLS nor DASH |
| `ErrNotMasterPlaylist` | HLS content is a media playlist, not a master |
| `ErrNoVariantsRemain` | All variants were filtered out |
| `ErrFetchFailed` | Upstream URL fetch failed |
| `ErrParseFailure` | Manifest content could not be parsed |
| `ErrEmptyVariantList` | `Build()` called with no variants or representations added |
| `ErrInvalidVariant` | A variant is missing a required field (`URI` or `Bandwidth`) |
| `ErrOrphanedGroupID` | A variant references an `AudioGroupID` or `SubtitleGroupID` with no matching `#EXT-X-MEDIA` entry |
| `ErrInvalidLanguageTag` | DASH `AdaptationSet` `lang` value is not a valid BCP-47 tag |

---

## 7. Testing Requirements

| ID | Requirement |
|---|---|
| TS-01 | Unit tests for each filter option (HLS and DASH separately) |
| TS-02 | Integration tests using real Bento4 / Shaka Packager output samples |
| TS-03 | Test: composed filters (codec + resolution + bandwidth) |
| TS-04 | Test: serialized output is re-parseable |
| TS-05 | Test: empty result returns `ErrNoVariantsRemain` |
| TS-06 | Test: relative URIs correctly resolved to absolute |
| TS-07 | Test: CDN base URL replacement is idempotent |
| TS-08 | Benchmark: parse + filter + serialize < 5ms for 50KB manifest |
| TS-09 | HTTP server handler tests with mock upstream |
| TS-10 | CLI tests via `exec.Command` |
| TS-11 | Build HLS master playlist with 3+ variants and audio tracks; verify output is re-parseable by the HLS parser |
| TS-12 | Build DASH MPD with video + audio AdaptationSets; verify output is re-parseable by the DASH parser |
| TS-13 | Build → filter round trip: build a manifest, then filter the result by codec and resolution |
| TS-14 | Builder returns `ErrInvalidVariant` when `URI` or `Bandwidth` is missing from any variant |
| TS-15 | Builder returns `ErrEmptyVariantList` when `Build()` is called with no variants added |
| TS-16 | HLS builder returns `ErrOrphanedGroupID` when `AudioGroupID` references a non-existent group |
| TS-17 | DASH builder returns `ErrInvalidLanguageTag` for a malformed BCP-47 `lang` value |
| TS-18 | HLS builder with `#EXT-X-I-FRAME-STREAM-INF` entries produces output accepted by hls.js |
| TS-19 | DASH builder with `SegmentTemplate` produces output accepted by shaka-player |
| TS-20 | DASH builder with `SegmentBase` produces output accepted by dash.js |
| TS-21 | Benchmark: `Build()` with 100 variants completes in under 2ms |

---

## 8. Real-World Use Case

### 8.1 VOD Manifest Transformation (Vieon-style)

A real production use case driving several of the requirements above. Given a Bento4-generated DASH MPD:

```
Input:  Multi-codec MPD (AVC1 + HEVC video sets, single Tajik audio)
```

Desired output:

```
Output: Filtered + transformed MPD with:
  1. Only HEVC (h265) video, max 1280×720
  2. <BaseURL> on every Representation rewritten to absolute CDN URL
  3. Original audio track kept (lang="tg") with name="Tiếng Gốc"
  4. New dubbed audio AdaptationSet injected (lang="tm", name="Thuyết Minh",
     different BaseURL from a separate encode)
  5. Subtitle AdaptationSet added (contentType="text", mimeType="text/vtt",
     lang="vi", <Role value="subtitle"/>, BaseURL pointing to .vtt file)
```

This requires all of: T-07 (BaseURL rewrite), M-01–M-06 (inject + name + subtitle + Role),
and BD-13–BD-16 (name, BaseURL, AudioChannelConfiguration, Role on builder side).

### 8.2 Sample Manifests to Support (Test Fixtures)

- Bento4 `mp4-dash.py` output (mixed h264/h265, `<BaseURL>` + `<SegmentBase>` per representation)
- Shaka Packager HLS output
- AWS MediaConvert HLS output
- Azure Media Services DASH output
- Standard MPEG-DASH ISOFF-on-demand profile
- Live DASH profile with `SegmentTemplate`
- Vieon-style VOD MPD: HEVC filtered, CDN BaseURL rewrite, injected dubbed audio, VTT subtitle track

---

## 9. Phased Delivery

### Phase 1 — Core Library (MVP)

- HLS parser + writer
- DASH parser + writer
- Filters: codec, max/min resolution, max/min bandwidth
- Transformers: absolute URIs, CDN rewrite, auth token
- **HLS Master Playlist builder** (video variants + audio tracks; B-01 – B-06, BH-01 – BH-09)
- **DASH MPD builder** with `SegmentTemplate` and `SegmentBase` (BD-01 – BD-11)
- Builder errors: `ErrEmptyVariantList`, `ErrInvalidVariant`, `ErrOrphanedGroupID`
- Unit tests + sample fixtures

### Phase 2 — Extended Filters + Server

- Filters: frame rate, audio language, MIME type, custom fn
- Transformer: custom fn
- HTTP proxy server (GET /filter + POST /build)
- Full error handling (`ErrNoVariantsRemain`, `ErrInvalidLanguageTag`, etc.)
- HLS builder: I-Frame streams (BH-07), subtitle tracks (BH-05 – BH-06)
- DASH builder: BCP-47 language validation (BD-12)

### Phase 3 — CLI + Polish

- CLI tool
- Benchmarks
- README with full examples
- GitHub Actions CI (test + lint + coverage)

---

## 10. Repo Structure

```
manifestor/
├── README.md
├── go.mod
│
├── hls/
│   ├── parser.go        # parse m3u8 → MasterPlaylist struct
│   ├── writer.go        # MasterPlaylist → m3u8 string
│   ├── builder.go       # MasterBuilder — build HLS master from scratch
│   ├── filter.go        # VariantFilter / VariantTransformer types
│   └── options.go       # WithCodec, WithResolution, WithCDN, etc.
│
├── dash/
│   ├── parser.go        # parse MPD XML → MPD struct
│   ├── writer.go        # MPD struct → XML string
│   ├── builder.go       # MPDBuilder — build DASH MPD from scratch
│   ├── filter.go        # RepresentationFilter / Transformer types
│   └── options.go       # same option pattern
│
├── manifest/
│   └── manifest.go      # unified Auto() that detects HLS vs DASH
│
├── server/
│   └── server.go        # optional HTTP proxy server
│
└── cmd/
    └── manifestor/
        └── main.go      # CLI tool
```

---

*Document version: 1.2 — April 2026 (added DASH BaseURL/inject/subtitle/name requirements from real VOD use case)*
