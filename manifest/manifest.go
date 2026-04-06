// Package manifest provides a unified, format-agnostic API for parsing,
// filtering, transforming, and building HLS and DASH manifests.
//
// Format detection is performed automatically from content or Content-Type headers.
package manifest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/alanzng/manifestor/dash"
	"github.com/alanzng/manifestor/hls"
)

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

// Detect returns the Format of the given manifest content or ErrInvalidFormat
// if the content cannot be identified.
func Detect(content string) (Format, error) {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "#EXTM3U") {
		return FormatHLS, nil
	}
	if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<MPD") {
		return FormatDASH, nil
	}
	return 0, ErrInvalidFormat
}

// Filter auto-detects the manifest format from content and applies opts.
// It returns the filtered manifest serialized in its original format.
//
// Filter is safe for concurrent use.
func Filter(content string, opts ...Option) (string, error) {
	format, err := Detect(content)
	if err != nil {
		return "", err
	}
	switch format {
	case FormatHLS:
		return hls.Filter(content, toHLSOpts(opts)...)
	case FormatDASH:
		return dash.Filter(content, toDASHOpts(opts)...)
	}
	return "", ErrInvalidFormat
}

// FilterFromURL fetches the manifest at rawURL, auto-detects its format,
// applies opts, and returns the filtered manifest.
func FilterFromURL(rawURL string, opts ...Option) (string, error) {
	resp, err := http.Get(rawURL) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d from %s", ErrFetchFailed, resp.StatusCode, rawURL)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: reading body: %v", ErrFetchFailed, err)
	}
	return Filter(string(b), opts...)
}

// FilterFromFile reads the manifest at path, auto-detects its format, applies
// opts, and returns the filtered manifest.
func FilterFromFile(path string, opts ...Option) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	return Filter(string(b), opts...)
}

// Build builds a manifest of the given format using opts and returns the
// serialized output. For HLS, it constructs a MasterPlaylist via hls.MasterBuilder.
// For DASH, it constructs an MPD via dash.MPDBuilder.
//
// At minimum you must supply at least one format-specific option that adds a
// variant/representation; otherwise ErrEmptyVariantList is returned from the
// underlying builder.
func Build(format Format, opts ...Option) (string, error) {
	switch format {
	case FormatHLS:
		b := hls.NewMasterBuilder()
		for _, o := range opts {
			if o.hlsOption() {
				applyHLSBuildOption(b, o)
			}
		}
		return b.Build()
	case FormatDASH:
		cfg := extractDASHConfig(opts)
		b := dash.NewMPDBuilder(cfg)
		for _, o := range opts {
			if o.dashOption() {
				applyDASHBuildOption(b, o)
			}
		}
		return b.Build()
	}
	return "", ErrInvalidFormat
}

// toHLSOpts converts []Option to []hls.Option by mapping each unified option.
func toHLSOpts(opts []Option) []hls.Option {
	out := make([]hls.Option, 0, len(opts))
	for _, o := range opts {
		if !o.hlsOption() {
			continue
		}
		switch v := o.(type) {
		case codecOption:
			out = append(out, hls.WithCodec(v.codec))
		case maxResOption:
			out = append(out, hls.WithMaxResolution(v.w, v.h))
		case minResOption:
			out = append(out, hls.WithMinResolution(v.w, v.h))
		case exactResOption:
			out = append(out, hls.WithExactResolution(v.w, v.h))
		case maxBwOption:
			out = append(out, hls.WithMaxBandwidth(v.bps))
		case minBwOption:
			out = append(out, hls.WithMinBandwidth(v.bps))
		case maxFPSOption:
			out = append(out, hls.WithMaxFrameRate(v.fps))
		case audioLangOption:
			out = append(out, hls.WithAudioLanguage(v.lang))
		case cdnOption:
			out = append(out, hls.WithCDNBaseURL(v.base))
		case absoluteURIsOption:
			out = append(out, hls.WithAbsoluteURIs(v.origin))
		case authTokenOption:
			out = append(out, hls.WithAuthToken(v.token))
		}
	}
	return out
}

// toDASHOpts converts []Option to []dash.Option.
func toDASHOpts(opts []Option) []dash.Option {
	out := make([]dash.Option, 0, len(opts))
	for _, o := range opts {
		if !o.dashOption() {
			continue
		}
		switch v := o.(type) {
		case codecOption:
			out = append(out, dash.WithCodec(v.codec))
		case maxResOption:
			out = append(out, dash.WithMaxResolution(v.w, v.h))
		case minResOption:
			out = append(out, dash.WithMinResolution(v.w, v.h))
		case exactResOption:
			out = append(out, dash.WithExactResolution(v.w, v.h))
		case maxBwOption:
			out = append(out, dash.WithMaxBandwidth(v.bps))
		case minBwOption:
			out = append(out, dash.WithMinBandwidth(v.bps))
		case maxFPSOption:
			out = append(out, dash.WithMaxFrameRate(v.fps))
		case audioLangOption:
			out = append(out, dash.WithAudioLanguage(v.lang))
		case mimeTypeOption:
			out = append(out, dash.WithMimeType(v.mime))
		case cdnOption:
			out = append(out, dash.WithCDNBaseURL(v.base))
		case absoluteURIsOption:
			out = append(out, dash.WithAbsoluteURIs(v.origin))
		case authTokenOption:
			out = append(out, dash.WithAuthToken(v.token))
		}
	}
	return out
}

// extractDASHConfig pulls top-level MPD configuration from opts.
func extractDASHConfig(opts []Option) dash.MPDConfig {
	cfg := dash.MPDConfig{}
	for _, o := range opts {
		switch v := o.(type) {
		case dashConfigOption:
			cfg = v.cfg
		}
	}
	return cfg
}

// applyHLSBuildOption applies a build-time option to an hls.MasterBuilder.
func applyHLSBuildOption(b *hls.MasterBuilder, o Option) {
	switch v := o.(type) {
	case hlsVersionOption:
		b.SetVersion(v.version)
	case hlsVariantOption:
		b.AddVariant(v.params)
	case hlsAudioTrackOption:
		b.AddAudioTrack(v.params)
	case hlsSubtitleTrackOption:
		b.AddSubtitleTrack(v.params)
	case hlsIFrameOption:
		b.AddIFrameStream(v.params)
	}
}

// applyDASHBuildOption applies a build-time option to a dash.MPDBuilder.
func applyDASHBuildOption(b *dash.MPDBuilder, o Option) {
	switch v := o.(type) {
	case dashAdaptationSetOption:
		b.AddAdaptationSet(v.params)
	}
}
