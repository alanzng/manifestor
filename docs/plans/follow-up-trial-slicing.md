# Follow-ups: Trial Slicing Primitives

These were deliberately deferred from the initial trial slicing PR. Each is self-contained and suitable for an outside contributor.

## 1. Typed `MediaPlaylist` parser

`hls.Parse` returns `*MasterPlaylist`. There is no parsed representation of an HLS media playlist. Add `hls.MediaPlaylist` (segments, init, target duration, playlist type) and `hls.ParseMediaPlaylist` / `hls.SerializeMediaPlaylist`. `SliceMediaPlaylist` can stay as a string-level fast path; the typed API is for callers that need structured access.

## 2. `manifest/` unified-API integration

The `manifest` package auto-detects HLS vs. DASH. Add `manifest.Slice(raw, opts...)` and option constructors `WithTrialSegments(n)`, `WithTrialDuration(seconds)` so consumers using the unified API don't have to branch on format.

## 3. `manifestor slice` CLI subcommand

`cmd/manifestor` exposes `filter`, `build`, `serve`. Add `slice` taking `--input`, `--output`, `--n` or `--duration` flags and dispatching to the right primitive based on file extension.

## 4. `/slice` HTTP endpoint in `server/`

Mirror the CLI: `POST /slice` with a JSON body `{url, n}` or `{url, duration}` returning the sliced manifest.

## 5. Query-string-aware `AddTrialSuffix`

`path.Ext("audio.m3u8?token=abc")` returns `.m3u8?token=abc`, so the suffix lands after the query. Add a query-aware variant or option that splits on `?` first.

## 6. Live / event playlist support

`SliceMediaPlaylist` always re-emits `#EXT-X-ENDLIST`, which is wrong for `EXT-X-PLAYLIST-TYPE:EVENT` and live playlists. Decide on behavior (reject? preserve event semantics?) and implement.
