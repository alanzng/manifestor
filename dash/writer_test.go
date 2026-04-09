package dash

import (
	"strings"
	"testing"
)

// ---- Round-trip helpers ----

func roundTrip(t *testing.T, fixture string, wantPeriods, wantAdaptationSets, wantReps int) {
	t.Helper()
	content := mustReadFixture(t, fixture)
	m, err := Parse(content)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := Serialize(m)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	m2, err := Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(m2.Periods) != wantPeriods {
		t.Errorf("re-parsed periods = %d, want %d", len(m2.Periods), wantPeriods)
	}
	totalAS := 0
	totalReps := 0
	for _, p := range m2.Periods {
		totalAS += len(p.AdaptationSets)
		for _, as := range p.AdaptationSets {
			totalReps += len(as.Representations)
		}
	}
	if totalAS != wantAdaptationSets {
		t.Errorf("re-parsed adaptation sets = %d, want %d", totalAS, wantAdaptationSets)
	}
	if totalReps != wantReps {
		t.Errorf("re-parsed representations = %d, want %d", totalReps, wantReps)
	}
}

// ---- Round-trip tests ----

func TestSerialize_RoundTrip_OnDemand(t *testing.T) {
	roundTrip(t, "../testdata/dash/isoff_ondemand.mpd", 1, 2, 5)
}

func TestSerialize_RoundTrip_Live(t *testing.T) {
	roundTrip(t, "../testdata/dash/isoff_live.mpd", 1, 3, 5)
}

func TestSerialize_RoundTrip_Azure(t *testing.T) {
	roundTrip(t, "../testdata/dash/azure_media_services.mpd", 1, 2, 5)
}

func TestSerialize_RoundTrip_Bento4Mixed(t *testing.T) {
	roundTrip(t, "../testdata/dash/bento4_mixed_codecs.mpd", 1, 3, 10)
}

// ---- Output format checks ----

func TestSerialize_StartsWithXMLDeclaration(t *testing.T) {
	m := &MPD{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}
	out, _ := Serialize(m)
	if !strings.HasPrefix(out, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>") {
		t.Errorf("output does not start with XML declaration:\n%s", out[:80])
	}
}

func TestSerialize_ContainsMPDNamespace(t *testing.T) {
	m := &MPD{}
	out, _ := Serialize(m)
	if !strings.Contains(out, "urn:mpeg:dash:schema:mpd:2011") {
		t.Errorf("missing DASH MPD namespace in output:\n%s", out)
	}
}

func TestSerialize_StaticType(t *testing.T) {
	m := &MPD{Duration: "PT1M0S"}
	out, _ := Serialize(m)
	if !strings.Contains(out, `type="static"`) {
		t.Errorf("expected type=static:\n%s", out)
	}
}

func TestSerialize_DynamicType_WhenMinUpdatePeriodSet(t *testing.T) {
	m := &MPD{MinUpdatePeriod: "PT2S"}
	out, _ := Serialize(m)
	if !strings.Contains(out, `type="dynamic"`) {
		t.Errorf("expected type=dynamic:\n%s", out)
	}
	if !strings.Contains(out, `minimumUpdatePeriod="PT2S"`) {
		t.Errorf("expected minimumUpdatePeriod in output:\n%s", out)
	}
}

func TestSerialize_ProfilePreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)
	if m2.Profile != m.Profile {
		t.Errorf("Profile = %q, want %q", m2.Profile, m.Profile)
	}
}

func TestSerialize_DurationPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)
	if m2.Duration != m.Duration {
		t.Errorf("Duration = %q, want %q", m2.Duration, m.Duration)
	}
}

func TestSerialize_MinBufferTimePreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)
	if m2.MinBufferTime != m.MinBufferTime {
		t.Errorf("MinBufferTime = %q, want %q", m2.MinBufferTime, m.MinBufferTime)
	}
}

// ---- Field preservation ----

func TestSerialize_RepresentationFieldsPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	orig := m.Periods[0].AdaptationSets[0].Representations[0]
	got := m2.Periods[0].AdaptationSets[0].Representations[0]

	if got.ID != orig.ID {
		t.Errorf("ID = %q, want %q", got.ID, orig.ID)
	}
	if got.Bandwidth != orig.Bandwidth {
		t.Errorf("Bandwidth = %d, want %d", got.Bandwidth, orig.Bandwidth)
	}
	if got.Codecs != orig.Codecs {
		t.Errorf("Codecs = %q, want %q", got.Codecs, orig.Codecs)
	}
	if got.Width != orig.Width || got.Height != orig.Height {
		t.Errorf("Resolution = %dx%d, want %dx%d", got.Width, got.Height, orig.Width, orig.Height)
	}
	if got.FrameRate != orig.FrameRate {
		t.Errorf("FrameRate = %q, want %q", got.FrameRate, orig.FrameRate)
	}
}

func TestSerialize_AdaptationSetFieldsPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	orig := m.Periods[0].AdaptationSets[1] // English audio
	got := m2.Periods[0].AdaptationSets[1]

	if got.Lang != orig.Lang {
		t.Errorf("Lang = %q, want %q", got.Lang, orig.Lang)
	}
	if got.MimeType != orig.MimeType {
		t.Errorf("MimeType = %q, want %q", got.MimeType, orig.MimeType)
	}
	if got.ContentType != orig.ContentType {
		t.Errorf("ContentType = %q, want %q", got.ContentType, orig.ContentType)
	}
}

func TestSerialize_SegmentTemplatePreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	orig := m.Periods[0].AdaptationSets[0].SegmentTemplate
	got := m2.Periods[0].AdaptationSets[0].SegmentTemplate

	if got == nil {
		t.Fatal("SegmentTemplate is nil after round-trip")
	}
	if got.Initialization != orig.Initialization {
		t.Errorf("Initialization = %q, want %q", got.Initialization, orig.Initialization)
	}
	if got.Media != orig.Media {
		t.Errorf("Media = %q, want %q", got.Media, orig.Media)
	}
	if got.Timescale != orig.Timescale {
		t.Errorf("Timescale = %d, want %d", got.Timescale, orig.Timescale)
	}
	if got.Duration != orig.Duration {
		t.Errorf("Duration = %d, want %d", got.Duration, orig.Duration)
	}
	if got.StartNumber != orig.StartNumber {
		t.Errorf("StartNumber = %d, want %d", got.StartNumber, orig.StartNumber)
	}
}

func TestSerialize_LangPreserved_MultipleTracks(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	as := m2.Periods[0].AdaptationSets
	if as[1].Lang != "en" {
		t.Errorf("as[1].Lang = %q, want en", as[1].Lang)
	}
	if as[2].Lang != "fr" {
		t.Errorf("as[2].Lang = %q, want fr", as[2].Lang)
	}
}

// ---- Bento4 mixed-codec round-trip ----

func TestSerialize_Bento4Mixed_AllRepresentationsPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	hevc := m2.Periods[0].AdaptationSets[0].Representations
	if len(hevc) != 5 {
		t.Errorf("HEVC reps = %d, want 5", len(hevc))
	}
	avc := m2.Periods[0].AdaptationSets[1].Representations
	if len(avc) != 4 {
		t.Errorf("AVC1 reps = %d, want 4", len(avc))
	}
	audio := m2.Periods[0].AdaptationSets[2].Representations
	if len(audio) != 1 {
		t.Errorf("audio reps = %d, want 1", len(audio))
	}
}

func TestSerialize_Bento4Mixed_CodecsPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	r := m2.Periods[0].AdaptationSets[0].Representations[0]
	if r.Codecs != "hvc1.1.2.L120.90" {
		t.Errorf("Codecs = %q, want hvc1.1.2.L120.90", r.Codecs)
	}
	r2 := m2.Periods[0].AdaptationSets[1].Representations[1]
	if r2.Codecs != "avc1.64001F" {
		t.Errorf("Codecs = %q, want avc1.64001F", r2.Codecs)
	}
}

func TestSerialize_Bento4Mixed_AudioLangPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	as := m2.Periods[0].AdaptationSets[2]
	if as.Lang != "tg" {
		t.Errorf("audio Lang = %q, want tg", as.Lang)
	}
}

func TestSerialize_BaseURL_RoundTrip(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	r := m2.Periods[0].AdaptationSets[0].Representations[0]
	if r.BaseURL != "media-video-hvc1-1.mp4" {
		t.Errorf("BaseURL = %q, want media-video-hvc1-1.mp4", r.BaseURL)
	}
}

func TestSerialize_AudioChannelConfiguration_RoundTrip(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m, _ := Parse(content)
	out, _ := Serialize(m)
	m2, _ := Parse(out)

	r := m2.Periods[0].AdaptationSets[2].Representations[0]
	if r.AudioChannelConfiguration == nil {
		t.Fatal("AudioChannelConfiguration is nil after round-trip")
	}
	if r.AudioChannelConfiguration.Value != "2" {
		t.Errorf("Value = %q, want 2", r.AudioChannelConfiguration.Value)
	}
}

func TestSerialize_AdaptationSetLabel(t *testing.T) {
	m := &MPD{
		Periods: []Period{{
			AdaptationSets: []AdaptationSet{{
				MimeType: "video/mp4",
				Name:     "Main Video",
				Representations: []Representation{{
					ID: "v1", Bandwidth: 1000000, Codecs: "avc1.64001F",
				}},
			}},
		}},
	}
	out, err := Serialize(m)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if !strings.Contains(out, `label="Main Video"`) {
		t.Errorf("output missing label attribute:\n%s", out)
	}
}

func TestSerialize_SegmentBase_RoundTrip(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")
	m, _ := Parse(content)
	out, err := Serialize(m)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	m2, _ := Parse(out)

	r := m2.Periods[0].AdaptationSets[0].Representations[0]
	if r.SegmentBase == nil {
		t.Fatal("SegmentBase is nil after round-trip")
	}
	if r.SegmentBase.IndexRange != "824-6591" {
		t.Errorf("IndexRange = %q, want 824-6591", r.SegmentBase.IndexRange)
	}
	if r.SegmentBase.InitializationRange != "0-823" {
		t.Errorf("InitializationRange = %q, want 0-823", r.SegmentBase.InitializationRange)
	}
}

func TestSerialize_SegmentBase_OnAdaptationSet_RoundTrip(t *testing.T) {
	m := &MPD{
		Periods: []Period{{
			AdaptationSets: []AdaptationSet{{
				MimeType: "video/mp4",
				SegmentBase: &SegmentBase{
					IndexRange:     "0-819",
					Initialization: "init.mp4",
				},
				Representations: []Representation{{
					ID: "v1", Bandwidth: 3000000, Codecs: "avc1.64001F",
				}},
			}},
		}},
	}
	out, err := Serialize(m)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	m2, _ := Parse(out)

	as := m2.Periods[0].AdaptationSets[0]
	if as.SegmentBase == nil {
		t.Fatal("SegmentBase is nil after round-trip")
	}
	if as.SegmentBase.IndexRange != "0-819" {
		t.Errorf("IndexRange = %q, want 0-819", as.SegmentBase.IndexRange)
	}
	if as.SegmentBase.Initialization != "init.mp4" {
		t.Errorf("Initialization = %q, want init.mp4", as.SegmentBase.Initialization)
	}
}

func TestSerialize_AdaptationSetRole(t *testing.T) {
	m := &MPD{
		Periods: []Period{{
			AdaptationSets: []AdaptationSet{{
				MimeType: "audio/mp4",
				Roles:    []Role{{SchemeIDURI: "urn:mpeg:dash:role:2011", Value: "main"}},
				Representations: []Representation{{
					ID: "a1", Bandwidth: 128000,
				}},
			}},
		}},
	}
	out, err := Serialize(m)
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if !strings.Contains(out, `schemeIdUri="urn:mpeg:dash:role:2011"`) {
		t.Errorf("output missing Role schemeIdUri:\n%s", out)
	}
	if !strings.Contains(out, `value="main"`) {
		t.Errorf("output missing Role value:\n%s", out)
	}
}
