// Package manifest provides a unified, format-agnostic API for parsing,
// filtering, transforming, and building HLS and DASH manifests.
//
// Format detection is performed automatically from content or Content-Type headers.
package manifest

import "errors"

// Format identifies the manifest format.
type Format int

const (
	// FormatHLS identifies an HLS manifest.
	FormatHLS Format = iota
	// FormatDASH identifies a DASH MPD manifest.
	FormatDASH
)

// Sentinel errors common to both formats.
var (
	// ErrInvalidFormat is returned when the content is neither valid HLS nor DASH.
	ErrInvalidFormat = errors.New("manifest: content is neither valid HLS nor DASH")

	// ErrFetchFailed is returned when fetching an upstream URL fails.
	ErrFetchFailed = errors.New("manifest: failed to fetch upstream manifest")
)

// Option configures the behaviour of Filter and Build.
// Use the With* constructors from this package to create options.
type Option interface {
	hlsOption() bool
	dashOption() bool
}

// Filter auto-detects the manifest format from content and applies opts.
// It returns the filtered manifest serialized in its original format.
//
// Filter is safe for concurrent use.
func Filter(content string, opts ...Option) (string, error) {
	// TODO: implement
	panic("not implemented")
}

// FilterFromURL fetches the manifest at url, auto-detects its format, applies
// opts, and returns the filtered manifest.
func FilterFromURL(url string, opts ...Option) (string, error) {
	// TODO: implement
	panic("not implemented")
}

// FilterFromFile reads the manifest at path, auto-detects its format, applies
// opts, and returns the filtered manifest.
func FilterFromFile(path string, opts ...Option) (string, error) {
	// TODO: implement
	panic("not implemented")
}

// Build builds a manifest of the given format using opts and returns the
// serialized output.
func Build(format Format, opts ...Option) (string, error) {
	// TODO: implement
	panic("not implemented")
}

// Detect returns the Format of the given manifest content or ErrInvalidFormat
// if the content cannot be identified.
func Detect(content string) (Format, error) {
	// TODO: implement
	panic("not implemented")
}
