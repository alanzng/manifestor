// Package dash provides parsing, filtering, building, and serialization of
// MPEG-DASH Media Presentation Description (MPD) documents.
package dash

// MPD represents a parsed MPEG-DASH Media Presentation Description.
type MPD struct {
	Profile         string
	Duration        string
	MinBufferTime   string
	MinUpdatePeriod string
	Periods         []Period
	// Raw holds the original XML bytes for attributes not mapped to fields.
	Raw []byte
}

// Period represents a single <Period> element in the MPD.
type Period struct {
	ID             string
	Start          string
	Duration       string
	AdaptationSets []AdaptationSet
}

// AudioChannelConfiguration represents an <AudioChannelConfiguration> element.
type AudioChannelConfiguration struct {
	SchemeIDURI string
	Value       string
}

// Role represents a <Role> element.
type Role struct {
	SchemeIDURI string
	Value       string
}

// AdaptationSet represents an <AdaptationSet> element.
type AdaptationSet struct {
	ID              string
	ContentType     string
	MimeType        string
	Lang            string
	Name            string
	Roles           []Role
	SegmentTemplate *SegmentTemplate
	SegmentBase     *SegmentBase
	Representations []Representation
}

// Representation represents a <Representation> element.
type Representation struct {
	ID                        string
	Bandwidth                 int
	Codecs                    string
	Width                     int
	Height                    int
	FrameRate                 string
	MimeType                  string
	StartWithSAP              int
	BaseURL                   string
	AudioChannelConfiguration *AudioChannelConfiguration
}

// SegmentTemplate represents a <SegmentTemplate> element.
type SegmentTemplate struct {
	Initialization string
	Media          string
	Timescale      int
	Duration       int
	StartNumber    int
}

// SegmentBase represents a <SegmentBase> element.
type SegmentBase struct {
	IndexRange     string
	Initialization string
}

// MPDConfig holds the top-level configuration for the DASH MPD builder.
type MPDConfig struct {
	Profile         string // required: "isoff-on-demand" | "isoff-live"
	Duration        string // optional ISO 8601 e.g. "PT4M0.00S"
	MinBufferTime   string // optional, default "PT1.5S"
	MinUpdatePeriod string // optional, live only
}

// RepresentationParams holds parameters for a single Representation in the builder.
type RepresentationParams struct {
	ID                        string // required, unique within AdaptationSet
	Bandwidth                 int    // required
	Codecs                    string
	Width                     int
	Height                    int
	FrameRate                 string
	MimeType                  string
	StartWithSAP              int
	BaseURL                   string
	AudioChannelConfiguration *AudioChannelConfiguration
}

// SegmentTemplateParams holds parameters for a SegmentTemplate.
type SegmentTemplateParams struct {
	Initialization string
	Media          string
	Timescale      int
	Duration       int
	StartNumber    int
}

// SegmentBaseParams holds parameters for a SegmentBase.
type SegmentBaseParams struct {
	IndexRange     string
	Initialization string
}

// AdaptationSetParams holds parameters for an AdaptationSet in the builder.
type AdaptationSetParams struct {
	ContentType     string // inferred from MimeType if blank
	MimeType        string
	Lang            string // BCP-47
	Name            string
	Roles           []Role
	SegmentTemplate *SegmentTemplateParams
	SegmentBase     *SegmentBaseParams
	Representations []RepresentationParams
}
