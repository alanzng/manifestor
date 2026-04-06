package hls

import (
	"errors"
	"os"
	"testing"
)

// mustReadFixture reads a testdata file and fails the test on error.
func mustReadFixture(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(b)
}

// ---- Parse: basic valid cases ----

func TestParse_Bento4Master(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Version != 6 {
		t.Errorf("Version = %d, want 6", p.Version)
	}
	if len(p.Variants) != 5 {
		t.Errorf("len(Variants) = %d, want 5", len(p.Variants))
	}
	if len(p.AudioTracks) != 2 {
		t.Errorf("len(AudioTracks) = %d, want 2", len(p.AudioTracks))
	}
	if len(p.IFrames) != 2 {
		t.Errorf("len(IFrames) = %d, want 2", len(p.IFrames))
	}
}

func TestParse_ShakaMaster(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/shaka_master.m3u8")
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Variants) != 3 {
		t.Errorf("len(Variants) = %d, want 3", len(p.Variants))
	}
	if len(p.Subtitles) != 1 {
		t.Errorf("len(Subtitles) = %d, want 1", len(p.Subtitles))
	}
}

func TestParse_AWSMediaConvert(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/aws_mediaconvert_master.m3u8")
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Variants) != 4 {
		t.Errorf("len(Variants) = %d, want 4", len(p.Variants))
	}
}

// ---- Parse: variant field correctness ----

func TestParse_VariantFields(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v := p.Variants[0] // first variant: 1080p H.264
	if v.URI != "1080p/index.m3u8" {
		t.Errorf("URI = %q, want %q", v.URI, "1080p/index.m3u8")
	}
	if v.Bandwidth != 5000000 {
		t.Errorf("Bandwidth = %d, want 5000000", v.Bandwidth)
	}
	if v.AverageBandwidth != 4500000 {
		t.Errorf("AverageBandwidth = %d, want 4500000", v.AverageBandwidth)
	}
	if v.Width != 1920 || v.Height != 1080 {
		t.Errorf("Resolution = %dx%d, want 1920x1080", v.Width, v.Height)
	}
	if v.FrameRate != 29.970 {
		t.Errorf("FrameRate = %v, want 29.970", v.FrameRate)
	}
	if v.Codecs != "avc1.640028,mp4a.40.2" {
		t.Errorf("Codecs = %q, want %q", v.Codecs, "avc1.640028,mp4a.40.2")
	}
	if v.AudioGroupID != "audio-en" {
		t.Errorf("AudioGroupID = %q, want %q", v.AudioGroupID, "audio-en")
	}
}

// ---- Parse: audio track field correctness ----

func TestParse_AudioTrackFields(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := p.AudioTracks[0]
	if a.GroupID != "audio-en" {
		t.Errorf("GroupID = %q, want %q", a.GroupID, "audio-en")
	}
	if a.Language != "en" {
		t.Errorf("Language = %q, want %q", a.Language, "en")
	}
	if a.Name != "English" {
		t.Errorf("Name = %q, want %q", a.Name, "English")
	}
	if !a.Default {
		t.Error("Default = false, want true")
	}
	if !a.AutoSelect {
		t.Error("AutoSelect = false, want true")
	}
	if a.URI != "audio/en/index.m3u8" {
		t.Errorf("URI = %q, want %q", a.URI, "audio/en/index.m3u8")
	}
}

// ---- Parse: I-frame stream field correctness ----

func TestParse_IFrameFields(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := p.IFrames[0]
	if f.URI != "1080p/iframe.m3u8" {
		t.Errorf("URI = %q, want %q", f.URI, "1080p/iframe.m3u8")
	}
	if f.Bandwidth != 150000 {
		t.Errorf("Bandwidth = %d, want 150000", f.Bandwidth)
	}
	if f.Width != 1920 || f.Height != 1080 {
		t.Errorf("Resolution = %dx%d, want 1920x1080", f.Width, f.Height)
	}
}

// ---- Parse: error cases ----

func TestParse_EmptyContent(t *testing.T) {
	_, err := Parse("")
	if !errors.Is(err, ErrParseFailure) {
		t.Errorf("got %v, want ErrParseFailure", err)
	}
}

func TestParse_MissingExtM3U(t *testing.T) {
	_, err := Parse("#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:BANDWIDTH=1000\nstream.m3u8")
	if !errors.Is(err, ErrParseFailure) {
		t.Errorf("got %v, want ErrParseFailure", err)
	}
}

func TestParse_MediaPlaylist(t *testing.T) {
	content := "#EXTM3U\n#EXT-X-TARGETDURATION:10\n#EXTINF:9.009,\nsegment.ts\n"
	_, err := Parse(content)
	if !errors.Is(err, ErrNotMasterPlaylist) {
		t.Errorf("got %v, want ErrNotMasterPlaylist", err)
	}
}

// ---- Parse: edge cases ----

func TestParse_CRLFLineEndings(t *testing.T) {
	content := "#EXTM3U\r\n#EXT-X-VERSION:3\r\n#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1280x720\r\nvideo.m3u8\r\n"
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Variants) != 1 {
		t.Errorf("len(Variants) = %d, want 1", len(p.Variants))
	}
	if p.Variants[0].URI != "video.m3u8" {
		t.Errorf("URI = %q, want %q", p.Variants[0].URI, "video.m3u8")
	}
}

func TestParse_UnknownTagsStoredInRaw(t *testing.T) {
	content := "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-CUSTOM-TAG:value\n#EXT-X-STREAM-INF:BANDWIDTH=1000000\nvideo.m3u8\n"
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Raw) != 1 || p.Raw[0] != "#EXT-X-CUSTOM-TAG:value" {
		t.Errorf("Raw = %v, want [\"#EXT-X-CUSTOM-TAG:value\"]", p.Raw)
	}
}

func TestParse_BlankLineBetweenStreamInfAndURI(t *testing.T) {
	// Some encoders insert a blank line between EXT-X-STREAM-INF and the URI.
	content := "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:BANDWIDTH=1000000\n\nvideo.m3u8\n"
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Variants) != 1 || p.Variants[0].URI != "video.m3u8" {
		t.Errorf("expected variant with URI=video.m3u8, got %+v", p.Variants)
	}
}

func TestParse_CodecWithCommaInsideQuotes(t *testing.T) {
	content := "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=5000000,CODECS=\"avc1.640028,mp4a.40.2\",RESOLUTION=1920x1080\nvideo.m3u8\n"
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Variants[0].Codecs != "avc1.640028,mp4a.40.2" {
		t.Errorf("Codecs = %q, want %q", p.Variants[0].Codecs, "avc1.640028,mp4a.40.2")
	}
}

func TestParse_MultipleVariants_PreservesOrder(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, _ := Parse(content)

	want := []string{
		"1080p/index.m3u8",
		"1080p-hevc/index.m3u8",
		"720p/index.m3u8",
		"480p/index.m3u8",
		"360p/index.m3u8",
	}
	for i, v := range p.Variants {
		if v.URI != want[i] {
			t.Errorf("Variants[%d].URI = %q, want %q", i, v.URI, want[i])
		}
	}
}

// ---- parseAttrs unit tests ----

func TestParseAttrs_UnquotedValues(t *testing.T) {
	attrs := parseAttrs("BANDWIDTH=5000000,RESOLUTION=1920x1080")
	if attrs["BANDWIDTH"] != "5000000" {
		t.Errorf("BANDWIDTH = %q, want %q", attrs["BANDWIDTH"], "5000000")
	}
	if attrs["RESOLUTION"] != "1920x1080" {
		t.Errorf("RESOLUTION = %q, want %q", attrs["RESOLUTION"], "1920x1080")
	}
}

func TestParseAttrs_QuotedValues(t *testing.T) {
	attrs := parseAttrs(`GROUP-ID="audio-en",NAME="English"`)
	if attrs["GROUP-ID"] != "audio-en" {
		t.Errorf("GROUP-ID = %q, want %q", attrs["GROUP-ID"], "audio-en")
	}
	if attrs["NAME"] != "English" {
		t.Errorf("NAME = %q, want %q", attrs["NAME"], "English")
	}
}

func TestParseAttrs_QuotedValueWithComma(t *testing.T) {
	attrs := parseAttrs(`CODECS="avc1.640028,mp4a.40.2",BANDWIDTH=5000000`)
	if attrs["CODECS"] != "avc1.640028,mp4a.40.2" {
		t.Errorf("CODECS = %q, want %q", attrs["CODECS"], "avc1.640028,mp4a.40.2")
	}
	if attrs["BANDWIDTH"] != "5000000" {
		t.Errorf("BANDWIDTH = %q, want %q", attrs["BANDWIDTH"], "5000000")
	}
}

func TestParseAttrs_MixedQuotedAndUnquoted(t *testing.T) {
	attrs := parseAttrs(`TYPE=AUDIO,GROUP-ID="audio-en",DEFAULT=YES`)
	if attrs["TYPE"] != "AUDIO" {
		t.Errorf("TYPE = %q, want %q", attrs["TYPE"], "AUDIO")
	}
	if attrs["GROUP-ID"] != "audio-en" {
		t.Errorf("GROUP-ID = %q, want %q", attrs["GROUP-ID"], "audio-en")
	}
	if attrs["DEFAULT"] != "YES" {
		t.Errorf("DEFAULT = %q, want %q", attrs["DEFAULT"], "YES")
	}
}

// ---- parseResolution unit tests ----

func TestParseResolution(t *testing.T) {
	tests := []struct {
		input string
		w, h  int
	}{
		{"1920x1080", 1920, 1080},
		{"1280x720", 1280, 720},
		{"854x480", 854, 480},
		{"640x360", 640, 360},
		{"invalid", 0, 0},
		{"", 0, 0},
	}
	for _, tt := range tests {
		w, h := parseResolution(tt.input)
		if w != tt.w || h != tt.h {
			t.Errorf("parseResolution(%q) = %d×%d, want %d×%d", tt.input, w, h, tt.w, tt.h)
		}
	}
}
