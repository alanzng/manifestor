package dash

// Option configures the behaviour of Filter().
type Option func(*filterConfig)

type filterConfig struct {
	codec           string
	maxWidth        int
	maxHeight       int
	minWidth        int
	minHeight       int
	exactWidth      int
	exactHeight     int
	maxBandwidth    int
	minBandwidth    int
	maxFrameRate    float64
	audioLanguage   string
	mimeType        string
	cdnBaseURL      string
	absoluteOrigin  string
	authToken       string
	injectSets      []AdaptationSetParams
	customFilter    func(*Representation) bool
	customTransform func(*Representation)
}

// WithCodec keeps only representations whose Codecs field matches the given codec family.
// Accepted values: "h264", "h265", "vp9", "av1".
func WithCodec(codec string) Option {
	return func(c *filterConfig) { c.codec = codec }
}

// WithMaxResolution excludes representations wider or taller than w×h.
func WithMaxResolution(w, h int) Option {
	return func(c *filterConfig) { c.maxWidth = w; c.maxHeight = h }
}

// WithMinResolution excludes representations smaller than w×h.
func WithMinResolution(w, h int) Option {
	return func(c *filterConfig) { c.minWidth = w; c.minHeight = h }
}

// WithExactResolution keeps only representations with exactly w×h.
func WithExactResolution(w, h int) Option {
	return func(c *filterConfig) { c.exactWidth = w; c.exactHeight = h }
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
func WithMimeType(mime string) Option {
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
