package manifest

// URISigner rewrites a single absolute URL and returns the replacement.
// Implementations MUST leave inputs they don't recognise (wrong host, etc.)
// unchanged. The signer is only invoked for absolute URLs — relative URIs
// are passed through untouched.
type URISigner func(absoluteURL string) string

type uriSignerOption struct {
	sharedOption
	fn URISigner
}

// WithURISigner applies fn to every absolute URI produced by Filter:
// HLS variant URIs, audio/subtitle media URIs, I-frame URIs, and DASH
// representation BaseURLs (including injected ones). Pass nil to disable.
func WithURISigner(fn URISigner) Option { return uriSignerOption{fn: fn} }
