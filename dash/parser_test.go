package dash

import (
	"errors"
	"os"
	"testing"
)

func mustReadFixture(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(b)
}

// ---- isoff-on-demand fixture ----

func TestParse_OnDemand_TopLevel(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mpd.Duration != "PT4M0.00S" {
		t.Errorf("Duration = %q, want PT4M0.00S", mpd.Duration)
	}
	if mpd.MinBufferTime != "PT1.5S" {
		t.Errorf("MinBufferTime = %q, want PT1.5S", mpd.MinBufferTime)
	}
	if mpd.MinUpdatePeriod != "" {
		t.Errorf("MinUpdatePeriod = %q, want empty", mpd.MinUpdatePeriod)
	}
}

func TestParse_OnDemand_Profile(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	want := "urn:mpeg:dash:profile:isoff-on-demand:2011"
	if mpd.Profile != want {
		t.Errorf("Profile = %q, want %q", mpd.Profile, want)
	}
}

func TestParse_OnDemand_PeriodCount(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	if len(mpd.Periods) != 1 {
		t.Fatalf("Periods = %d, want 1", len(mpd.Periods))
	}
	if mpd.Periods[0].ID != "1" {
		t.Errorf("Period.ID = %q, want 1", mpd.Periods[0].ID)
	}
}

func TestParse_OnDemand_AdaptationSets(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	as := mpd.Periods[0].AdaptationSets
	if len(as) != 2 {
		t.Fatalf("AdaptationSets = %d, want 2", len(as))
	}
	if as[0].ContentType != "video" {
		t.Errorf("AdaptationSets[0].ContentType = %q, want video", as[0].ContentType)
	}
	if as[1].ContentType != "audio" {
		t.Errorf("AdaptationSets[1].ContentType = %q, want audio", as[1].ContentType)
	}
	if as[1].Lang != "en" {
		t.Errorf("AdaptationSets[1].Lang = %q, want en", as[1].Lang)
	}
}

func TestParse_OnDemand_VideoRepresentations(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 4 {
		t.Fatalf("video representations = %d, want 4", len(reps))
	}

	tests := []struct {
		id        string
		bandwidth int
		width     int
		height    int
		codecs    string
	}{
		{"v1", 5000000, 1920, 1080, "avc1.640028"},
		{"v2", 2800000, 1280, 720, "avc1.4d401f"},
		{"v3", 1400000, 854, 480, "avc1.4d401e"},
		{"v4", 600000, 640, 360, "avc1.42c01e"},
	}
	for i, tt := range tests {
		r := reps[i]
		if r.ID != tt.id {
			t.Errorf("[%d] ID = %q, want %q", i, r.ID, tt.id)
		}
		if r.Bandwidth != tt.bandwidth {
			t.Errorf("[%d] Bandwidth = %d, want %d", i, r.Bandwidth, tt.bandwidth)
		}
		if r.Width != tt.width || r.Height != tt.height {
			t.Errorf("[%d] Resolution = %dx%d, want %dx%d", i, r.Width, r.Height, tt.width, tt.height)
		}
		if r.Codecs != tt.codecs {
			t.Errorf("[%d] Codecs = %q, want %q", i, r.Codecs, tt.codecs)
		}
	}
}

func TestParse_OnDemand_SegmentTemplate(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	st := mpd.Periods[0].AdaptationSets[0].SegmentTemplate
	if st == nil {
		t.Fatal("SegmentTemplate is nil")
	}
	if st.Initialization != "$RepresentationID$/init.mp4" {
		t.Errorf("Initialization = %q", st.Initialization)
	}
	if st.Media != "$RepresentationID$/$Number$.m4s" {
		t.Errorf("Media = %q", st.Media)
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

func TestParse_OnDemand_FrameRate(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	r := mpd.Periods[0].AdaptationSets[0].Representations[0]
	if r.FrameRate != "30" {
		t.Errorf("FrameRate = %q, want 30", r.FrameRate)
	}
}

func TestParse_OnDemand_RawPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	if len(mpd.Raw) == 0 {
		t.Error("Raw is empty, want original XML bytes")
	}
	if string(mpd.Raw) != content {
		t.Error("Raw does not match input content")
	}
}

// ---- isoff-live fixture ----

func TestParse_Live_MinUpdatePeriod(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	mpd, _ := Parse(content)
	if mpd.MinUpdatePeriod != "PT2S" {
		t.Errorf("MinUpdatePeriod = %q, want PT2S", mpd.MinUpdatePeriod)
	}
}

func TestParse_Live_MultipleAudioTracks(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	mpd, _ := Parse(content)
	as := mpd.Periods[0].AdaptationSets
	if len(as) != 3 {
		t.Fatalf("AdaptationSets = %d, want 3", len(as))
	}
	// as[1] = English audio, as[2] = French audio
	if as[1].Lang != "en" {
		t.Errorf("as[1].Lang = %q, want en", as[1].Lang)
	}
	if as[2].Lang != "fr" {
		t.Errorf("as[2].Lang = %q, want fr", as[2].Lang)
	}
}

func TestParse_Live_VideoRepresentations(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 3 {
		t.Fatalf("video representations = %d, want 3", len(reps))
	}
}

// ---- Azure Media Services fixture ----

func TestParse_Azure_CodecsInheritedFromAdaptationSet(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/azure_media_services.mpd")
	mpd, _ := Parse(content)
	// Azure puts codecs on AdaptationSet; representations should inherit.
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	if len(reps) == 0 {
		t.Fatal("no representations")
	}
	for i, r := range reps {
		if r.Codecs != "avc1.640028" {
			t.Errorf("reps[%d].Codecs = %q, want avc1.640028 (inherited)", i, r.Codecs)
		}
	}
}

func TestParse_Azure_MimeTypeInherited(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/azure_media_services.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	for i, r := range reps {
		if r.MimeType != "video/mp4" {
			t.Errorf("reps[%d].MimeType = %q, want video/mp4 (inherited)", i, r.MimeType)
		}
	}
}

func TestParse_Azure_PeriodWithoutID(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/azure_media_services.mpd")
	mpd, _ := Parse(content)
	// Azure fixture has a Period with no id attribute.
	if len(mpd.Periods) != 1 {
		t.Fatalf("Periods = %d, want 1", len(mpd.Periods))
	}
	// No assertion on ID value — it is legitimately empty.
}

func TestParse_Azure_VideoRepresentationCount(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/azure_media_services.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 4 {
		t.Errorf("video representations = %d, want 4", len(reps))
	}
}

// ---- Error paths ----

func TestParse_ErrParseFailure_NotXML(t *testing.T) {
	_, err := Parse("not xml at all")
	if !errors.Is(err, ErrParseFailure) {
		t.Errorf("got %v, want ErrParseFailure", err)
	}
}

func TestParse_ErrParseFailure_EmptyString(t *testing.T) {
	_, err := Parse("")
	if !errors.Is(err, ErrParseFailure) {
		t.Errorf("got %v, want ErrParseFailure", err)
	}
}

func TestParse_ErrParseFailure_HLSContent(t *testing.T) {
	hls := "#EXTM3U\n#EXT-X-VERSION:3\n"
	_, err := Parse(hls)
	if !errors.Is(err, ErrParseFailure) {
		t.Errorf("got %v, want ErrParseFailure", err)
	}
}

// ---- StartWithSAP inheritance ----

func TestParse_OnDemand_StartWithSAPInherited(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	for i, r := range reps {
		if r.StartWithSAP != 1 {
			t.Errorf("reps[%d].StartWithSAP = %d, want 1 (inherited)", i, r.StartWithSAP)
		}
	}
}

// ---- Bento4 mixed-codecs fixture ----

func TestParse_Bento4Mixed_TopLevel(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mpd.Duration != "PT1H22M24.040S" {
		t.Errorf("Duration = %q, want PT1H22M24.040S", mpd.Duration)
	}
	if mpd.MinBufferTime != "PT4.00S" {
		t.Errorf("MinBufferTime = %q, want PT4.00S", mpd.MinBufferTime)
	}
}

func TestParse_Bento4Mixed_AdaptationSetCount(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	as := mpd.Periods[0].AdaptationSets
	// 2 video sets + 1 audio set
	if len(as) != 3 {
		t.Fatalf("AdaptationSets = %d, want 3", len(as))
	}
}

func TestParse_Bento4Mixed_HEVCRepresentations(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 5 {
		t.Fatalf("HEVC representations = %d, want 5", len(reps))
	}

	tests := []struct {
		id        string
		bandwidth int
		width     int
		height    int
		codecs    string
		frameRate string
	}{
		{"video-hvc1-1", 2455725, 1280, 720, "hvc1.1.2.L120.90", "50"},
		{"video-hvc1-2", 5541941, 1920, 1080, "hvc1.1.2.L123.90", "50"},
		{"video-hvc1-3", 757752, 854, 480, "hvc1.1.2.L93.90", "50"},
		{"video-hvc1-4", 8638946, 2560, 1440, "hvc1.1.2.L150.90", "50"},
		{"video-hvc1-5", 21552440, 3840, 2160, "hvc1.1.2.L153.90", "50"},
	}
	for i, tt := range tests {
		r := reps[i]
		if r.ID != tt.id {
			t.Errorf("[%d] ID = %q, want %q", i, r.ID, tt.id)
		}
		if r.Bandwidth != tt.bandwidth {
			t.Errorf("[%d] Bandwidth = %d, want %d", i, r.Bandwidth, tt.bandwidth)
		}
		if r.Width != tt.width || r.Height != tt.height {
			t.Errorf("[%d] Resolution = %dx%d, want %dx%d", i, r.Width, r.Height, tt.width, tt.height)
		}
		if r.Codecs != tt.codecs {
			t.Errorf("[%d] Codecs = %q, want %q", i, r.Codecs, tt.codecs)
		}
		if r.FrameRate != tt.frameRate {
			t.Errorf("[%d] FrameRate = %q, want %q", i, r.FrameRate, tt.frameRate)
		}
	}
}

func TestParse_Bento4Mixed_AVC1Representations(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	reps := mpd.Periods[0].AdaptationSets[1].Representations
	if len(reps) != 4 {
		t.Fatalf("AVC1 representations = %d, want 4", len(reps))
	}

	tests := []struct {
		id        string
		bandwidth int
		codecs    string
	}{
		{"video-avc1-1", 4616502, "avc1.640028"},
		{"video-avc1-2", 2363259, "avc1.64001F"},
		{"video-avc1-3", 930116, "avc1.4D401E"},
		{"video-avc1-4", 1528065, "avc1.4D401E"},
	}
	for i, tt := range tests {
		r := reps[i]
		if r.ID != tt.id {
			t.Errorf("[%d] ID = %q, want %q", i, r.ID, tt.id)
		}
		if r.Bandwidth != tt.bandwidth {
			t.Errorf("[%d] Bandwidth = %d, want %d", i, r.Bandwidth, tt.bandwidth)
		}
		if r.Codecs != tt.codecs {
			t.Errorf("[%d] Codecs = %q, want %q", i, r.Codecs, tt.codecs)
		}
	}
}

func TestParse_Bento4Mixed_4KResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	r := mpd.Periods[0].AdaptationSets[0].Representations[4] // video-hvc1-5
	if r.Width != 3840 || r.Height != 2160 {
		t.Errorf("4K resolution = %dx%d, want 3840x2160", r.Width, r.Height)
	}
}

func TestParse_Bento4Mixed_AudioAdaptationSet(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	as := mpd.Periods[0].AdaptationSets[2]
	if as.MimeType != "audio/mp4" {
		t.Errorf("audio MimeType = %q, want audio/mp4", as.MimeType)
	}
	if as.Lang != "tg" {
		t.Errorf("audio Lang = %q, want tg", as.Lang)
	}
	if len(as.Representations) != 1 {
		t.Fatalf("audio representations = %d, want 1", len(as.Representations))
	}
	r := as.Representations[0]
	if r.ID != "audio-tg-mp4a.40.2" {
		t.Errorf("audio ID = %q, want audio-tg-mp4a.40.2", r.ID)
	}
	if r.Bandwidth != 197665 {
		t.Errorf("audio Bandwidth = %d, want 197665", r.Bandwidth)
	}
	if r.Codecs != "mp4a.40.2" {
		t.Errorf("audio Codecs = %q, want mp4a.40.2", r.Codecs)
	}
}

func TestParse_Bento4Mixed_SegmentBase(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	// Representations use SegmentBase (not SegmentTemplate) in this fixture.
	// AdaptationSet has no SegmentTemplate.
	as := mpd.Periods[0].AdaptationSets[0]
	if as.SegmentTemplate != nil {
		t.Error("expected no SegmentTemplate on AdaptationSet")
	}
}

func TestParse_Bento4Mixed_MimeTypeInheritedByAllReps(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	for _, as := range mpd.Periods[0].AdaptationSets {
		for i, r := range as.Representations {
			if r.MimeType != as.MimeType {
				t.Errorf("as[%s] rep[%d].MimeType = %q, want %q (inherited)", as.ID, i, r.MimeType, as.MimeType)
			}
		}
	}
}

func TestParse_Bento4Mixed_BaseURL(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	// First AdaptationSet is HEVC video; first representation has BaseURL.
	r := mpd.Periods[0].AdaptationSets[0].Representations[0]
	if r.BaseURL != "media-video-hvc1-1.mp4" {
		t.Errorf("BaseURL = %q, want media-video-hvc1-1.mp4", r.BaseURL)
	}
}

func TestParse_Bento4Mixed_AudioChannelConfiguration(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	mpd, _ := Parse(content)
	// Audio AdaptationSet is index 2.
	r := mpd.Periods[0].AdaptationSets[2].Representations[0]
	if r.AudioChannelConfiguration == nil {
		t.Fatal("AudioChannelConfiguration is nil, want non-nil")
	}
	if r.AudioChannelConfiguration.SchemeIDURI != "urn:mpeg:mpegB:cicp:ChannelConfiguration" {
		t.Errorf("SchemeIDURI = %q", r.AudioChannelConfiguration.SchemeIDURI)
	}
	if r.AudioChannelConfiguration.Value != "2" {
		t.Errorf("Value = %q, want 2", r.AudioChannelConfiguration.Value)
	}
}

func TestParse_SegmentBase(t *testing.T) {
	mpd := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static">
  <Period>
    <AdaptationSet mimeType="video/mp4" contentType="video">
      <SegmentBase indexRange="0-819">
        <Initialization sourceURL="init.mp4"/>
      </SegmentBase>
      <Representation id="v1" bandwidth="3000000"/>
    </AdaptationSet>
  </Period>
</MPD>`
	m, err := Parse(mpd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Periods) == 0 || len(m.Periods[0].AdaptationSets) == 0 {
		t.Fatal("expected adaptation set")
	}
	as := m.Periods[0].AdaptationSets[0]
	if len(as.Representations) == 0 {
		t.Fatal("expected representation")
	}
	if as.SegmentBase == nil {
		t.Fatal("expected SegmentBase to be parsed on AdaptationSet")
	}
	if as.SegmentBase.IndexRange != "0-819" {
		t.Errorf("IndexRange = %q, want 0-819", as.SegmentBase.IndexRange)
	}
	if as.SegmentBase.Initialization != "init.mp4" {
		t.Errorf("Initialization = %q, want init.mp4", as.SegmentBase.Initialization)
	}
}
