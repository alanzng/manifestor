package dash

import (
	"strings"

	manifestor "github.com/datvietvac-techhub/manifestor"
)

// Option configures the behaviour of Filter().
type Option func(*filterConfig)

type filterConfig struct {
	codec            manifestor.Codec
	maxWidth         int
	maxHeight        int
	minWidth         int
	minHeight        int
	exactWidth       int
	exactHeight      int
	maxBandwidth     int
	minBandwidth     int
	maxFrameRate     float64
	audioLanguage    string
	mimeType         manifestor.MimeType
	cdnBaseURL       string
	absoluteOrigin   string
	authToken        string
	injectSets       []AdaptationSetParams
	customFilter     func(*Representation) bool
	customTransform  func(*Representation)
	uriSigner        func(string) string
	clearAudio       bool
	audioLabelByLang map[string]string
}

// WithCodec keeps only representations whose Codecs field matches the given codec family.
func WithCodec(codec manifestor.Codec) Option {
	return func(c *filterConfig) { c.codec = codec }
}

// WithMaxResolution excludes representations wider or taller than r.
func WithMaxResolution(r manifestor.Resolution) Option {
	return func(c *filterConfig) { c.maxWidth = r.Width; c.maxHeight = r.Height }
}

// WithMinResolution excludes representations smaller than r.
func WithMinResolution(r manifestor.Resolution) Option {
	return func(c *filterConfig) { c.minWidth = r.Width; c.minHeight = r.Height }
}

// WithExactResolution keeps only representations with exactly r.
func WithExactResolution(r manifestor.Resolution) Option {
	return func(c *filterConfig) { c.exactWidth = r.Width; c.exactHeight = r.Height }
}

// WithMaxBandwidth excludes representations above bps bits/s.
func WithMaxBandwidth(bps int) Option {
	return func(c *filterConfig) { c.maxBandwidth = bps }
}

// WithMinBandwidth excludes representations below bps bits/s.
func WithMinBandwidth(bps int) Option {
	return func(c *filterConfig) { c.minBandwidth = bps }
}

// WithMaxFrameRate excludes representations with a frame rate above fps.
func WithMaxFrameRate(fps float64) Option {
	return func(c *filterConfig) { c.maxFrameRate = fps }
}

// WithAudioLanguage keeps only audio AdaptationSets whose lang attribute matches lang (BCP-47).
func WithAudioLanguage(lang string) Option {
	return func(c *filterConfig) { c.audioLanguage = lang }
}

// WithMimeType keeps only representations matching the given MIME type.
func WithMimeType(mime manifestor.MimeType) Option {
	return func(c *filterConfig) { c.mimeType = mime }
}

// WithCDNBaseURL rewrites all representation URIs to use base as the CDN origin.
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

// WithCustomFilter applies a user-defined filter function to each representation.
// Representations for which fn returns false are removed.
func WithCustomFilter(fn func(*Representation) bool) Option {
	return func(c *filterConfig) { c.customFilter = fn }
}

// WithCustomTransformer applies a user-defined transformer to each surviving representation.
func WithCustomTransformer(fn func(*Representation)) Option {
	return func(c *filterConfig) { c.customTransform = fn }
}

// WithInjectAdaptationSet appends an AdaptationSet built from p to every Period
// after filtering. Useful for injecting subtitle or secondary audio tracks.
func WithInjectAdaptationSet(p AdaptationSetParams) Option {
	return func(c *filterConfig) { c.injectSets = append(c.injectSets, p) }
}

// WithURISigner applies fn to every absolute Representation BaseURL emitted
// by Filter, including BaseURLs in injected AdaptationSets. fn is only
// invoked for absolute URLs.
func WithURISigner(fn func(string) string) Option {
	return func(c *filterConfig) { c.uriSigner = fn }
}

// WithClearAudioTracks removes every audio AdaptationSet from each Period
// before inject options are applied. Use to replace origin audio entirely.
func WithClearAudioTracks() Option {
	return func(c *filterConfig) { c.clearAudio = true }
}

// WithAudioLabelByLanguage rewrites the label (AdaptationSet.Name) of every
// ORIGIN audio AdaptationSet whose lang attribute matches one of the map keys
// (case-insensitive BCP-47 language tag). Video, text, and injected
// AdaptationSets are not affected. An entry with an empty-string value is
// ignored (original label preserved), which prevents accidentally blanking
// a name.
//
// Example:
//
//	dash.Filter(content, dash.WithAudioLabelByLanguage(map[string]string{
//	    "tg": "Tiếng gốc",
//	}))
func WithAudioLabelByLanguage(m map[string]string) Option {
	// Normalize keys to lowercase at construction time so lookups during
	// Filter do not need to re-lowercase. Empty values are dropped here so
	// the filter never has to guard against them.
	normalized := make(map[string]string, len(m))
	for k, v := range m {
		if v == "" {
			continue
		}
		normalized[strings.ToLower(k)] = v
	}
	return func(c *filterConfig) { c.audioLabelByLang = normalized }
}
