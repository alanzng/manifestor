package dash

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// xmlMPD is the XML decode target for the root <MPD> element.
type xmlMPD struct {
	XMLName                   xml.Name    `xml:"MPD"`
	Profiles                  string      `xml:"profiles,attr"`
	Type                      string      `xml:"type,attr"`
	MediaPresentationDuration string      `xml:"mediaPresentationDuration,attr"`
	MinBufferTime             string      `xml:"minBufferTime,attr"`
	MinimumUpdatePeriod       string      `xml:"minimumUpdatePeriod,attr"`
	Periods                   []xmlPeriod `xml:"Period"`
}

type xmlPeriod struct {
	ID             string          `xml:"id,attr"`
	Start          string          `xml:"start,attr"`
	Duration       string          `xml:"duration,attr"`
	AdaptationSets []xmlAdaptation `xml:"AdaptationSet"`
}

type xmlAdaptation struct {
	ID              string              `xml:"id,attr"`
	ContentType     string              `xml:"contentType,attr"`
	MimeType        string              `xml:"mimeType,attr"`
	Codecs          string              `xml:"codecs,attr"`
	Lang            string              `xml:"lang,attr"`
	StartWithSAP    string              `xml:"startWithSAP,attr"`
	Label           string              `xml:"label,attr"`
	Roles           []xmlRole           `xml:"Role"`
	SegmentTemplate *xmlSegmentTemplate `xml:"SegmentTemplate"`
	SegmentBase     *xmlSegmentBase     `xml:"SegmentBase"`
	Representations []xmlRepresentation `xml:"Representation"`
}

type xmlRole struct {
	SchemeIDURI string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

type xmlAudioChannelConfiguration struct {
	SchemeIDURI string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

type xmlRepresentation struct {
	ID                        string                        `xml:"id,attr"`
	Bandwidth                 string                        `xml:"bandwidth,attr"`
	Codecs                    string                        `xml:"codecs,attr"`
	Width                     string                        `xml:"width,attr"`
	Height                    string                        `xml:"height,attr"`
	FrameRate                 string                        `xml:"frameRate,attr"`
	MimeType                  string                        `xml:"mimeType,attr"`
	StartWithSAP              string                        `xml:"startWithSAP,attr"`
	BaseURL                   string                        `xml:"BaseURL"`
	AudioChannelConfiguration *xmlAudioChannelConfiguration `xml:"AudioChannelConfiguration"`
}

type xmlSegmentTemplate struct {
	Initialization string `xml:"initialization,attr"`
	Media          string `xml:"media,attr"`
	Timescale      string `xml:"timescale,attr"`
	Duration       string `xml:"duration,attr"`
	StartNumber    string `xml:"startNumber,attr"`
}

type xmlSegmentBase struct {
	IndexRange     string               `xml:"indexRange,attr"`
	Initialization *xmlSBInitialization `xml:"Initialization"`
}

type xmlSBInitialization struct {
	SourceURL string `xml:"sourceURL,attr"`
}

// Parse parses a DASH MPD document from a raw XML string and returns an MPD.
// Returns ErrParseFailure if the content cannot be parsed.
func Parse(content string) (*MPD, error) {
	var x xmlMPD
	if err := xml.Unmarshal([]byte(content), &x); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseFailure, err)
	}

	mpd := &MPD{
		Profile:         x.Profiles,
		Duration:        x.MediaPresentationDuration,
		MinBufferTime:   x.MinBufferTime,
		MinUpdatePeriod: x.MinimumUpdatePeriod,
		Raw:             []byte(content),
	}

	for _, xp := range x.Periods {
		p := Period{
			ID:       xp.ID,
			Start:    xp.Start,
			Duration: xp.Duration,
		}
		for _, xa := range xp.AdaptationSets {
			p.AdaptationSets = append(p.AdaptationSets, convertAdaptation(xa))
		}
		mpd.Periods = append(mpd.Periods, p)
	}

	return mpd, nil
}

func convertAdaptation(xa xmlAdaptation) AdaptationSet {
	as := AdaptationSet{
		ID:          xa.ID,
		ContentType: xa.ContentType,
		MimeType:    xa.MimeType,
		Lang:        xa.Lang,
		Name:        xa.Label,
	}
	for _, xr := range xa.Roles {
		as.Roles = append(as.Roles, Role(xr))
	}

	if xa.SegmentTemplate != nil {
		as.SegmentTemplate = convertSegmentTemplate(xa.SegmentTemplate)
	}
	if xa.SegmentBase != nil {
		as.SegmentBase = convertSegmentBase(xa.SegmentBase)
	}

	for _, xr := range xa.Representations {
		r := convertRepresentation(xr)
		// Inherit codecs from AdaptationSet if not set on Representation.
		if r.Codecs == "" && xa.Codecs != "" {
			r.Codecs = xa.Codecs
		}
		// Inherit MimeType from AdaptationSet if not set on Representation.
		if r.MimeType == "" && xa.MimeType != "" {
			r.MimeType = xa.MimeType
		}
		// Inherit StartWithSAP from AdaptationSet if not set on Representation.
		if r.StartWithSAP == 0 && xa.StartWithSAP != "" {
			r.StartWithSAP = atoi(xa.StartWithSAP)
		}
		as.Representations = append(as.Representations, r)
	}

	return as
}

func convertRepresentation(xr xmlRepresentation) Representation {
	r := Representation{
		ID:           xr.ID,
		Bandwidth:    atoi(xr.Bandwidth),
		Codecs:       xr.Codecs,
		Width:        atoi(xr.Width),
		Height:       atoi(xr.Height),
		FrameRate:    xr.FrameRate,
		MimeType:     xr.MimeType,
		StartWithSAP: atoi(xr.StartWithSAP),
		BaseURL:      xr.BaseURL,
	}
	if xr.AudioChannelConfiguration != nil {
		r.AudioChannelConfiguration = &AudioChannelConfiguration{
			SchemeIDURI: xr.AudioChannelConfiguration.SchemeIDURI,
			Value:       xr.AudioChannelConfiguration.Value,
		}
	}
	return r
}

func convertSegmentTemplate(x *xmlSegmentTemplate) *SegmentTemplate {
	return &SegmentTemplate{
		Initialization: x.Initialization,
		Media:          x.Media,
		Timescale:      atoi(x.Timescale),
		Duration:       atoi(x.Duration),
		StartNumber:    atoi(x.StartNumber),
	}
}

func convertSegmentBase(x *xmlSegmentBase) *SegmentBase {
	sb := &SegmentBase{IndexRange: x.IndexRange}
	if x.Initialization != nil {
		sb.Initialization = x.Initialization.SourceURL
	}
	return sb
}

// atoi converts a string to int, returning 0 on error.
func atoi(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}
