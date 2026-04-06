package dash

import (
	"sort"
	"strings"
	"unicode"
)

// MPDBuilder builds a DASH MPD document from scratch.
// Use NewMPDBuilder to create one.
type MPDBuilder struct {
	cfg            MPDConfig
	adaptationSets []AdaptationSetParams
}

// NewMPDBuilder returns a new MPDBuilder with the given top-level configuration.
func NewMPDBuilder(cfg MPDConfig) *MPDBuilder {
	if cfg.MinBufferTime == "" {
		cfg.MinBufferTime = "PT1.5S"
	}
	return &MPDBuilder{cfg: cfg}
}

// AddAdaptationSet appends an AdaptationSet to the single Period.
// Representations within the set are sorted by ascending Bandwidth at build time.
func (b *MPDBuilder) AddAdaptationSet(p AdaptationSetParams) *MPDBuilder {
	b.adaptationSets = append(b.adaptationSets, p)
	return b
}

// Build validates the configuration and serializes the MPD to a valid XML string.
//
// Returns ErrEmptyVariantList if no AdaptationSets (or no Representations) were added.
// Returns ErrInvalidVariant if any representation is missing ID or Bandwidth.
// Returns ErrInvalidLanguageTag if any AdaptationSet has a malformed BCP-47 lang.
func (b *MPDBuilder) Build() (string, error) {
	if len(b.adaptationSets) == 0 {
		return "", ErrEmptyVariantList
	}

	totalReps := 0
	for _, as := range b.adaptationSets {
		totalReps += len(as.Representations)
	}
	if totalReps == 0 {
		return "", ErrEmptyVariantList
	}

	// Validate all representations and language tags.
	for _, as := range b.adaptationSets {
		if as.Lang != "" && !isValidBCP47(as.Lang) {
			return "", ErrInvalidLanguageTag
		}
		for _, r := range as.Representations {
			if r.ID == "" || r.Bandwidth == 0 {
				return "", ErrInvalidVariant
			}
		}
	}

	period := Period{}
	for _, asp := range b.adaptationSets {
		as := convertAdaptationSetParams(asp)
		// Sort representations by ascending Bandwidth.
		sort.Slice(as.Representations, func(i, j int) bool {
			return as.Representations[i].Bandwidth < as.Representations[j].Bandwidth
		})
		period.AdaptationSets = append(period.AdaptationSets, as)
	}

	m := &MPD{
		Profile:         b.cfg.Profile,
		Duration:        b.cfg.Duration,
		MinBufferTime:   b.cfg.MinBufferTime,
		MinUpdatePeriod: b.cfg.MinUpdatePeriod,
		Periods:         []Period{period},
	}
	return Serialize(m)
}

// convertAdaptationSetParams converts AdaptationSetParams to AdaptationSet.
func convertAdaptationSetParams(p AdaptationSetParams) AdaptationSet {
	as := AdaptationSet{
		ContentType: p.ContentType,
		MimeType:    p.MimeType,
		Lang:        p.Lang,
	}

	// Infer ContentType from MimeType if not explicitly set.
	if as.ContentType == "" && p.MimeType != "" {
		if strings.HasPrefix(p.MimeType, "video/") {
			as.ContentType = "video"
		} else if strings.HasPrefix(p.MimeType, "audio/") {
			as.ContentType = "audio"
		}
	}

	if p.SegmentTemplate != nil {
		as.SegmentTemplate = &SegmentTemplate{
			Initialization: p.SegmentTemplate.Initialization,
			Media:          p.SegmentTemplate.Media,
			Timescale:      p.SegmentTemplate.Timescale,
			Duration:       p.SegmentTemplate.Duration,
			StartNumber:    p.SegmentTemplate.StartNumber,
		}
	}
	if p.SegmentBase != nil {
		as.SegmentBase = &SegmentBase{
			IndexRange:     p.SegmentBase.IndexRange,
			Initialization: p.SegmentBase.Initialization,
		}
	}

	for _, r := range p.Representations {
		as.Representations = append(as.Representations, Representation(r))
	}
	return as
}

// isValidBCP47 performs a lightweight check that lang looks like a BCP-47 tag.
// It accepts tags of the form "ll", "ll-RR", "ll-Ssss", "ll-Ssss-RR", etc.
// where subtags are 2–8 alphanumeric characters separated by hyphens.
func isValidBCP47(lang string) bool {
	if lang == "" {
		return false
	}
	for _, sub := range strings.Split(lang, "-") {
		if len(sub) < 1 || len(sub) > 8 {
			return false
		}
		for _, c := range sub {
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}
