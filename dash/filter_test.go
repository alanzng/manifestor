package dash

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// ---- Codec filter ----

func TestFilter_Codec_H264(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, p := range m.Periods {
		for _, as := range p.AdaptationSets {
			// Codec filter applies to video only; audio sets are preserved as-is.
			if isAudioAdaptationSet(&as) {
				continue
			}
			for _, r := range as.Representations {
				if r.Codecs != "" && !matchesCodec(r.Codecs, "h264") {
					t.Errorf("non-h264 video rep survived: %s (%s)", r.ID, r.Codecs)
				}
			}
		}
	}
}

func TestFilter_Codec_H265(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithCodec("h265"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	videoAS := m.Periods[0].AdaptationSets[0]
	if len(videoAS.Representations) != 5 {
		t.Errorf("h265 reps = %d, want 5", len(videoAS.Representations))
	}
}

func TestFilter_Codec_NoMatch_ReturnsErr(t *testing.T) {
	// A video-only MPD with no audio: filtering for av1 removes all reps.
	videoOnly := `<?xml version="1.0"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011">
  <Period>
    <AdaptationSet mimeType="video/mp4">
      <Representation id="v1" bandwidth="1000000" codecs="avc1.64001F" width="1280" height="720"/>
    </AdaptationSet>
  </Period>
</MPD>`
	_, err := Filter(videoOnly, WithCodec("av1"))
	if !errors.Is(err, ErrNoVariantsRemain) {
		t.Errorf("got %v, want ErrNoVariantsRemain", err)
	}
}

func TestFilter_Codec_CaseInsensitive(t *testing.T) {
	// bento4 fixture has "avc1.64001F" with uppercase — must still match h264.
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	_, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Errorf("unexpected error: %v (uppercase codec should match h264)", err)
	}
}

// ---- Resolution filters ----

func TestFilter_MaxResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Width > 1280 || r.Height > 720 {
			t.Errorf("rep %s (%dx%d) exceeds max resolution", r.ID, r.Width, r.Height)
		}
	}
}

func TestFilter_MinResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMinResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Width < 1280 || r.Height < 720 {
			t.Errorf("rep %s (%dx%d) below min resolution", r.ID, r.Width, r.Height)
		}
	}
}

func TestFilter_ExactResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithExactResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	reps := m.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 1 {
		t.Fatalf("reps = %d, want 1", len(reps))
	}
	if reps[0].Width != 1280 || reps[0].Height != 720 {
		t.Errorf("resolution = %dx%d, want 1280x720", reps[0].Width, reps[0].Height)
	}
}

// ---- Bandwidth filters ----

func TestFilter_MaxBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMaxBandwidth(2000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Bandwidth > 2000000 {
			t.Errorf("rep %s bandwidth %d exceeds max", r.ID, r.Bandwidth)
		}
	}
}

func TestFilter_MinBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMinBandwidth(2000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Bandwidth < 2000000 {
			t.Errorf("rep %s bandwidth %d below min", r.ID, r.Bandwidth)
		}
	}
}

// ---- Frame rate filter ----

func TestFilter_MaxFrameRate(t *testing.T) {
	// bento4: HEVC at 50fps, AVC at 25fps — max 30 should keep only AVC.
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithMaxFrameRate(30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, p := range m.Periods {
		for _, as := range p.AdaptationSets {
			for _, r := range as.Representations {
				fps := parseFrameRate(r.FrameRate)
				if fps > 30 {
					t.Errorf("rep %s frameRate %s exceeds max 30", r.ID, r.FrameRate)
				}
			}
		}
	}
}

func TestFilter_ParseFrameRate_Fraction(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"30", 30},
		{"50", 50},
		{"30000/1001", 29.97002997},
		{"", 0},
		{"bad", 0},
		{"0/1", 0},
	}
	for _, tt := range tests {
		got := parseFrameRate(tt.input)
		// Allow small float tolerance for fractional rates.
		diff := got - tt.want
		if diff < -0.001 || diff > 0.001 {
			t.Errorf("parseFrameRate(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ---- MimeType filter ----

func TestFilter_MimeType_VideoOnly(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMimeType("video/mp4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	// Only video AdaptationSet should survive.
	for _, as := range m.Periods[0].AdaptationSets {
		if as.MimeType == "audio/mp4" {
			t.Error("audio AdaptationSet survived video/mp4 mime filter")
		}
	}
}

// ---- Audio language filter ----

func TestFilter_AudioLanguage(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	out, err := Filter(content, WithAudioLanguage("fr"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, as := range m.Periods[0].AdaptationSets {
		if isAudioAdaptationSet(&as) && as.Lang != "fr" {
			t.Errorf("audio as lang = %q survived, want only fr", as.Lang)
		}
	}
}

func TestFilter_AudioLanguage_PreservesVideo(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	out, err := Filter(content, WithAudioLanguage("en"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	videoCount := 0
	for _, as := range m.Periods[0].AdaptationSets {
		if as.ContentType == "video" {
			videoCount++
		}
	}
	if videoCount != 1 {
		t.Errorf("video adaptation sets = %d, want 1", videoCount)
	}
}

// ---- Custom filter ----

func TestFilter_CustomFilter(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithCustomFilter(func(r *Representation) bool {
		return r.Bandwidth >= 5000000
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Bandwidth < 5000000 {
			t.Errorf("rep %s bandwidth %d below custom filter threshold", r.ID, r.Bandwidth)
		}
	}
}

// ---- Custom transformer ----

func TestFilter_CustomTransformer(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithCustomTransformer(func(r *Representation) {
		r.ID = "transformed-" + r.ID
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if len(r.ID) < 12 || r.ID[:12] != "transformed-" {
			t.Errorf("rep ID %q not transformed", r.ID)
		}
	}
}

// ---- Composed filters ----

func TestFilter_Composed(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	// h264 AND exact 1280x720 — should leave only video-avc1-2.
	out, err := Filter(content, WithCodec("h264"), WithExactResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	reps := m.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 1 {
		t.Fatalf("reps = %d, want 1", len(reps))
	}
	if reps[0].ID != "video-avc1-2" {
		t.Errorf("surviving rep = %q, want video-avc1-2", reps[0].ID)
	}
}

// ---- ErrNoVariantsRemain ----

func TestFilter_ErrNoVariantsRemain_AllFiltered(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	_, err := Filter(content, WithMaxBandwidth(1))
	if !errors.Is(err, ErrNoVariantsRemain) {
		t.Errorf("got %v, want ErrNoVariantsRemain", err)
	}
}

// ---- Output is re-parseable across all fixtures ----

func TestFilter_Reparseable_OnDemand(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMaxBandwidth(3000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := Parse(out); err != nil {
		t.Errorf("re-parse failed: %v", err)
	}
}

func TestFilter_Reparseable_Live(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	out, err := Filter(content, WithAudioLanguage("en"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := Parse(out); err != nil {
		t.Errorf("re-parse failed: %v", err)
	}
}

func TestFilter_Reparseable_Bento4Mixed(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithCodec("h265"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := Parse(out); err != nil {
		t.Errorf("re-parse failed: %v", err)
	}
}

// ---- matchesCodec table ----

func TestMatchesCodec(t *testing.T) {
	tests := []struct {
		field string
		want  string
		match bool
	}{
		{"avc1.640028", "h264", true},
		{"avc1.64001F", "h264", true}, // uppercase hex
		{"avc3.640028", "h264", true},
		{"hvc1.1.2.L120.90", "h265", true},
		{"hev1.1.2.L120.90", "h265", true},
		{"vp09.00.10.08", "vp9", true},
		{"vp9", "vp9", true},
		{"av01.0.04M.08", "av1", true},
		{"avc1.640028", "h265", false},
		{"hvc1.1.2.L120.90", "h264", false},
		{"mp4a.40.2", "h264", false},
		{"", "h264", false},
	}
	for _, tt := range tests {
		got := matchesCodec(tt.field, tt.want)
		if got != tt.match {
			t.Errorf("matchesCodec(%q, %q) = %v, want %v", tt.field, tt.want, got, tt.match)
		}
	}
}

// ---- Benchmark ----

func BenchmarkFilter(b *testing.B) {
	content := mustReadFixtureB(b, "../testdata/dash/bento4_mixed_codecs.mpd")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Filter(content, WithCodec("h264"), WithMaxBandwidth(5000000))
	}
}

func mustReadFixtureB(b *testing.B, path string) string {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("read fixture %s: %v", path, err)
	}
	return string(data)
}

// ---- BaseURL URI transformers ----

func TestFilter_WithAbsoluteURIs(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithAbsoluteURIs("https://cdn.example.com/videos/"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	r := m.Periods[0].AdaptationSets[0].Representations[0]
	if r.BaseURL != "https://cdn.example.com/videos/media-video-hvc1-1.mp4" {
		t.Errorf("BaseURL = %q", r.BaseURL)
	}
}

func TestFilter_WithCDNBaseURL(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	// First make absolute, then rewrite CDN.
	out, err := Filter(content,
		WithAbsoluteURIs("https://origin.example.com/"),
		WithCDNBaseURL("https://cdn.example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	r := m.Periods[0].AdaptationSets[0].Representations[0]
	if !strings.HasPrefix(r.BaseURL, "https://cdn.example.com") {
		t.Errorf("BaseURL not rewritten to CDN: %q", r.BaseURL)
	}
}

func TestFilter_WithAuthToken(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content,
		WithAbsoluteURIs("https://cdn.example.com/"),
		WithAuthToken("abc123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	r := m.Periods[0].AdaptationSets[0].Representations[0]
	if !strings.Contains(r.BaseURL, "token=abc123") {
		t.Errorf("BaseURL missing token: %q", r.BaseURL)
	}
}

func TestFilter_WithCDNBaseURL_RelativeURI(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithCDNBaseURL("https://cdn.example.com/path"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	r := m.Periods[0].AdaptationSets[0].Representations[0]
	if !strings.HasPrefix(r.BaseURL, "https://cdn.example.com") {
		t.Errorf("BaseURL = %q", r.BaseURL)
	}
}

// ---- WithInjectAdaptationSet ----

func TestFilter_WithAbsoluteURIs_NoTrailingSlash(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	// Origin without trailing slash — transformer must still produce a valid absolute URL.
	out, err := Filter(content, WithAbsoluteURIs("https://cdn.example.com/videos"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	r := m.Periods[0].AdaptationSets[0].Representations[0]
	// The origin is treated as a base by appending "/" internally, so the
	// resulting BaseURL must start with "https://cdn.example.com/videos/".
	if !strings.HasPrefix(r.BaseURL, "https://cdn.example.com/videos/") {
		t.Errorf("BaseURL = %q, want prefix https://cdn.example.com/videos/", r.BaseURL)
	}
}

func TestFilter_WithInjectAdaptationSet(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m0, _ := Parse(content)
	originalCount := len(m0.Periods[0].AdaptationSets)

	subtitle := AdaptationSetParams{
		MimeType: "text/vtt",
		Lang:     "en",
		Roles:    []Role{{SchemeIDURI: "urn:mpeg:dash:role:2011", Value: "subtitle"}},
		Representations: []RepresentationParams{
			{ID: "sub-en", Bandwidth: 10000, BaseURL: "https://cdn.example.com/sub-en.vtt"},
		},
	}
	out, err := Filter(content, WithInjectAdaptationSet(subtitle))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)
	got := len(m.Periods[0].AdaptationSets)
	if got != originalCount+1 {
		t.Errorf("AdaptationSets = %d, want %d", got, originalCount+1)
	}
	// Injected set is last.
	injected := m.Periods[0].AdaptationSets[got-1]
	if injected.Lang != "en" {
		t.Errorf("injected Lang = %q, want en", injected.Lang)
	}
	if len(injected.Representations) != 1 {
		t.Fatalf("injected Representations = %d, want 1", len(injected.Representations))
	}
	if injected.Representations[0].BaseURL != "https://cdn.example.com/sub-en.vtt" {
		t.Errorf("injected BaseURL = %q", injected.Representations[0].BaseURL)
	}
}
