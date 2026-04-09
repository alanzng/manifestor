package dash

import (
	"encoding/xml"
	"fmt"
)

// Serialize serializes an MPD back to a valid DASH MPD XML string.
// The output always begins with an XML declaration and uses the standard
// DASH MPD namespace. Unknown attributes and elements are not preserved
// (use MPD.Raw if the original bytes are needed verbatim).
func Serialize(m *MPD) (string, error) {
	x := buildXMLMPD(m)
	b, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		return "", fmt.Errorf("dash: serialize failed: %w", err)
	}
	return xml.Header + string(b) + "\n", nil
}

// buildXMLMPD converts an MPD to its xml-encodable representation.
func buildXMLMPD(m *MPD) *xmlOutMPD {
	x := &xmlOutMPD{
		XMLNS:                     "urn:mpeg:dash:schema:mpd:2011",
		Profiles:                  m.Profile,
		Type:                      "static",
		MediaPresentationDuration: m.Duration,
		MinBufferTime:             m.MinBufferTime,
	}
	if m.MinUpdatePeriod != "" {
		x.Type = "dynamic"
		x.MinimumUpdatePeriod = m.MinUpdatePeriod
	}

	for _, p := range m.Periods {
		xp := xmlOutPeriod{ID: p.ID, Start: p.Start, Duration: p.Duration}
		for _, as := range p.AdaptationSets {
			xp.AdaptationSets = append(xp.AdaptationSets, buildXMLAdaptation(as))
		}
		x.Periods = append(x.Periods, xp)
	}
	return x
}

func buildXMLAdaptation(as AdaptationSet) xmlOutAdaptation {
	xa := xmlOutAdaptation{
		ID:          as.ID,
		ContentType: as.ContentType,
		MimeType:    as.MimeType,
		Lang:        as.Lang,
		Label:       as.Name,
	}
	for _, role := range as.Roles {
		xa.Roles = append(xa.Roles, xmlOutRole(role))
	}
	if as.SegmentTemplate != nil {
		st := as.SegmentTemplate
		xa.SegmentTemplate = &xmlOutSegmentTemplate{
			Initialization: st.Initialization,
			Media:          st.Media,
			Timescale:      omitZeroInt(st.Timescale),
			Duration:       omitZeroInt(st.Duration),
			StartNumber:    omitZeroInt(st.StartNumber),
		}
	}
	if as.SegmentBase != nil {
		sb := as.SegmentBase
		xsb := &xmlOutSegmentBase{IndexRange: sb.IndexRange}
		if sb.Initialization != "" || sb.InitializationRange != "" {
			xsb.Initialization = &xmlOutSBInitialization{
				SourceURL: sb.Initialization,
				Range:     sb.InitializationRange,
			}
		}
		xa.SegmentBase = xsb
	}
	for _, r := range as.Representations {
		xr := xmlOutRepresentation{
			ID:        r.ID,
			Bandwidth: r.Bandwidth,
			Codecs:    r.Codecs,
			FrameRate: r.FrameRate,
			BaseURL:   r.BaseURL,
		}
		if r.Width > 0 && r.Height > 0 {
			xr.Width = r.Width
			xr.Height = r.Height
		}
		// Omit MimeType on Representation when it matches the AdaptationSet.
		if r.MimeType != "" && r.MimeType != as.MimeType {
			xr.MimeType = r.MimeType
		}
		if r.AudioChannelConfiguration != nil {
			xr.AudioChannelConfiguration = &xmlOutAudioChannelConfiguration{
				SchemeIDURI: r.AudioChannelConfiguration.SchemeIDURI,
				Value:       r.AudioChannelConfiguration.Value,
			}
		}
		xa.Representations = append(xa.Representations, xr)
	}
	return xa
}

// omitZeroInt returns the string form of n, or "" when n is 0.
func omitZeroInt(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

// ---- XML output structs ----

type xmlOutMPD struct {
	XMLName                   xml.Name       `xml:"MPD"`
	XMLNS                     string         `xml:"xmlns,attr"`
	Profiles                  string         `xml:"profiles,attr,omitempty"`
	Type                      string         `xml:"type,attr,omitempty"`
	MediaPresentationDuration string         `xml:"mediaPresentationDuration,attr,omitempty"`
	MinBufferTime             string         `xml:"minBufferTime,attr,omitempty"`
	MinimumUpdatePeriod       string         `xml:"minimumUpdatePeriod,attr,omitempty"`
	Periods                   []xmlOutPeriod `xml:"Period"`
}

type xmlOutPeriod struct {
	ID             string             `xml:"id,attr,omitempty"`
	Start          string             `xml:"start,attr,omitempty"`
	Duration       string             `xml:"duration,attr,omitempty"`
	AdaptationSets []xmlOutAdaptation `xml:"AdaptationSet"`
}

type xmlOutRole struct {
	SchemeIDURI string `xml:"schemeIdUri,attr,omitempty"`
	Value       string `xml:"value,attr,omitempty"`
}

type xmlOutAudioChannelConfiguration struct {
	SchemeIDURI string `xml:"schemeIdUri,attr,omitempty"`
	Value       string `xml:"value,attr,omitempty"`
}

type xmlOutAdaptation struct {
	ID              string                 `xml:"id,attr,omitempty"`
	ContentType     string                 `xml:"contentType,attr,omitempty"`
	MimeType        string                 `xml:"mimeType,attr,omitempty"`
	Lang            string                 `xml:"lang,attr,omitempty"`
	Label           string                 `xml:"label,attr,omitempty"`
	Roles           []xmlOutRole           `xml:"Role,omitempty"`
	SegmentTemplate *xmlOutSegmentTemplate `xml:"SegmentTemplate,omitempty"`
	SegmentBase     *xmlOutSegmentBase     `xml:"SegmentBase,omitempty"`
	Representations []xmlOutRepresentation `xml:"Representation"`
}

type xmlOutSegmentBase struct {
	IndexRange     string                  `xml:"indexRange,attr,omitempty"`
	Initialization *xmlOutSBInitialization `xml:"Initialization,omitempty"`
}

type xmlOutSBInitialization struct {
	SourceURL string `xml:"sourceURL,attr,omitempty"`
	Range     string `xml:"range,attr,omitempty"`
}

type xmlOutRepresentation struct {
	ID                        string                           `xml:"id,attr,omitempty"`
	Bandwidth                 int                              `xml:"bandwidth,attr"`
	Codecs                    string                           `xml:"codecs,attr,omitempty"`
	Width                     int                              `xml:"width,attr,omitempty"`
	Height                    int                              `xml:"height,attr,omitempty"`
	FrameRate                 string                           `xml:"frameRate,attr,omitempty"`
	MimeType                  string                           `xml:"mimeType,attr,omitempty"`
	BaseURL                   string                           `xml:"BaseURL,omitempty"`
	AudioChannelConfiguration *xmlOutAudioChannelConfiguration `xml:"AudioChannelConfiguration,omitempty"`
}

type xmlOutSegmentTemplate struct {
	Initialization string `xml:"initialization,attr,omitempty"`
	Media          string `xml:"media,attr,omitempty"`
	Timescale      string `xml:"timescale,attr,omitempty"`
	Duration       string `xml:"duration,attr,omitempty"`
	StartNumber    string `xml:"startNumber,attr,omitempty"`
}
