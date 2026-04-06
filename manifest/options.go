package manifest

import (
	"github.com/alanzng/manifestor/dash"
	"github.com/alanzng/manifestor/hls"
)

// sharedOption is embedded by all unified option types so they satisfy Option.
type sharedOption struct{}

func (sharedOption) hlsOption() bool  { return true }
func (sharedOption) dashOption() bool { return true }

// codecOption implements WithCodec.
type codecOption struct {
	sharedOption
	codec string
}

// WithCodec keeps only variants/representations matching the given codec family.
// Accepted values: "h264", "h265", "vp9", "av1".
func WithCodec(codec string) Option { return codecOption{codec: codec} }

// maxResOption implements WithMaxResolution.
type maxResOption struct {
	sharedOption
	w, h int
}

// WithMaxResolution excludes variants/representations wider or taller than w×h.
func WithMaxResolution(w, h int) Option { return maxResOption{w: w, h: h} }

// minResOption implements WithMinResolution.
type minResOption struct {
	sharedOption
	w, h int
}

// WithMinResolution excludes variants/representations smaller than w×h.
func WithMinResolution(w, h int) Option { return minResOption{w: w, h: h} }

// exactResOption implements WithExactResolution.
type exactResOption struct {
	sharedOption
	w, h int
}

// WithExactResolution keeps only variants/representations with exactly w×h.
func WithExactResolution(w, h int) Option { return exactResOption{w: w, h: h} }

// maxBwOption implements WithMaxBandwidth.
type maxBwOption struct {
	sharedOption
	bps int
}

// WithMaxBandwidth excludes variants/representations above bps bits/s.
func WithMaxBandwidth(bps int) Option { return maxBwOption{bps: bps} }

// minBwOption implements WithMinBandwidth.
type minBwOption struct {
	sharedOption
	bps int
}

// WithMinBandwidth excludes variants/representations below bps bits/s.
func WithMinBandwidth(bps int) Option { return minBwOption{bps: bps} }

// maxFPSOption implements WithMaxFrameRate.
type maxFPSOption struct {
	sharedOption
	fps float64
}

// WithMaxFrameRate excludes variants/representations with frame rate above fps.
func WithMaxFrameRate(fps float64) Option { return maxFPSOption{fps: fps} }

// audioLangOption implements WithAudioLanguage.
type audioLangOption struct {
	sharedOption
	lang string
}

// WithAudioLanguage keeps only audio tracks/AdaptationSets matching lang (BCP-47).
func WithAudioLanguage(lang string) Option { return audioLangOption{lang: lang} }

// mimeTypeOption implements WithMimeType.
type mimeTypeOption struct {
	sharedOption
	mime string
}

// WithMimeType keeps only representations matching the given MIME type.
func WithMimeType(mime string) Option { return mimeTypeOption{mime: mime} }

// cdnOption implements WithCDNBaseURL.
type cdnOption struct {
	sharedOption
	base string
}

// WithCDNBaseURL rewrites all variant/representation URIs to use base as the CDN origin.
func WithCDNBaseURL(base string) Option { return cdnOption{base: base} }

// absoluteURIsOption implements WithAbsoluteURIs.
type absoluteURIsOption struct {
	sharedOption
	origin string
}

// WithAbsoluteURIs resolves relative URIs to absolute using origin as the base URL.
func WithAbsoluteURIs(origin string) Option { return absoluteURIsOption{origin: origin} }

// authTokenOption implements WithAuthToken.
type authTokenOption struct {
	sharedOption
	token string
}

// WithAuthToken appends token as a query string parameter to all URIs.
func WithAuthToken(token string) Option { return authTokenOption{token: token} }

// ---- HLS build-only options ----

// hlsOnlyOption is embedded by HLS-only build options.
type hlsOnlyOption struct{}

func (hlsOnlyOption) hlsOption() bool  { return true }
func (hlsOnlyOption) dashOption() bool { return false }

// hlsVersionOption sets the HLS playlist version.
type hlsVersionOption struct {
	hlsOnlyOption
	version int
}

// WithHLSVersion sets the #EXT-X-VERSION value in the built HLS playlist.
func WithHLSVersion(v int) Option { return hlsVersionOption{version: v} }

// hlsVariantOption adds a variant stream to the HLS playlist.
type hlsVariantOption struct {
	hlsOnlyOption
	params hls.VariantParams
}

// WithHLSVariant adds a video variant stream to the built HLS Master Playlist.
func WithHLSVariant(p hls.VariantParams) Option { return hlsVariantOption{params: p} }

// hlsAudioTrackOption adds an audio track to the HLS playlist.
type hlsAudioTrackOption struct {
	hlsOnlyOption
	params hls.AudioTrackParams
}

// WithHLSAudioTrack adds an #EXT-X-MEDIA AUDIO entry to the built HLS playlist.
func WithHLSAudioTrack(p hls.AudioTrackParams) Option { return hlsAudioTrackOption{params: p} }

// hlsSubtitleTrackOption adds a subtitle track to the HLS playlist.
type hlsSubtitleTrackOption struct {
	hlsOnlyOption
	params hls.SubtitleTrackParams
}

// WithHLSSubtitleTrack adds an #EXT-X-MEDIA SUBTITLES entry to the built HLS playlist.
func WithHLSSubtitleTrack(p hls.SubtitleTrackParams) Option {
	return hlsSubtitleTrackOption{params: p}
}

// hlsIFrameOption adds an I-frame stream to the HLS playlist.
type hlsIFrameOption struct {
	hlsOnlyOption
	params hls.IFrameParams
}

// WithHLSIFrameStream adds an #EXT-X-I-FRAME-STREAM-INF entry to the built HLS playlist.
func WithHLSIFrameStream(p hls.IFrameParams) Option { return hlsIFrameOption{params: p} }

// ---- DASH build-only options ----

// dashOnlyOption is embedded by DASH-only build options.
type dashOnlyOption struct{}

func (dashOnlyOption) hlsOption() bool  { return false }
func (dashOnlyOption) dashOption() bool { return true }

// dashConfigOption sets the top-level MPD configuration.
type dashConfigOption struct {
	dashOnlyOption
	cfg dash.MPDConfig
}

// WithDASHConfig sets the top-level MPD configuration for the built DASH manifest.
func WithDASHConfig(cfg dash.MPDConfig) Option { return dashConfigOption{cfg: cfg} }

// dashAdaptationSetOption adds an AdaptationSet to the DASH MPD.
type dashAdaptationSetOption struct {
	dashOnlyOption
	params dash.AdaptationSetParams
}

// WithDASHAdaptationSet adds an AdaptationSet to the built DASH MPD.
func WithDASHAdaptationSet(p dash.AdaptationSetParams) Option {
	return dashAdaptationSetOption{params: p}
}
