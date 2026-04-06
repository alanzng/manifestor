package manifest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alanzng/manifestor/dash"
	"github.com/alanzng/manifestor/hls"
)

// ---- Detect ----

func TestDetect_HLS(t *testing.T) {
	f, err := Detect("#EXTM3U\n#EXT-X-VERSION:3\n")
	if err != nil || f != FormatHLS {
		t.Errorf("Detect(HLS) = %v, %v; want FormatHLS, nil", f, err)
	}
}

func TestDetect_HLS_WithLeadingWhitespace(t *testing.T) {
	f, err := Detect("  \n#EXTM3U\n")
	if err != nil || f != FormatHLS {
		t.Errorf("Detect(HLS with whitespace) = %v, %v; want FormatHLS, nil", f, err)
	}
}

func TestDetect_DASH_XMLDeclaration(t *testing.T) {
	f, err := Detect(`<?xml version="1.0"?><MPD xmlns="urn:mpeg:dash:schema:mpd:2011"/>`)
	if err != nil || f != FormatDASH {
		t.Errorf("Detect(DASH xml) = %v, %v; want FormatDASH, nil", f, err)
	}
}

func TestDetect_DASH_MPDTag(t *testing.T) {
	f, err := Detect(`<MPD xmlns="urn:mpeg:dash:schema:mpd:2011"/>`)
	if err != nil || f != FormatDASH {
		t.Errorf("Detect(DASH MPD tag) = %v, %v; want FormatDASH, nil", f, err)
	}
}

func TestDetect_ErrInvalidFormat(t *testing.T) {
	_, err := Detect("not a manifest")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("got %v, want ErrInvalidFormat", err)
	}
}

func TestDetect_ErrInvalidFormat_Empty(t *testing.T) {
	_, err := Detect("")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("got %v, want ErrInvalidFormat", err)
	}
}

// ---- Filter (HLS) ----

func TestFilter_HLS_Codec(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "#EXTM3U") {
		t.Error("HLS output does not start with #EXTM3U")
	}
	if strings.Contains(out, "hvc1") {
		t.Error("HEVC variant survived h264 codec filter")
	}
}

func TestFilter_HLS_MaxBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithMaxBandwidth(2000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.Bandwidth > 2000000 {
			t.Errorf("variant %s bandwidth %d exceeds max", v.URI, v.Bandwidth)
		}
	}
}

func TestFilter_HLS_ErrNoVariantsRemain(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	_, err := Filter(content, WithMaxBandwidth(1))
	if !errors.Is(err, hls.ErrNoVariantsRemain) {
		t.Errorf("got %v, want hls.ErrNoVariantsRemain", err)
	}
}

// ---- Filter (DASH) ----

func TestFilter_DASH_Codec(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithCodec("h264"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<?xml") {
		t.Error("DASH output does not contain XML declaration")
	}
	if strings.Contains(out, "hvc1") {
		t.Error("HEVC representation survived h264 codec filter")
	}
}

func TestFilter_DASH_AudioLanguage(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_live.mpd")
	out, err := Filter(content, WithAudioLanguage("fr"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	for _, as := range m.Periods[0].AdaptationSets {
		if strings.EqualFold(as.ContentType, "audio") && as.Lang != "fr" {
			t.Errorf("audio lang %q survived fr filter", as.Lang)
		}
	}
}

func TestFilter_DASH_ErrNoVariantsRemain(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	_, err := Filter(content, WithMaxBandwidth(1))
	if !errors.Is(err, dash.ErrNoVariantsRemain) {
		t.Errorf("got %v, want dash.ErrNoVariantsRemain", err)
	}
}

// ---- Filter — invalid content ----

func TestFilter_ErrInvalidFormat(t *testing.T) {
	_, err := Filter("not a manifest")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("got %v, want ErrInvalidFormat", err)
	}
}

// ---- FilterFromFile ----

func TestFilterFromFile_HLS(t *testing.T) {
	path := filepath.Join("..", "testdata", "hls", "bento4_master.m3u8")
	out, err := FilterFromFile(path, WithMaxBandwidth(5000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "#EXTM3U") {
		t.Error("output does not start with #EXTM3U")
	}
}

func TestFilterFromFile_DASH(t *testing.T) {
	path := filepath.Join("..", "testdata", "dash", "isoff_ondemand.mpd")
	out, err := FilterFromFile(path, WithMaxBandwidth(5000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<?xml") {
		t.Error("output does not contain XML declaration")
	}
}

func TestFilterFromFile_NotFound(t *testing.T) {
	_, err := FilterFromFile("/does/not/exist.m3u8")
	if !errors.Is(err, ErrFetchFailed) {
		t.Errorf("got %v, want ErrFetchFailed", err)
	}
}

// ---- FilterFromURL ----

func TestFilterFromURL_HLS(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	defer srv.Close()

	out, err := FilterFromURL(srv.URL+"/master.m3u8", WithMaxBandwidth(5000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "#EXTM3U") {
		t.Error("output does not start with #EXTM3U")
	}
}

func TestFilterFromURL_HTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := FilterFromURL(srv.URL + "/missing.m3u8")
	if !errors.Is(err, ErrFetchFailed) {
		t.Errorf("got %v, want ErrFetchFailed", err)
	}
}

// ---- Build (HLS) ----

func TestBuild_HLS(t *testing.T) {
	out, err := Build(FormatHLS,
		WithHLSVariant(hls.VariantParams{URI: "hd.m3u8", Bandwidth: 5000000, Width: 1920, Height: 1080}),
		WithHLSVariant(hls.VariantParams{URI: "sd.m3u8", Bandwidth: 1500000, Width: 1280, Height: 720}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, err := hls.Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(p.Variants) != 2 {
		t.Errorf("variants = %d, want 2", len(p.Variants))
	}
}

func TestBuild_HLS_WithVersion(t *testing.T) {
	out, _ := Build(FormatHLS,
		WithHLSVersion(6),
		WithHLSVariant(hls.VariantParams{URI: "v.m3u8", Bandwidth: 1000000}),
	)
	if !strings.Contains(out, "#EXT-X-VERSION:6") {
		t.Errorf("expected #EXT-X-VERSION:6 in output:\n%s", out)
	}
}

func TestBuild_HLS_WithAudioAndSubtitles(t *testing.T) {
	out, err := Build(FormatHLS,
		WithHLSAudioTrack(hls.AudioTrackParams{GroupID: "audio", Name: "English", Language: "en", URI: "audio.m3u8", Default: true, AutoSelect: true}),
		WithHLSSubtitleTrack(hls.SubtitleTrackParams{GroupID: "subs", Name: "English", Language: "en", URI: "subs.m3u8"}),
		WithHLSVariant(hls.VariantParams{URI: "v.m3u8", Bandwidth: 5000000, AudioGroupID: "audio", SubtitleGroupID: "subs"}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	if len(p.AudioTracks) != 1 {
		t.Errorf("audio tracks = %d, want 1", len(p.AudioTracks))
	}
	if len(p.Subtitles) != 1 {
		t.Errorf("subtitles = %d, want 1", len(p.Subtitles))
	}
}

// ---- Build (DASH) ----

func TestBuild_DASH(t *testing.T) {
	out, err := Build(FormatDASH,
		WithDASHConfig(dash.MPDConfig{
			Profile:  "urn:mpeg:dash:profile:isoff-on-demand:2011",
			Duration: "PT4M0S",
		}),
		WithDASHAdaptationSet(dash.AdaptationSetParams{
			MimeType: "video/mp4",
			Representations: []dash.RepresentationParams{
				{ID: "v1", Bandwidth: 5000000, Codecs: "avc1.640028", Width: 1920, Height: 1080},
			},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, err := dash.Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(m.Periods[0].AdaptationSets) != 1 {
		t.Errorf("adaptation sets = %d, want 1", len(m.Periods[0].AdaptationSets))
	}
}

func TestBuild_InvalidFormat(t *testing.T) {
	_, err := Build(Format(99))
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("got %v, want ErrInvalidFormat", err)
	}
}

// ---- Filter HLS — resolution / framerate / CDN / token options (toHLSOpts coverage) ----

func TestFilter_HLS_MinBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithMinBandwidth(3000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.Bandwidth < 3000000 {
			t.Errorf("variant %s bandwidth %d below min", v.URI, v.Bandwidth)
		}
	}
}

func TestFilter_HLS_MaxResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.Width > 1280 || v.Height > 720 {
			t.Errorf("variant %s resolution %dx%d exceeds max", v.URI, v.Width, v.Height)
		}
	}
}

func TestFilter_HLS_MinResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithMinResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.Width < 1280 || v.Height < 720 {
			t.Errorf("variant %s resolution %dx%d below min", v.URI, v.Width, v.Height)
		}
	}
}

func TestFilter_HLS_ExactResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithExactResolution(1920, 1080))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.Width != 1920 || v.Height != 1080 {
			t.Errorf("variant %s resolution %dx%d != 1920x1080", v.URI, v.Width, v.Height)
		}
	}
}

func TestFilter_HLS_MaxFrameRate(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithMaxFrameRate(30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.FrameRate > 30 {
			t.Errorf("variant %s frameRate %.3f exceeds max 30", v.URI, v.FrameRate)
		}
	}
}

func TestFilter_HLS_AudioLanguage(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content, WithAudioLanguage("en"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, t2 := range p.AudioTracks {
		if !strings.EqualFold(t2.Language, "en") {
			t.Errorf("audio track lang %q survived en filter", t2.Language)
		}
	}
}

func TestFilter_HLS_CDNBaseURL(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content,
		WithAbsoluteURIs("https://origin.example.com/"),
		WithCDNBaseURL("https://cdn.example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cdn.example.com") {
		t.Error("CDN rewrite not applied")
	}
}

func TestFilter_HLS_AuthToken(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content,
		WithAbsoluteURIs("https://origin.example.com/"),
		WithAuthToken("abc123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "token=abc123") {
		t.Error("auth token not appended")
	}
}

// ---- Filter DASH — resolution / framerate / mime / CDN options (toDASHOpts coverage) ----

func TestFilter_DASH_MaxResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Width > 1280 || r.Height > 720 {
			t.Errorf("rep %s resolution %dx%d exceeds max", r.ID, r.Width, r.Height)
		}
	}
}

func TestFilter_DASH_MinResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMinResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Width < 1280 || r.Height < 720 {
			t.Errorf("rep %s resolution %dx%d below min", r.ID, r.Width, r.Height)
		}
	}
}

func TestFilter_DASH_ExactResolution(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithExactResolution(1920, 1080))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	reps := m.Periods[0].AdaptationSets[0].Representations
	if len(reps) != 1 || reps[0].Width != 1920 {
		t.Errorf("expected exactly 1920x1080 rep, got %d reps", len(reps))
	}
}

func TestFilter_DASH_MinBandwidth(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMinBandwidth(2000000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	for _, r := range m.Periods[0].AdaptationSets[0].Representations {
		if r.Bandwidth < 2000000 {
			t.Errorf("rep %s bandwidth %d below min", r.ID, r.Bandwidth)
		}
	}
}

func TestFilter_DASH_MaxFrameRate(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	out, err := Filter(content, WithMaxFrameRate(30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	for _, as := range m.Periods[0].AdaptationSets {
		for _, r := range as.Representations {
			if r.FrameRate == "50" {
				t.Errorf("rep %s frameRate 50 survived max 30 filter", r.ID)
			}
		}
	}
}

func TestFilter_DASH_MimeType(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content, WithMimeType("video/mp4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	for _, as := range m.Periods[0].AdaptationSets {
		if strings.HasPrefix(as.MimeType, "audio/") {
			t.Error("audio adaptation set survived video/mp4 mime filter")
		}
	}
}

func TestFilter_DASH_CDNBaseURL(t *testing.T) {
	// CDN rewrite registered but DASH reps have no URI — should not error.
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	_, err := Filter(content, WithCDNBaseURL("https://cdn.example.com"))
	if err != nil {
		t.Errorf("unexpected error with CDN option on DASH: %v", err)
	}
}

func TestFilter_DASH_AbsoluteURIs(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	_, err := Filter(content, WithAbsoluteURIs("https://origin.example.com/"))
	if err != nil {
		t.Errorf("unexpected error with AbsoluteURIs option on DASH: %v", err)
	}
}

func TestFilter_DASH_AuthToken(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	_, err := Filter(content, WithAuthToken("tok123"))
	if err != nil {
		t.Errorf("unexpected error with AuthToken option on DASH: %v", err)
	}
}

// ---- Format-specific options ignored on the other format ----
// These exercise hlsOnlyOption.dashOption() == false and dashOnlyOption.hlsOption() == false.

func TestFilter_HLS_IgnoresDASHOnlyOptions(t *testing.T) {
	// DASH-only options (WithDASHConfig, WithDASHAdaptationSet) should be silently
	// ignored when filtering HLS content — no error, HLS output returned.
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	out, err := Filter(content,
		WithDASHConfig(dash.MPDConfig{Profile: "urn:mpeg:dash:profile:isoff-on-demand:2011"}),
		WithMaxBandwidth(10000000),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "#EXTM3U") {
		t.Error("output is not HLS")
	}
}

func TestFilter_DASH_IgnoresHLSOnlyOptions(t *testing.T) {
	// HLS-only options (WithHLSVariant, WithHLSVersion) should be silently
	// ignored when filtering DASH content — no error, DASH output returned.
	content := mustReadFixture(t, "../testdata/dash/isoff_ondemand.mpd")
	out, err := Filter(content,
		WithHLSVariant(hls.VariantParams{URI: "v.m3u8", Bandwidth: 1000000}),
		WithMaxBandwidth(10000000),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<?xml") {
		t.Error("output is not DASH XML")
	}
}

// ---- Build HLS — IFrameStream (applyHLSBuildOption coverage) ----

func TestBuild_HLS_WithIFrameStream(t *testing.T) {
	out, err := Build(FormatHLS,
		WithHLSVariant(hls.VariantParams{URI: "v.m3u8", Bandwidth: 5000000}),
		WithHLSIFrameStream(hls.IFrameParams{URI: "iframe.m3u8", Bandwidth: 200000}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#EXT-X-I-FRAME-STREAM-INF") {
		t.Error("expected #EXT-X-I-FRAME-STREAM-INF in output")
	}
}

// ---- helpers ----

func mustReadFixture(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(b)
}

// ---- Inject options via unified API ----

func TestFilter_WithHLSInjectVariant(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	p0, _ := hls.Parse(content)
	original := len(p0.Variants)

	out, err := Filter(content, WithHLSInjectVariant(hls.VariantParams{
		URI:       "https://cdn.example.com/4k.m3u8",
		Bandwidth: 8000000,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	if len(p.Variants) != original+1 {
		t.Errorf("Variants = %d, want %d", len(p.Variants), original+1)
	}
}

func TestFilter_WithHLSInjectAudioTrack(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	p0, _ := hls.Parse(content)
	original := len(p0.AudioTracks)

	out, err := Filter(content, WithHLSInjectAudioTrack(hls.AudioTrackParams{
		GroupID:  "audio",
		Name:     "French",
		Language: "fr",
		URI:      "fr/audio.m3u8",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	if len(p.AudioTracks) != original+1 {
		t.Errorf("AudioTracks = %d, want %d", len(p.AudioTracks), original+1)
	}
}

func TestFilter_WithHLSInjectSubtitle(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	p0, _ := hls.Parse(content)
	original := len(p0.Subtitles)

	out, err := Filter(content, WithHLSInjectSubtitle(hls.SubtitleTrackParams{
		GroupID:  "subs",
		Name:     "English",
		Language: "en",
		URI:      "en/subs.m3u8",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	if len(p.Subtitles) != original+1 {
		t.Errorf("Subtitles = %d, want %d", len(p.Subtitles), original+1)
	}
}

func TestFilter_WithDASHInjectAdaptationSet(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	m0, _ := dash.Parse(content)
	original := len(m0.Periods[0].AdaptationSets)

	out, err := Filter(content, WithDASHInjectAdaptationSet(dash.AdaptationSetParams{
		MimeType: "text/vtt",
		Lang:     "en",
		Representations: []dash.RepresentationParams{
			{ID: "sub-en", Bandwidth: 10000, BaseURL: "https://cdn.example.com/sub-en.vtt"},
		},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := dash.Parse(out)
	got := len(m.Periods[0].AdaptationSets)
	if got != original+1 {
		t.Errorf("AdaptationSets = %d, want %d", got, original+1)
	}
}

func TestFilter_HLSInjectIgnoredForDASH(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	// HLS-only inject option must be silently ignored for DASH content.
	_, err := Filter(content, WithHLSInjectVariant(hls.VariantParams{
		URI: "https://cdn.example.com/4k.m3u8", Bandwidth: 8000000,
	}))
	if err != nil {
		t.Errorf("unexpected error for DASH content with HLS inject option: %v", err)
	}
}

func TestFilter_DASHInjectIgnoredForHLS(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	// DASH-only inject option must be silently ignored for HLS content.
	_, err := Filter(content, WithDASHInjectAdaptationSet(dash.AdaptationSetParams{
		MimeType: "text/vtt",
		Representations: []dash.RepresentationParams{
			{ID: "sub-en", Bandwidth: 10000},
		},
	}))
	if err != nil {
		t.Errorf("unexpected error for HLS content with DASH inject option: %v", err)
	}
}

func TestFilter_WithHLSVariantSubtitleGroup(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	out, err := Filter(content, WithHLSVariantSubtitleGroup("subs"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := hls.Parse(out)
	for _, v := range p.Variants {
		if v.SubtitleGroupID != "subs" {
			t.Errorf("variant %s SubtitleGroupID = %q, want subs", v.URI, v.SubtitleGroupID)
		}
	}
}

func TestFilter_WithHLSVariantSubtitleGroup_IgnoredForDASH(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/bento4_mixed_codecs.mpd")
	_, err := Filter(content, WithHLSVariantSubtitleGroup("subs"))
	if err != nil {
		t.Errorf("unexpected error for DASH content with HLS subtitle group option: %v", err)
	}
}
