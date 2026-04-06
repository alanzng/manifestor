package manifest

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
