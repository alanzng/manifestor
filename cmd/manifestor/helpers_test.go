package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// moduleRoot returns the module root directory, derived from the location of
// this source file.
func moduleRoot() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "..", "..")
}

// testFixture returns the absolute path to a testdata file.
func testFixture(parts ...string) string {
	return filepath.Join(append([]string{moduleRoot(), "testdata"}, parts...)...)
}

// ---- parseResolution ----

func TestParseResolution_Valid(t *testing.T) {
	w, h, err := parseResolution("1920x1080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w != 1920 || h != 1080 {
		t.Errorf("got %dx%d, want 1920x1080", w, h)
	}
}

func TestParseResolution_MissingX(t *testing.T) {
	_, _, err := parseResolution("1920")
	if err == nil {
		t.Error("expected error for missing x separator")
	}
}

func TestParseResolution_InvalidWidth(t *testing.T) {
	_, _, err := parseResolution("abcx1080")
	if err == nil {
		t.Error("expected error for invalid width")
	}
}

func TestParseResolution_InvalidHeight(t *testing.T) {
	_, _, err := parseResolution("1920xabc")
	if err == nil {
		t.Error("expected error for invalid height")
	}
}

// ---- writeOutput ----

func TestWriteOutput_Stdout(t *testing.T) {
	// Writing to "" goes to stdout — just verify no error.
	err := writeOutput("", "#EXTM3U\n")
	if err != nil {
		t.Errorf("unexpected error writing to stdout: %v", err)
	}
}

func TestWriteOutput_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.m3u8")
	content := "#EXTM3U\n#EXT-X-VERSION:3\n"
	if err := writeOutput(path, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", data, content)
	}
}

// ---- buildHLS ----

func TestBuildHLS_SimpleVariant(t *testing.T) {
	req := &buildRequest{
		Version: 3,
		Variants: []variantJSON{
			{URI: "video-720p.m3u8", Bandwidth: 3000000, Codecs: "avc1.640028", Width: 1280, Height: 720},
		},
	}
	out, err := buildHLS(req, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "#EXTM3U") {
		t.Errorf("expected #EXTM3U, got: %s", out)
	}
	if !strings.Contains(out, "video-720p.m3u8") {
		t.Errorf("expected variant URI in output")
	}
}

func TestBuildHLS_WithVersionFlag(t *testing.T) {
	req := &buildRequest{
		Variants: []variantJSON{
			{URI: "v.m3u8", Bandwidth: 1000000},
		},
	}
	out, err := buildHLS(req, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#EXT-X-VERSION:6") {
		t.Errorf("expected version 6 in output, got:\n%s", out)
	}
}

func TestBuildHLS_WithAudioAndSubtitles(t *testing.T) {
	req := &buildRequest{
		Version: 3,
		AudioTracks: []audioTrackJSON{
			{GroupID: "audio-en", Name: "English", Language: "en", Default: true, AutoSelect: true},
		},
		Subtitles: []subtitleTrackJSON{
			{GroupID: "subs", Name: "Vietnamese", Language: "vi", URI: "vi.m3u8"},
		},
		Variants: []variantJSON{
			{URI: "v.m3u8", Bandwidth: 3000000, AudioGroupID: "audio-en", SubtitleGroupID: "subs"},
		},
	}
	out, err := buildHLS(req, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "TYPE=AUDIO") {
		t.Errorf("expected AUDIO media track")
	}
	if !strings.Contains(out, "TYPE=SUBTITLES") {
		t.Errorf("expected SUBTITLES media track")
	}
}

func TestBuildHLS_WithIFrames(t *testing.T) {
	req := &buildRequest{
		Variants: []variantJSON{
			{URI: "v.m3u8", Bandwidth: 3000000},
		},
		IFrames: []iframeJSON{
			{URI: "iframe.m3u8", Bandwidth: 500000, Codecs: "avc1.640028", Width: 1280, Height: 720},
		},
	}
	out, err := buildHLS(req, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "EXT-X-I-FRAME-STREAM-INF") {
		t.Errorf("expected I-frame stream in output")
	}
}

func TestBuildHLS_EmptyVariants(t *testing.T) {
	req := &buildRequest{}
	_, err := buildHLS(req, 0)
	if err == nil {
		t.Error("expected error for empty variants")
	}
}

// ---- buildDASH ----

func TestBuildDASH_SimpleVideo(t *testing.T) {
	req := &buildRequest{
		Profile:  "isoff-on-demand",
		Duration: "PT4M0.00S",
		AdaptationSets: []adaptationSetJSON{
			{
				ContentType: "video",
				MimeType:    "video/mp4",
				Representations: []representationJSON{
					{ID: "v1", Bandwidth: 5000000, Codecs: "avc1.640028", Width: 1920, Height: 1080},
				},
			},
		},
	}
	out, err := buildDASH(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<MPD") {
		t.Errorf("expected <MPD in output")
	}
	if !strings.Contains(out, "avc1.640028") {
		t.Errorf("expected codec in output")
	}
}

func TestBuildDASH_WithSegmentTemplate(t *testing.T) {
	req := &buildRequest{
		Profile: "isoff-live",
		AdaptationSets: []adaptationSetJSON{
			{
				ContentType: "video",
				MimeType:    "video/mp4",
				SegmentTemplate: &segmentTemplateJSON{
					Initialization: "$RepresentationID$/init.mp4",
					Media:          "$RepresentationID$/$Number$.m4s",
					Timescale:      90000,
					Duration:       270000,
					StartNumber:    1,
				},
				Representations: []representationJSON{
					{ID: "v1", Bandwidth: 3000000, Codecs: "avc1.640028"},
				},
			},
		},
	}
	out, err := buildDASH(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "SegmentTemplate") {
		t.Errorf("expected SegmentTemplate in output")
	}
}

func TestBuildDASH_WithSegmentBase(t *testing.T) {
	req := &buildRequest{
		Profile: "isoff-on-demand",
		AdaptationSets: []adaptationSetJSON{
			{
				ContentType: "video",
				MimeType:    "video/mp4",
				SegmentBase: &segmentBaseJSON{
					IndexRange:     "0-819",
					Initialization: "0-499",
				},
				Representations: []representationJSON{
					{ID: "v1", Bandwidth: 3000000},
				},
			},
		},
	}
	out, err := buildDASH(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<MPD") {
		t.Errorf("expected <MPD in output, got:\n%s", out)
	}
}

// ---- filterManifest ----

func TestFilterManifest_CodecH264(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	var buf bytes.Buffer
	err := filterManifest([]string{"--input", fixture, "--codec", "h264"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "avc1") {
		t.Errorf("expected avc1 in output, got:\n%s", buf.String())
	}
}

func TestFilterManifest_NoArgs(t *testing.T) {
	err := filterManifest([]string{}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when no args provided")
	}
}

func TestFilterManifest_BothURLAndInput(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--url", "http://example.com/test.m3u8", "--input", fixture}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when both --url and --input provided")
	}
}

func TestFilterManifest_MaxRes(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--max-res", "1920x1080"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_MinRes(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--min-res", "640x360"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_MaxBw(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--max-bw", "5000000"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_MinBw(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--min-bw", "100000"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_FPS(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--fps", "60"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_CDN(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--cdn", "https://cdn.example.com"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_Token(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--token", "abc123"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_Lang(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--lang", "tg"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_Origin(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--origin", "https://example.com/path"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFilterManifest_InvalidMaxRes(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--max-res", "bad"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for invalid --max-res")
	}
}

func TestFilterManifest_InvalidMinRes(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	err := filterManifest([]string{"--input", fixture, "--min-res", "bad"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for invalid --min-res")
	}
}

func TestFilterManifest_CodecFiltersEverything(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	// av1 codec doesn't exist in the fixture, so all variants are filtered out.
	err := filterManifest([]string{"--input", fixture, "--codec", "av1"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when all variants filtered out")
	}
	if !strings.Contains(err.Error(), "no variants") {
		t.Errorf("expected 'no variants' in error, got: %v", err)
	}
}

func TestFilterManifest_OutputFile(t *testing.T) {
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.m3u8")
	err := filterManifest([]string{"--input", fixture, "--output", outPath}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "#EXTM3U") {
		t.Errorf("expected #EXTM3U in output file, got:\n%s", data)
	}
}

// ---- buildManifest ----

func TestBuildManifest_FormatHLS(t *testing.T) {
	fixture := testFixture("build", "hls_simple.json")
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "hls", "--variants", fixture}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "#EXTM3U") {
		t.Errorf("expected #EXTM3U in output, got:\n%s", buf.String())
	}
}

func TestBuildManifest_FormatDASH(t *testing.T) {
	fixture := testFixture("build", "dash_simple.json")
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "dash", "--variants", fixture}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "<MPD") {
		t.Errorf("expected <MPD in output, got:\n%s", buf.String())
	}
}

func TestBuildManifest_NoFormat(t *testing.T) {
	fixture := testFixture("build", "hls_simple.json")
	err := buildManifest([]string{"--variants", fixture}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when --format not provided")
	}
}

func TestBuildManifest_NoVariants(t *testing.T) {
	err := buildManifest([]string{"--format", "hls"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error when --variants not provided")
	}
}

func TestBuildManifest_NonexistentVariantsFile(t *testing.T) {
	err := buildManifest([]string{"--format", "hls", "--variants", "/nonexistent/path/file.json"}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for nonexistent variants file")
	}
}

func TestBuildManifest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	badJSON := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := buildManifest([]string{"--format", "hls", "--variants", badJSON}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBuildManifest_UnknownFormat(t *testing.T) {
	fixture := testFixture("build", "hls_simple.json")
	err := buildManifest([]string{"--format", "xml", "--variants", fixture}, &bytes.Buffer{})
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestBuildManifest_HLSWithCDN(t *testing.T) {
	// Write a temp JSON file with absolute URIs so CDN rewriting has effect.
	dir := t.TempDir()
	variantsFile := filepath.Join(dir, "variants.json")
	jsonContent := `{"variants":[{"uri":"https://origin.example.com/video-720p.m3u8","bandwidth":3000000}]}`
	if err := os.WriteFile(variantsFile, []byte(jsonContent), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "hls", "--variants", variantsFile, "--cdn", "https://cdn.example.com"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cdn.example.com") {
		t.Errorf("expected cdn.example.com in output, got:\n%s", buf.String())
	}
}

func TestBuildManifest_DASHWithDuration(t *testing.T) {
	fixture := testFixture("build", "dash_simple.json")
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "dash", "--variants", fixture, "--duration", "PT5M"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "<MPD") {
		t.Errorf("expected <MPD in output, got:\n%s", buf.String())
	}
}

func TestBuildManifest_HLSOutputFile(t *testing.T) {
	fixture := testFixture("build", "hls_simple.json")
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.m3u8")
	err := buildManifest([]string{"--format", "hls", "--variants", fixture, "--output", outPath}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "#EXTM3U") {
		t.Errorf("expected #EXTM3U in output file, got:\n%s", data)
	}
}

func TestBuildManifest_WithCDN(t *testing.T) {
	// Use absolute URIs so CDN rewrite has effect.
	spec := `{"format":"hls","variants":[{"uri":"https://origin.example.com/v.m3u8","bandwidth":3000000}]}`
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	_, _ = f.WriteString(spec)
	_ = f.Close()
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "hls", "--variants", f.Name(), "--cdn", "https://cdn.example.com"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "cdn.example.com") {
		t.Errorf("expected CDN rewrite in output, got:\n%s", buf.String())
	}
}

func TestBuildManifest_DASHWithProfile(t *testing.T) {
	fixture := testFixture("build", "dash_simple.json")
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "dash", "--variants", fixture, "--profile", "isoff-live"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "<MPD") {
		t.Errorf("expected <MPD in output")
	}
}

func TestBuildManifest_WithToken(t *testing.T) {
	spec := `{"format":"hls","variants":[{"uri":"https://origin.example.com/v.m3u8","bandwidth":3000000}]}`
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	_, _ = f.WriteString(spec)
	_ = f.Close()
	var buf bytes.Buffer
	err := buildManifest([]string{"--format", "hls", "--variants", f.Name(), "--token", "mytoken"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "mytoken") {
		t.Errorf("expected token in output, got:\n%s", buf.String())
	}
}

func TestFilterManifest_FromURL(t *testing.T) {
	// Serve a small HLS fixture via httptest and filter via --url.
	fixture := testFixture("hls", "bento4_mixed_codecs.m3u8")
	content, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write(content)
	}))
	defer ts.Close()

	var buf bytes.Buffer
	err = filterManifest([]string{"--url", ts.URL + "/test.m3u8", "--codec", "h264"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "avc1") {
		t.Errorf("expected avc1 codec in output, got:\n%s", buf.String())
	}
}
