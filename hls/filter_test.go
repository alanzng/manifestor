package hls

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// ---- Codec filter (F-01) ----

func TestFilter_Codec_H264(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	if len(p.Variants) != 4 {
		t.Errorf("variants = %d, want 4", len(p.Variants))
	}
	for _, v := range p.Variants {
		if !matchesCodec(v.Codecs, "h264") {
			t.Errorf("variant %q has non-h264 codec %q", v.URI, v.Codecs)
		}
	}
}

func TestFilter_Codec_H265(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithCodec("h265"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	if len(p.Variants) != 5 {
		t.Errorf("variants = %d, want 5", len(p.Variants))
	}
}

func TestFilter_Codec_NoMatch_ReturnsErrNoVariantsRemain(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	_, err := Filter(content, WithCodec("vp9"))
	if !errors.Is(err, ErrNoVariantsRemain) {
		t.Errorf("got %v, want ErrNoVariantsRemain", err)
	}
}

func TestFilter_Codec_CaseInsensitiveCodecField(t *testing.T) {
	// bento4_mixed_codecs has uppercase codec like "avc1.64001F" — must still match h264.
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	if len(p.Variants) != 4 {
		t.Errorf("variants = %d, want 4 (uppercase codec not matched)", len(p.Variants))
	}
}

// ---- Resolution filters (F-02, F-03, F-04) ----

func TestFilter_MaxResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	// Keep only variants ≤ 1280x720.
	out, err := Filter(content, WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.Width > 1280 || v.Height > 720 {
			t.Errorf("variant %q resolution %dx%d exceeds max 1280x720", v.URI, v.Width, v.Height)
		}
	}
}

func TestFilter_MinResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	// Keep only variants ≥ 1920x1080.
	out, err := Filter(content, WithMinResolution(1920, 1080))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.Width < 1920 || v.Height < 1080 {
			t.Errorf("variant %q resolution %dx%d below min 1920x1080", v.URI, v.Width, v.Height)
		}
	}
}

func TestFilter_ExactResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithExactResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.Width != 1280 || v.Height != 720 {
			t.Errorf("variant %q resolution %dx%d, want exactly 1280x720", v.URI, v.Width, v.Height)
		}
	}
}

// ---- Bandwidth filters (F-05, F-06) ----

func TestFilter_MaxBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithMaxBandwidth(3000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.Bandwidth > 3000000 {
			t.Errorf("variant %q bandwidth %d exceeds max 3000000", v.URI, v.Bandwidth)
		}
	}
}

func TestFilter_MinBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithMinBandwidth(2000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.Bandwidth < 2000000 {
			t.Errorf("variant %q bandwidth %d below min 2000000", v.URI, v.Bandwidth)
		}
	}
}

// ---- Frame rate filter (F-07) ----

func TestFilter_MaxFrameRate(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	// Mixed codec fixture has 25fps (avc1) and 50fps (hvc1). Keep only ≤ 30fps.
	out, err := Filter(content, WithMaxFrameRate(30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.FrameRate > 30 {
			t.Errorf("variant %q frame rate %.3f exceeds max 30", v.URI, v.FrameRate)
		}
	}
	// Only the 4 avc1 variants at 25fps should survive.
	if len(p.Variants) != 4 {
		t.Errorf("variants = %d, want 4 (25fps only)", len(p.Variants))
	}
}

// ---- Audio language filter (F-08, F-13) ----

func TestFilter_AudioLanguage_FiltersToMatchingTracks(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithAudioLanguage("fr"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	if len(p.AudioTracks) != 1 {
		t.Fatalf("audio tracks = %d, want 1", len(p.AudioTracks))
	}
	if p.AudioTracks[0].Language != "fr" {
		t.Errorf("language = %q, want %q", p.AudioTracks[0].Language, "fr")
	}
}

func TestFilter_AudioLanguage_NoFilter_PreservesAllTracks(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	if len(p.AudioTracks) != 2 {
		t.Errorf("audio tracks = %d, want 2 (F-13: preserved by default)", len(p.AudioTracks))
	}
}

// ---- Custom filter (F-11) ----

func TestFilter_CustomFilter(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	// Keep only the highest-bandwidth variant.
	out, err := Filter(content, WithCustomFilter(func(v *Variant) bool {
		return v.Bandwidth >= 5000000
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.Bandwidth < 5000000 {
			t.Errorf("variant %q bandwidth %d should have been filtered", v.URI, v.Bandwidth)
		}
	}
}

// ---- Composed filters (F-10, TS-03) ----

func TestFilter_ComposedFilters(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content,
		WithCodec("h264"),
		WithMaxResolution(1280, 720),
		WithMaxBandwidth(4000000),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !matchesCodec(v.Codecs, "h264") {
			t.Errorf("variant %q has non-h264 codec", v.URI)
		}
		if v.Width > 1280 || v.Height > 720 {
			t.Errorf("variant %q resolution %dx%d exceeds max", v.URI, v.Width, v.Height)
		}
		if v.Bandwidth > 4000000 {
			t.Errorf("variant %q bandwidth %d exceeds max", v.URI, v.Bandwidth)
		}
	}
}

func TestFilter_AllFiltersOut_ReturnsError(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	_, err := Filter(content,
		WithMaxBandwidth(100), // no variant has bandwidth < 100
	)
	if !errors.Is(err, ErrNoVariantsRemain) {
		t.Errorf("got %v, want ErrNoVariantsRemain", err)
	}
}

// ---- I-frame filtering (F-14) ----

func TestFilter_IFrames_FollowCodecFilter(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, f := range p.IFrames {
		if !matchesCodec(f.Codecs, "h264") {
			t.Errorf("I-frame %q has non-h264 codec %q", f.URI, f.Codecs)
		}
	}
}

func TestFilter_IFrames_FollowResolutionFilter(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, f := range p.IFrames {
		if f.Width > 1280 || f.Height > 720 {
			t.Errorf("I-frame %q resolution %dx%d exceeds max", f.URI, f.Width, f.Height)
		}
	}
}

// ---- URI transformers (T-01, T-02, T-03) ----

func TestFilter_AbsoluteURIs(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	origin := "https://origin.example.com/live/"
	out, err := Filter(content, WithAbsoluteURIs(origin))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !strings.HasPrefix(v.URI, "https://") {
			t.Errorf("variant %q URI not absolute after WithAbsoluteURIs", v.URI)
		}
	}
}

func TestFilter_CDNBaseURL(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	origin := "https://origin.example.com/live/"
	cdn := "https://cdn.cloudfront.net"

	// First make absolute, then rewrite to CDN.
	out, err := Filter(content,
		WithAbsoluteURIs(origin),
		WithCDNBaseURL(cdn),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !strings.HasPrefix(v.URI, "https://cdn.cloudfront.net") {
			t.Errorf("variant %q URI not rewritten to CDN", v.URI)
		}
	}
}

func TestFilter_CDNRewrite_IsIdempotent(t *testing.T) {
	// TS-07: applying CDN rewrite twice must produce the same result.
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	origin := "https://origin.example.com/live/"
	cdn := "https://cdn.cloudfront.net"

	once, _ := Filter(content, WithAbsoluteURIs(origin), WithCDNBaseURL(cdn))
	twice, _ := Filter(once, WithCDNBaseURL(cdn))

	p1, _ := Parse(once)
	p2, _ := Parse(twice)
	for i := range p1.Variants {
		if p1.Variants[i].URI != p2.Variants[i].URI {
			t.Errorf("CDN rewrite not idempotent: %q vs %q",
				p1.Variants[i].URI, p2.Variants[i].URI)
		}
	}
}

func TestFilter_AuthToken(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	origin := "https://origin.example.com/live/"
	out, err := Filter(content,
		WithAbsoluteURIs(origin),
		WithAuthToken("abc123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !strings.Contains(v.URI, "token=abc123") {
			t.Errorf("variant %q URI missing auth token", v.URI)
		}
	}
}

func TestFilter_AuthToken_IsIdempotent(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	origin := "https://origin.example.com/live/"

	once, _ := Filter(content, WithAbsoluteURIs(origin), WithAuthToken("abc123"))
	twice, _ := Filter(once, WithAuthToken("abc123"))

	p1, _ := Parse(once)
	p2, _ := Parse(twice)
	for i := range p1.Variants {
		if p1.Variants[i].URI != p2.Variants[i].URI {
			t.Errorf("token append not idempotent: %q vs %q",
				p1.Variants[i].URI, p2.Variants[i].URI)
		}
	}
}

// ---- Custom transformer (T-04, T-06) ----

func TestFilter_CustomTransformer(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithCustomTransformer(func(v *Variant) {
		v.URI = "rewritten/" + v.URI
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !strings.HasPrefix(v.URI, "rewritten/") {
			t.Errorf("variant %q URI not transformed", v.URI)
		}
	}
}

// ---- Transformer ordering: applied after filter (T-05) ----

func TestFilter_TransformerAppliedOnlyToSurvivors(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	var transformed []string
	out, err := Filter(content,
		WithCodec("h264"),
		WithCustomTransformer(func(v *Variant) {
			transformed = append(transformed, v.URI)
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	// Transformer must have run exactly once per surviving variant.
	if len(transformed) != len(p.Variants) {
		t.Errorf("transformer ran %d times, want %d (survivors only)",
			len(transformed), len(p.Variants))
	}
}

// ---- Round-trip: output is re-parseable (TS-04) ----

func TestFilter_OutputIsReparseable(t *testing.T) {
	fixtures := []string{
		"../testdata/hls/bento4_master.m3u8",
		"../testdata/hls/bento4_mixed_codecs.m3u8",
		"../testdata/hls/shaka_master.m3u8",
		"../testdata/hls/aws_mediaconvert_master.m3u8",
	}
	for _, path := range fixtures {
		t.Run(path, func(t *testing.T) {
			content := mustReadFixture(t, path)
			out, err := Filter(content, WithMaxBandwidth(999999999))
			if err != nil {
				t.Fatalf("filter: %v", err)
			}
			if _, err := Parse(out); err != nil {
				t.Errorf("re-parse failed: %v", err)
			}
		})
	}
}

// ---- matchesCodec unit tests ----

func TestMatchesCodec(t *testing.T) {
	tests := []struct {
		codecs string
		want   string
		match  bool
	}{
		{"avc1.640028,mp4a.40.2", "h264", true},
		{"avc3.640028", "h264", true},
		{"hvc1.1.6.L150.90,mp4a.40.2", "h265", true},
		{"hev1.1.6.L150.90", "h265", true},
		{"vp09.00.10.08", "vp9", true},
		{"vp9", "vp9", true},
		{"av01.0.04M.08", "av1", true},
		{"avc1.640028", "h265", false},
		{"", "h264", false},
		{"mp4a.40.2", "h264", false},
		// Case-insensitive: uppercase codec strings from some encoders.
		{"AVC1.640028,mp4a.40.2", "h264", true},
		{"HVC1.1.2.L120.90", "h265", true},
	}
	for _, tt := range tests {
		got := matchesCodec(tt.codecs, tt.want)
		if got != tt.match {
			t.Errorf("matchesCodec(%q, %q) = %v, want %v", tt.codecs, tt.want, got, tt.match)
		}
	}
}

// ---- Benchmark (TS-08) ----

func BenchmarkFilter_ParseFilterSerialize(b *testing.B) {
	raw, err := os.ReadFile("../testdata/hls/bento4_mixed_codecs.m3u8")
	if err != nil {
		b.Fatal(err)
	}
	content := string(raw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Filter(content, WithCodec("h264"), WithMaxResolution(1920, 1080))
	}
}

// mustReadFixture is defined in parser_test.go; shared across all test files in package.
