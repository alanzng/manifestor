package dash

import (
	"errors"
	"strings"
	"testing"
)

// ---- Validation errors ----

func TestBuild_ErrEmptyVariantList_NoAdaptationSets(t *testing.T) {
	b := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"})
	_, err := b.Build()
	if !errors.Is(err, ErrEmptyVariantList) {
		t.Errorf("got %v, want ErrEmptyVariantList", err)
	}
}

func TestBuild_ErrEmptyVariantList_NoRepresentations(t *testing.T) {
	b := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{MimeType: "video/mp4"})
	_, err := b.Build()
	if !errors.Is(err, ErrEmptyVariantList) {
		t.Errorf("got %v, want ErrEmptyVariantList", err)
	}
}

func TestBuild_ErrInvalidVariant_MissingID(t *testing.T) {
	b := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType: "video/mp4",
			Representations: []RepresentationParams{
				{Bandwidth: 5000000}, // no ID
			},
		})
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidVariant) {
		t.Errorf("got %v, want ErrInvalidVariant", err)
	}
}

func TestBuild_ErrInvalidVariant_MissingBandwidth(t *testing.T) {
	b := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType: "video/mp4",
			Representations: []RepresentationParams{
				{ID: "v1"}, // no Bandwidth
			},
		})
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidVariant) {
		t.Errorf("got %v, want ErrInvalidVariant", err)
	}
}

func TestBuild_ErrInvalidLanguageTag(t *testing.T) {
	b := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType: "audio/mp4",
			Lang:     "not a valid bcp47!!!",
			Representations: []RepresentationParams{
				{ID: "a1", Bandwidth: 128000},
			},
		})
	_, err := b.Build()
	if !errors.Is(err, ErrInvalidLanguageTag) {
		t.Errorf("got %v, want ErrInvalidLanguageTag", err)
	}
}

// ---- Minimal valid build ----

func TestBuild_Minimal(t *testing.T) {
	out, err := NewMPDBuilder(MPDConfig{
		Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011",
	}).AddAdaptationSet(AdaptationSetParams{
		MimeType: "video/mp4",
		Representations: []RepresentationParams{
			{ID: "v1", Bandwidth: 5000000},
		},
	}).Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "<?xml") {
		t.Error("output does not start with XML declaration")
	}
	if !strings.Contains(out, "urn:mpeg:dash:schema:mpd:2011") {
		t.Error("missing DASH namespace")
	}
}

// ---- Default MinBufferTime ----

func TestBuild_DefaultMinBufferTime(t *testing.T) {
	out, _ := NewMPDBuilder(MPDConfig{
		Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011",
	}).AddAdaptationSet(AdaptationSetParams{
		MimeType:        "video/mp4",
		Representations: []RepresentationParams{{ID: "v1", Bandwidth: 1000000}},
	}).Build()

	m, _ := Parse(out)
	if m.MinBufferTime != "PT1.5S" {
		t.Errorf("MinBufferTime = %q, want PT1.5S", m.MinBufferTime)
	}
}

func TestBuild_CustomMinBufferTime(t *testing.T) {
	out, _ := NewMPDBuilder(MPDConfig{
		Profile:       "urn:mpeg:dash:profile:isoff-on-demand:2011",
		MinBufferTime: "PT4S",
	}).AddAdaptationSet(AdaptationSetParams{
		MimeType:        "video/mp4",
		Representations: []RepresentationParams{{ID: "v1", Bandwidth: 1000000}},
	}).Build()

	m, _ := Parse(out)
	if m.MinBufferTime != "PT4S" {
		t.Errorf("MinBufferTime = %q, want PT4S", m.MinBufferTime)
	}
}

// ---- Profile and duration ----

func TestBuild_ProfilePreserved(t *testing.T) {
	profile := "urn:mpeg:dash:profile:isoff-on-demand:2011"
	out, _ := NewMPDBuilder(MPDConfig{Profile: profile}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType:        "video/mp4",
			Representations: []RepresentationParams{{ID: "v1", Bandwidth: 1000000}},
		}).Build()
	m, _ := Parse(out)
	if m.Profile != profile {
		t.Errorf("Profile = %q, want %q", m.Profile, profile)
	}
}

func TestBuild_DurationPreserved(t *testing.T) {
	out, _ := NewMPDBuilder(MPDConfig{
		Profile:  "urn:mpeg:dash:profile:isoff-on-demand:2011",
		Duration: "PT4M0.00S",
	}).AddAdaptationSet(AdaptationSetParams{
		MimeType:        "video/mp4",
		Representations: []RepresentationParams{{ID: "v1", Bandwidth: 1000000}},
	}).Build()
	m, _ := Parse(out)
	if m.Duration != "PT4M0.00S" {
		t.Errorf("Duration = %q, want PT4M0.00S", m.Duration)
	}
}

// ---- ContentType inferred from MimeType ----

func TestBuild_ContentTypeInferredVideo(t *testing.T) {
	out, _ := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType:        "video/mp4",
			Representations: []RepresentationParams{{ID: "v1", Bandwidth: 5000000}},
		}).Build()
	m, _ := Parse(out)
	if m.Periods[0].AdaptationSets[0].ContentType != "video" {
		t.Errorf("ContentType = %q, want video", m.Periods[0].AdaptationSets[0].ContentType)
	}
}

func TestBuild_ContentTypeInferredAudio(t *testing.T) {
	out, _ := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType:        "audio/mp4",
			Lang:            "en",
			Representations: []RepresentationParams{{ID: "a1", Bandwidth: 128000}},
		}).Build()
	m, _ := Parse(out)
	if m.Periods[0].AdaptationSets[0].ContentType != "audio" {
		t.Errorf("ContentType = %q, want audio", m.Periods[0].AdaptationSets[0].ContentType)
	}
}

// ---- Representations sorted by bandwidth ----

func TestBuild_RepresentationsSortedByBandwidth(t *testing.T) {
	out, err := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType: "video/mp4",
			Representations: []RepresentationParams{
				{ID: "v3", Bandwidth: 5000000},
				{ID: "v1", Bandwidth: 1000000},
				{ID: "v2", Bandwidth: 2800000},
			},
		}).Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	reps := m.Periods[0].AdaptationSets[0].Representations
	for i := 1; i < len(reps); i++ {
		if reps[i].Bandwidth < reps[i-1].Bandwidth {
			t.Errorf("reps[%d].Bandwidth=%d < reps[%d].Bandwidth=%d (not sorted)", i, reps[i].Bandwidth, i-1, reps[i-1].Bandwidth)
		}
	}
}

// ---- SegmentTemplate preserved ----

func TestBuild_SegmentTemplate(t *testing.T) {
	out, err := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType: "video/mp4",
			SegmentTemplate: &SegmentTemplateParams{
				Initialization: "$RepresentationID$/init.mp4",
				Media:          "$RepresentationID$/$Number$.m4s",
				Timescale:      90000,
				Duration:       270000,
				StartNumber:    1,
			},
			Representations: []RepresentationParams{
				{ID: "v1", Bandwidth: 5000000, Width: 1920, Height: 1080},
			},
		}).Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	st := m.Periods[0].AdaptationSets[0].SegmentTemplate
	if st == nil {
		t.Fatal("SegmentTemplate is nil")
	}
	if st.Initialization != "$RepresentationID$/init.mp4" {
		t.Errorf("Initialization = %q", st.Initialization)
	}
	if st.Timescale != 90000 {
		t.Errorf("Timescale = %d, want 90000", st.Timescale)
	}
	if st.Duration != 270000 {
		t.Errorf("Duration = %d, want 270000", st.Duration)
	}
	if st.StartNumber != 1 {
		t.Errorf("StartNumber = %d, want 1", st.StartNumber)
	}
}

// ---- Lang and multiple AdaptationSets ----

func TestBuild_MultipleAdaptationSets(t *testing.T) {
	out, err := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType:        "video/mp4",
			Representations: []RepresentationParams{{ID: "v1", Bandwidth: 5000000}},
		}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType:        "audio/mp4",
			Lang:            "en",
			Representations: []RepresentationParams{{ID: "a1", Bandwidth: 128000}},
		}).
		AddAdaptationSet(AdaptationSetParams{
			MimeType:        "audio/mp4",
			Lang:            "fr",
			Representations: []RepresentationParams{{ID: "a2", Bandwidth: 128000}},
		}).Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	if len(m.Periods[0].AdaptationSets) != 3 {
		t.Fatalf("adaptation sets = %d, want 3", len(m.Periods[0].AdaptationSets))
	}
	if m.Periods[0].AdaptationSets[1].Lang != "en" {
		t.Errorf("as[1].Lang = %q, want en", m.Periods[0].AdaptationSets[1].Lang)
	}
	if m.Periods[0].AdaptationSets[2].Lang != "fr" {
		t.Errorf("as[2].Lang = %q, want fr", m.Periods[0].AdaptationSets[2].Lang)
	}
}

// ---- BCP-47 validation ----

func TestBuild_ValidBCP47Tags(t *testing.T) {
	validTags := []string{"en", "en-US", "zh-Hans", "zh-Hans-CN", "tg", "fr"}
	for _, tag := range validTags {
		_, err := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
			AddAdaptationSet(AdaptationSetParams{
				MimeType: "audio/mp4",
				Lang:     tag,
				Representations: []RepresentationParams{
					{ID: "a1", Bandwidth: 128000},
				},
			}).Build()
		if err != nil {
			t.Errorf("valid BCP-47 tag %q rejected: %v", tag, err)
		}
	}
}

func TestBuild_InvalidBCP47Tags(t *testing.T) {
	invalidTags := []string{"not valid!!!", "en_US", "123!@#"}
	for _, tag := range invalidTags {
		_, err := NewMPDBuilder(MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}).
			AddAdaptationSet(AdaptationSetParams{
				MimeType: "audio/mp4",
				Lang:     tag,
				Representations: []RepresentationParams{
					{ID: "a1", Bandwidth: 128000},
				},
			}).Build()
		if !errors.Is(err, ErrInvalidLanguageTag) {
			t.Errorf("invalid BCP-47 tag %q: got %v, want ErrInvalidLanguageTag", tag, err)
		}
	}
}

// ---- Output is re-parseable ----

func TestBuild_OutputReparseable(t *testing.T) {
	out, err := NewMPDBuilder(MPDConfig{
		Profile:  "urn:mpeg:dash:profile:isoff-on-demand:2011",
		Duration: "PT4M0S",
	}).AddAdaptationSet(AdaptationSetParams{
		MimeType: "video/mp4",
		SegmentTemplate: &SegmentTemplateParams{
			Initialization: "$RepresentationID$/init.mp4",
			Media:          "$RepresentationID$/$Number$.m4s",
			Timescale:      90000,
			Duration:       270000,
			StartNumber:    1,
		},
		Representations: []RepresentationParams{
			{ID: "v1", Bandwidth: 5000000, Codecs: "avc1.640028", Width: 1920, Height: 1080, FrameRate: "30"},
			{ID: "v2", Bandwidth: 2800000, Codecs: "avc1.4d401f", Width: 1280, Height: 720, FrameRate: "30"},
		},
	}).AddAdaptationSet(AdaptationSetParams{
		MimeType: "audio/mp4",
		Lang:     "en",
		Representations: []RepresentationParams{
			{ID: "a1", Bandwidth: 128000, Codecs: "mp4a.40.2"},
		},
	}).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	m, err := Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(m.Periods) != 1 {
		t.Errorf("periods = %d, want 1", len(m.Periods))
	}
	if len(m.Periods[0].AdaptationSets) != 2 {
		t.Errorf("adaptation sets = %d, want 2", len(m.Periods[0].AdaptationSets))
	}
}
