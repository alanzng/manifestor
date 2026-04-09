package dash

import (
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Filter parses content as a DASH MPD, applies opts, and returns the filtered
// and transformed MPD serialized back to XML format.
//
// Filters compose with AND logic: a representation must pass every active
// filter to survive. Audio AdaptationSets are filtered by language separately.
// Empty AdaptationSets (all representations removed) are dropped. If all
// representations across all periods are removed, ErrNoVariantsRemain is returned.
//
// Filter is safe for concurrent use.
func Filter(content string, opts ...Option) (string, error) {
	m, err := Parse(content)
	if err != nil {
		return "", err
	}

	cfg := &filterConfig{}
	for _, o := range opts {
		o(cfg)
	}

	// Deep-copy periods so the parsed struct is never mutated.
	periods := make([]Period, len(m.Periods))
	copy(periods, m.Periods)

	totalReps := 0
	for pi := range periods {
		as := make([]AdaptationSet, len(periods[pi].AdaptationSets))
		copy(as, periods[pi].AdaptationSets)

		surviving := as[:0]
		for ai := range as {
			filtered := filterAdaptationSet(&as[ai], cfg)
			if filtered != nil {
				surviving = append(surviving, *filtered)
			}
		}
		periods[pi].AdaptationSets = surviving
		for _, a := range surviving {
			totalReps += len(a.Representations)
		}
	}

	if totalReps == 0 {
		return "", ErrNoVariantsRemain
	}

	// Inject additional AdaptationSets into every Period.
	if len(cfg.injectSets) > 0 {
		for pi := range periods {
			for _, sp := range cfg.injectSets {
				periods[pi].AdaptationSets = append(periods[pi].AdaptationSets, convertAdaptationSetParams(sp))
			}
		}
	}

	out := &MPD{
		Profile:         m.Profile,
		Duration:        m.Duration,
		MinBufferTime:   m.MinBufferTime,
		MinUpdatePeriod: m.MinUpdatePeriod,
		Periods:         periods,
		Raw:             m.Raw,
	}
	return Serialize(out)
}

// filterAdaptationSet returns a copy of as with only representations that pass
// all active filters, or nil if no representations survive or the set is
// excluded by the language filter.
func filterAdaptationSet(as *AdaptationSet, cfg *filterConfig) *AdaptationSet {
	// Language filter applies only to audio AdaptationSets.
	if cfg.audioLanguage != "" && isAudioAdaptationSet(as) {
		if !strings.EqualFold(as.Lang, cfg.audioLanguage) {
			return nil
		}
	}

	reps := make([]Representation, len(as.Representations))
	copy(reps, as.Representations)

	surviving := reps[:0]
	for i := range reps {
		if representationPasses(&reps[i], as, cfg) {
			r := reps[i]
			applyTransformers(&r, cfg)
			surviving = append(surviving, r)
		}
	}

	if len(surviving) == 0 {
		return nil
	}

	result := *as
	result.Representations = surviving
	return &result
}

// isAudioAdaptationSet reports whether as is an audio set by ContentType or MimeType.
func isAudioAdaptationSet(as *AdaptationSet) bool {
	if strings.EqualFold(as.ContentType, "audio") {
		return true
	}
	return strings.HasPrefix(strings.ToLower(as.MimeType), "audio/")
}

// isTextAdaptationSet reports whether as is a text/subtitle set by ContentType or MimeType.
func isTextAdaptationSet(as *AdaptationSet) bool {
	if strings.EqualFold(as.ContentType, "text") {
		return true
	}
	return strings.HasPrefix(strings.ToLower(as.MimeType), "text/")
}

// representationPasses reports whether r satisfies all active filters.
func representationPasses(r *Representation, as *AdaptationSet, cfg *filterConfig) bool {
	// Codec filter applies to video only; audio/text sets are unaffected.
	if cfg.codec != "" && !isAudioAdaptationSet(as) && !isTextAdaptationSet(as) {
		if !cfg.codec.MatchesCodec(r.Codecs) {
			return false
		}
	}

	// Resolve effective MimeType (may be inherited from AdaptationSet).
	mime := r.MimeType
	if mime == "" {
		mime = as.MimeType
	}
	if cfg.mimeType != "" && !strings.EqualFold(mime, string(cfg.mimeType)) {
		return false
	}

	if cfg.maxWidth > 0 && r.Width > cfg.maxWidth {
		return false
	}
	if cfg.maxHeight > 0 && r.Height > cfg.maxHeight {
		return false
	}
	if cfg.minWidth > 0 && r.Width < cfg.minWidth {
		return false
	}
	if cfg.minHeight > 0 && r.Height < cfg.minHeight {
		return false
	}
	if cfg.exactWidth > 0 && r.Width != cfg.exactWidth {
		return false
	}
	if cfg.exactHeight > 0 && r.Height != cfg.exactHeight {
		return false
	}
	if cfg.maxBandwidth > 0 && r.Bandwidth > cfg.maxBandwidth {
		return false
	}
	if cfg.minBandwidth > 0 && r.Bandwidth < cfg.minBandwidth {
		return false
	}
	if cfg.maxFrameRate > 0 {
		if fps := parseFrameRate(r.FrameRate); fps > cfg.maxFrameRate {
			return false
		}
	}
	if cfg.customFilter != nil && !cfg.customFilter(r) {
		return false
	}
	return true
}

// applyTransformers applies URI and custom transformers to a surviving representation.
// It parses the BaseURL only once for efficiency.
func applyTransformers(r *Representation, cfg *filterConfig) {
	if cfg.absoluteOrigin != "" || cfg.cdnBaseURL != "" || cfg.authToken != "" {
		uri := r.BaseURL
		u, err := url.Parse(uri)
		if err == nil {
			if cfg.absoluteOrigin != "" && !u.IsAbs() {
				origin := cfg.absoluteOrigin
				if !strings.HasSuffix(origin, "/") {
					origin += "/"
				}
				base, berr := url.Parse(origin)
				if berr == nil {
					u = base.ResolveReference(u)
				}
			}
			if cfg.cdnBaseURL != "" {
				cdn, cerr := url.Parse(cfg.cdnBaseURL)
				if cerr == nil {
					if u.IsAbs() {
						u.Scheme = cdn.Scheme
						u.Host = cdn.Host
					} else {
						cdn.Path = path.Join(cdn.Path, u.Path)
						u = cdn
					}
				}
			}
			if cfg.authToken != "" {
				q := u.Query()
				q.Set("token", cfg.authToken)
				u.RawQuery = q.Encode()
			}
			r.BaseURL = u.String()
		}
	}
	if cfg.customTransform != nil {
		cfg.customTransform(r)
	}
}

// parseFrameRate parses a DASH frameRate attribute ("30", "30000/1001", etc.)
// and returns the value as float64. Returns 0 on parse error or empty string.
func parseFrameRate(s string) float64 {
	if s == "" {
		return 0
	}
	if parts := strings.SplitN(s, "/", 2); len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil || den == 0 {
			return 0
		}
		return num / den
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
