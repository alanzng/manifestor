package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
