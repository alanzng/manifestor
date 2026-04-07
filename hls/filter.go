package hls

import (
	"net/url"
	"strings"
)

// Filter parses content as an HLS Master Playlist, applies opts, and returns
// the filtered and transformed playlist serialized back to m3u8 format.
//
// Filters compose with AND logic: a variant must pass every active filter to
// survive. Transformers are applied after filtering, only to surviving variants.
//
// Filter is safe for concurrent use.
func Filter(content string, opts ...Option) (string, error) {
	p, err := Parse(content)
	if err != nil {
		return "", err
	}

	cfg := &filterConfig{}
	for _, o := range opts {
		o(cfg)
	}

	// Deep-copy variants so the parsed struct is never mutated (thread safety).
	variants := make([]Variant, len(p.Variants))
	copy(variants, p.Variants)

	// Apply filters (AND logic).
	filtered := variants[:0]
	for i := range variants {
		if variantPasses(&variants[i], cfg) {
			filtered = append(filtered, variants[i])
		}
	}
	if len(filtered) == 0 {
		return "", ErrNoVariantsRemain
	}

	// Apply transformers to surviving variants.
	for i := range filtered {
		applyTransformers(&filtered[i], cfg)
		if cfg.subtitleGroupID != "" {
			filtered[i].SubtitleGroupID = cfg.subtitleGroupID
		}
	}

	// Filter I-frame streams with the same codec/resolution/bandwidth rules (F-14).
	iframes := make([]IFrameStream, len(p.IFrames))
	copy(iframes, p.IFrames)
	filteredIFrames := iframes[:0]
	for i := range iframes {
		if iframePasses(&iframes[i], cfg) {
			iframes[i].URI = rewriteURI(iframes[i].URI, cfg)
			filteredIFrames = append(filteredIFrames, iframes[i])
		}
	}

	// Filter audio tracks by language (F-08, F-13) and rewrite their URIs.
	audioTracks := filterAudioTracks(p.AudioTracks, cfg)
	for i := range audioTracks {
		audioTracks[i].URI = rewriteURI(audioTracks[i].URI, cfg)
	}

	// Inject additional variants, audio tracks, and subtitles.
	for _, vp := range cfg.injectVariants {
		filtered = append(filtered, Variant(vp))
	}
	for _, ap := range cfg.injectAudioTracks {
		audioTracks = append(audioTracks, MediaTrack{
			Type:       "AUDIO",
			GroupID:    ap.GroupID,
			Name:       ap.Name,
			Language:   ap.Language,
			URI:        ap.URI,
			Default:    ap.Default,
			AutoSelect: ap.AutoSelect,
			Forced:     ap.Forced,
		})
	}
	subtitles := p.Subtitles
	for _, sp := range cfg.injectSubtitles {
		subtitles = append(subtitles, MediaTrack{
			Type:     "SUBTITLES",
			GroupID:  sp.GroupID,
			Name:     sp.Name,
			Language: sp.Language,
			URI:      sp.URI,
			Default:  sp.Default,
			Forced:   sp.Forced,
		})
	}

	out := &MasterPlaylist{
		Version:     p.Version,
		Variants:    filtered,
		AudioTracks: audioTracks,
		Subtitles:   subtitles,
		IFrames:     filteredIFrames,
		Raw:         p.Raw,
	}
	return Serialize(out)
}

// variantPasses reports whether v satisfies all active filters in cfg.
func variantPasses(v *Variant, cfg *filterConfig) bool {
	if cfg.codec != "" && !matchesCodec(v.Codecs, cfg.codec) {
		return false
	}
	if cfg.maxWidth > 0 && v.Width > cfg.maxWidth {
		return false
	}
	if cfg.maxHeight > 0 && v.Height > cfg.maxHeight {
		return false
	}
	if cfg.minWidth > 0 && v.Width < cfg.minWidth {
		return false
	}
	if cfg.minHeight > 0 && v.Height < cfg.minHeight {
		return false
	}
	if cfg.exactWidth > 0 && v.Width != cfg.exactWidth {
		return false
	}
	if cfg.exactHeight > 0 && v.Height != cfg.exactHeight {
		return false
	}
	if cfg.maxBandwidth > 0 && v.Bandwidth > cfg.maxBandwidth {
		return false
	}
	if cfg.minBandwidth > 0 && v.Bandwidth < cfg.minBandwidth {
		return false
	}
	if cfg.maxFrameRate > 0 && v.FrameRate > cfg.maxFrameRate {
		return false
	}
	if cfg.customFilter != nil && !cfg.customFilter(v) {
		return false
	}
	return true
}

// iframePasses applies codec, resolution, and bandwidth filters to an I-frame stream (F-14).
func iframePasses(f *IFrameStream, cfg *filterConfig) bool {
	if cfg.codec != "" && !matchesCodec(f.Codecs, cfg.codec) {
		return false
	}
	if cfg.maxWidth > 0 && f.Width > cfg.maxWidth {
		return false
	}
	if cfg.maxHeight > 0 && f.Height > cfg.maxHeight {
		return false
	}
	if cfg.minWidth > 0 && f.Width < cfg.minWidth {
		return false
	}
	if cfg.minHeight > 0 && f.Height < cfg.minHeight {
		return false
	}
	if cfg.exactWidth > 0 && f.Width != cfg.exactWidth {
		return false
	}
	if cfg.exactHeight > 0 && f.Height != cfg.exactHeight {
		return false
	}
	if cfg.maxBandwidth > 0 && f.Bandwidth > cfg.maxBandwidth {
		return false
	}
	if cfg.minBandwidth > 0 && f.Bandwidth < cfg.minBandwidth {
		return false
	}
	return true
}

// filterAudioTracks returns the audio tracks that pass the language filter.
// If no language filter is set, all tracks are preserved (F-13).
func filterAudioTracks(tracks []MediaTrack, cfg *filterConfig) []MediaTrack {
	if cfg.audioLanguage == "" {
		return tracks
	}
	out := tracks[:0:0]
	for _, t := range tracks {
		if strings.EqualFold(t.Language, cfg.audioLanguage) {
			out = append(out, t)
		}
	}
	return out
}

// applyTransformers rewrites the URI of a surviving variant according to cfg.
// Order: absolute URIs → CDN rewrite → auth token (T-01 → T-02 → T-03).
// Custom transformer runs last (T-04, T-06).
func applyTransformers(v *Variant, cfg *filterConfig) {
	v.URI = rewriteURI(v.URI, cfg)
	if cfg.customTransform != nil {
		cfg.customTransform(v)
	}
}

// rewriteURI applies the active URI transformers to a single URI string.
// It parses the URI only once for efficiency.
func rewriteURI(uri string, cfg *filterConfig) string {
	if cfg.absoluteOrigin == "" && cfg.cdnBaseURL == "" && cfg.authToken == "" {
		return uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	if cfg.absoluteOrigin != "" && !u.IsAbs() {
		base, err := url.Parse(cfg.absoluteOrigin)
		if err == nil {
			u = base.ResolveReference(u)
		}
	}
	if cfg.cdnBaseURL != "" && u.IsAbs() {
		c, err := url.Parse(cfg.cdnBaseURL)
		if err == nil {
			u.Scheme = c.Scheme
			u.Host = c.Host
		}
	}
	if cfg.authToken != "" {
		q := u.Query()
		q.Set("token", cfg.authToken)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// makeAbsolute resolves a relative URI against origin (T-01).
func makeAbsolute(uri, origin string) string {
	u, err := url.Parse(uri)
	if err != nil || u.IsAbs() {
		return uri
	}
	base, err := url.Parse(origin)
	if err != nil {
		return uri
	}
	return base.ResolveReference(u).String()
}

// rewriteCDN replaces the scheme and host of uri with those from cdn (T-02).
// The original path, query, and fragment are preserved.
func rewriteCDN(uri, cdn string) string {
	u, err := url.Parse(uri)
	if err != nil || !u.IsAbs() {
		return uri
	}
	c, err := url.Parse(cdn)
	if err != nil {
		return uri
	}
	u.Scheme = c.Scheme
	u.Host = c.Host
	return u.String()
}

// appendToken appends token as a query parameter to uri (T-03).
// Uses Set (not Add) so repeated calls are idempotent.
func appendToken(uri, token string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri + "?token=" + token
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}

// matchesCodec reports whether the HLS CODECS attribute value contains a codec
// from the requested family. Matching is case-insensitive.
//
// Families: "h264" (avc1/avc3), "h265" (hvc1/hev1), "vp9" (vp09), "av1" (av01).
func matchesCodec(codecsField, want string) bool {
	for _, c := range strings.Split(codecsField, ",") {
		c = strings.ToLower(strings.TrimSpace(c))
		switch want {
		case "h264":
			if strings.HasPrefix(c, "avc1.") || strings.HasPrefix(c, "avc3.") {
				return true
			}
		case "h265":
			if strings.HasPrefix(c, "hvc1.") || strings.HasPrefix(c, "hev1.") {
				return true
			}
		case "vp9":
			if strings.HasPrefix(c, "vp09.") || c == "vp9" {
				return true
			}
		case "av1":
			if strings.HasPrefix(c, "av01.") {
				return true
			}
		}
	}
	return false
}
