package dash

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
	// TODO: implement
	panic("not implemented")
}
