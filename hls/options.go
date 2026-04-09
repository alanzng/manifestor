package hls

import (
	manifestor "github.com/alanzng/manifestor"
)

// Option configures the behaviour of Filter().
type Option func(*filterConfig)

type filterConfig struct {
	codec             manifestor.Codec
	maxWidth          int
	maxHeight         int
	minWidth          int
	minHeight         int
	exactWidth        int
	exactHeight       int
	maxBandwidth      int
	minBandwidth      int
	maxFrameRate      float64
	audioLanguage     string
	mimeType          manifestor.MimeType
	cdnBaseURL        string
	absoluteOrigin    string
	authToken         string
	injectVariants    []VariantParams
	injectAudioTracks []AudioTrackParams
	injectSubtitles   []SubtitleTrackParams
	subtitleGroupID   string
	customFilter      func(*Variant) bool
	customTransform   func(*Variant)
}

// WithCodec keeps only variants whose Codecs field matches the given codec family.
func WithCodec(codec manifestor.Codec) Option {
	return func(c *filterConfig) { c.codec = codec }
}

// WithMaxResolution excludes variants wider or taller than r.
func WithMaxResolution(r manifestor.Resolution) Option {
	return func(c *filterConfig) { c.maxWidth = r.Width; c.maxHeight = r.Height }
}

// WithMinResolution excludes variants smaller than r.
func WithMinResolution(r manifestor.Resolution) Option {
	return func(c *filterConfig) { c.minWidth = r.Width; c.minHeight = r.Height }
}

// WithExactResolution keeps only variants with exactly r.
func WithExactResolution(r manifestor.Resolution) Option {
	return func(c *filterConfig) { c.exactWidth = r.Width; c.exactHeight = r.Height }
}

// WithMaxBandwidth excludes variants above bps bits/s.
func WithMaxBandwidth(bps int) Option {
	return func(c *filterConfig) { c.maxBandwidth = bps }
}

// WithMinBandwidth excludes variants below bps bits/s.
func WithMinBandwidth(bps int) Option {
	return func(c *filterConfig) { c.minBandwidth = bps }
}

// WithMaxFrameRate excludes variants with a frame rate above fps.
func WithMaxFrameRate(fps float64) Option {
	return func(c *filterConfig) { c.maxFrameRate = fps }
}

// WithAudioLanguage keeps only audio tracks whose LANGUAGE attribute matches lang (BCP-47).
func WithAudioLanguage(lang string) Option {
	return func(c *filterConfig) { c.audioLanguage = lang }
}

// WithMimeType keeps only variants matching the given MIME type.
func WithMimeType(mime manifestor.MimeType) Option {
	return func(c *filterConfig) { c.mimeType = mime }
}

// WithCDNBaseURL rewrites all variant URIs to use base as the CDN origin.
func WithCDNBaseURL(base string) Option {
	return func(c *filterConfig) { c.cdnBaseURL = base }
}

// WithAbsoluteURIs resolves relative URIs to absolute using origin as the base URL.
func WithAbsoluteURIs(origin string) Option {
	return func(c *filterConfig) { c.absoluteOrigin = origin }
}

// WithAuthToken appends token as a query string parameter to all URIs.
func WithAuthToken(token string) Option {
	return func(c *filterConfig) { c.authToken = token }
}

// WithCustomFilter applies a user-defined filter function to each variant.
// Variants for which fn returns false are removed.
func WithCustomFilter(fn func(*Variant) bool) Option {
	return func(c *filterConfig) { c.customFilter = fn }
}

// WithCustomTransformer applies a user-defined transformer to each surviving variant.
func WithCustomTransformer(fn func(*Variant)) Option {
	return func(c *filterConfig) { c.customTransform = fn }
}

// WithInjectVariant appends a variant stream to the playlist after filtering.
func WithInjectVariant(p VariantParams) Option {
	return func(c *filterConfig) { c.injectVariants = append(c.injectVariants, p) }
}

// WithInjectAudioTrack appends an audio media track to the playlist after filtering.
func WithInjectAudioTrack(p AudioTrackParams) Option {
	return func(c *filterConfig) { c.injectAudioTracks = append(c.injectAudioTracks, p) }
}

// WithInjectSubtitle appends a subtitle media track to the playlist after filtering.
func WithInjectSubtitle(p SubtitleTrackParams) Option {
	return func(c *filterConfig) { c.injectSubtitles = append(c.injectSubtitles, p) }
}

// WithVariantSubtitleGroup sets the SUBTITLES group ID on all surviving variant streams.
// Use this together with WithInjectSubtitle to wire variants to a subtitle group.
func WithVariantSubtitleGroup(groupID string) Option {
	return func(c *filterConfig) { c.subtitleGroupID = groupID }
}
