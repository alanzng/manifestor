package hls

import (
	"strings"
	"testing"
)

// ---- Round-trip tests (parse → serialize → re-parse) ----

func TestSerialize_RoundTrip_Bento4Master(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	roundTrip(t, content, 5, 2, 2)
}

func TestSerialize_RoundTrip_ShakaMaster(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/shaka_master.m3u8")
	roundTrip(t, content, 3, 1, 0)
}

func TestSerialize_RoundTrip_AWSMediaConvert(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/aws_mediaconvert_master.m3u8")
	roundTrip(t, content, 4, 1, 0)
}

func TestSerialize_RoundTrip_Bento4MixedCodecs(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	roundTrip(t, content, 9, 1, 9)
}

// roundTrip parses content, serializes it, then re-parses and checks counts.
func roundTrip(t *testing.T, content string, wantVariants, wantAudio, wantIFrames int) {
	t.Helper()
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := Serialize(p)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	p2, err := Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(p2.Variants) != wantVariants {
		t.Errorf("re-parsed variants = %d, want %d", len(p2.Variants), wantVariants)
	}
	if len(p2.AudioTracks) != wantAudio {
		t.Errorf("re-parsed audio tracks = %d, want %d", len(p2.AudioTracks), wantAudio)
	}
	if len(p2.IFrames) != wantIFrames {
		t.Errorf("re-parsed I-frames = %d, want %d", len(p2.IFrames), wantIFrames)
	}
}

// ---- Field preservation through serialize → re-parse ----

func TestSerialize_PreservesVariantFields(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)
	p2, _ := Parse(out)

	orig := p.Variants[0]
	got := p2.Variants[0]

	if got.URI != orig.URI {
		t.Errorf("URI = %q, want %q", got.URI, orig.URI)
	}
	if got.Bandwidth != orig.Bandwidth {
		t.Errorf("Bandwidth = %d, want %d", got.Bandwidth, orig.Bandwidth)
	}
	if got.AverageBandwidth != orig.AverageBandwidth {
		t.Errorf("AverageBandwidth = %d, want %d", got.AverageBandwidth, orig.AverageBandwidth)
	}
	if got.Codecs != orig.Codecs {
		t.Errorf("Codecs = %q, want %q", got.Codecs, orig.Codecs)
	}
	if got.Width != orig.Width || got.Height != orig.Height {
		t.Errorf("Resolution = %dx%d, want %dx%d", got.Width, got.Height, orig.Width, orig.Height)
	}
	if got.FrameRate != orig.FrameRate {
		t.Errorf("FrameRate = %v, want %v", got.FrameRate, orig.FrameRate)
	}
	if got.AudioGroupID != orig.AudioGroupID {
		t.Errorf("AudioGroupID = %q, want %q", got.AudioGroupID, orig.AudioGroupID)
	}
}

func TestSerialize_PreservesAudioTrackFields(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)
	p2, _ := Parse(out)

	orig := p.AudioTracks[0]
	got := p2.AudioTracks[0]

	if got.GroupID != orig.GroupID {
		t.Errorf("GroupID = %q, want %q", got.GroupID, orig.GroupID)
	}
	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Language != orig.Language {
		t.Errorf("Language = %q, want %q", got.Language, orig.Language)
	}
	if got.Default != orig.Default {
		t.Errorf("Default = %v, want %v", got.Default, orig.Default)
	}
	if got.AutoSelect != orig.AutoSelect {
		t.Errorf("AutoSelect = %v, want %v", got.AutoSelect, orig.AutoSelect)
	}
	if got.URI != orig.URI {
		t.Errorf("URI = %q, want %q", got.URI, orig.URI)
	}
}

func TestSerialize_PreservesIFrameFields(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_master.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)
	p2, _ := Parse(out)

	orig := p.IFrames[0]
	got := p2.IFrames[0]

	if got.URI != orig.URI {
		t.Errorf("URI = %q, want %q", got.URI, orig.URI)
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
}

func TestSerialize_PreservesUnicodeAudioName(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)
	p2, _ := Parse(out)

	if p2.AudioTracks[0].Name != "тоҷикӣ; تاجیکی" {
		t.Errorf("Name = %q, want %q", p2.AudioTracks[0].Name, "тоҷикӣ; تاجیکی")
	}
}

// ---- Output format checks ----

func TestSerialize_StartsWithEXTM3U(t *testing.T) {
	p := &MasterPlaylist{Version: 3}
	out, _ := Serialize(p)
	if !strings.HasPrefix(out, "#EXTM3U\n") {
		t.Errorf("output does not start with #EXTM3U\\n: %q", out)
	}
}

func TestSerialize_VersionTag(t *testing.T) {
	p := &MasterPlaylist{Version: 6}
	out, _ := Serialize(p)
	if !strings.Contains(out, "#EXT-X-VERSION:6\n") {
		t.Errorf("missing #EXT-X-VERSION:6 in output:\n%s", out)
	}
}

func TestSerialize_OmitsVersionWhenZero(t *testing.T) {
	p := &MasterPlaylist{}
	out, _ := Serialize(p)
	if strings.Contains(out, "#EXT-X-VERSION") {
		t.Errorf("unexpected #EXT-X-VERSION in output when Version=0:\n%s", out)
	}
}

func TestSerialize_VariantOrderPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)
	p2, _ := Parse(out)

	wantURIs := []string{
		"video-hvc1-1.m3u8", "video-hvc1-2.m3u8", "video-hvc1-3.m3u8",
		"video-hvc1-4.m3u8", "video-hvc1-5.m3u8",
		"video-avc1-1.m3u8", "video-avc1-2.m3u8", "video-avc1-3.m3u8", "video-avc1-4.m3u8",
	}
	for i, want := range wantURIs {
		if p2.Variants[i].URI != want {
			t.Errorf("Variants[%d].URI = %q, want %q", i, p2.Variants[i].URI, want)
		}
	}
}

func TestSerialize_FrameRateFormat(t *testing.T) {
	p := &MasterPlaylist{
		Version: 3,
		Variants: []Variant{
			{URI: "v.m3u8", Bandwidth: 1000000, FrameRate: 29.97},
		},
	}
	out, _ := Serialize(p)
	if !strings.Contains(out, "FRAME-RATE=29.970") {
		t.Errorf("expected FRAME-RATE=29.970 in output:\n%s", out)
	}
}

func TestSerialize_OmitsOptionalFieldsWhenZero(t *testing.T) {
	p := &MasterPlaylist{
		Version: 3,
		Variants: []Variant{
			{URI: "v.m3u8", Bandwidth: 1000000},
		},
	}
	out, _ := Serialize(p)
	for _, unwanted := range []string{"CODECS", "RESOLUTION", "FRAME-RATE", "AVERAGE-BANDWIDTH", "AUDIO=", "SUBTITLES="} {
		if strings.Contains(out, unwanted) {
			t.Errorf("unexpected %q in output when field is zero:\n%s", unwanted, out)
		}
	}
}

func TestSerialize_RawLinesPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/bento4_mixed_codecs.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)

	for _, raw := range p.Raw {
		if !strings.Contains(out, raw) {
			t.Errorf("Raw line %q not found in serialized output", raw)
		}
	}
}

func TestSerialize_SubtitleTrack(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/shaka_master.m3u8")
	p, _ := Parse(content)
	out, _ := Serialize(p)
	p2, _ := Parse(out)

	if len(p2.Subtitles) != 1 {
		t.Fatalf("re-parsed subtitles = %d, want 1", len(p2.Subtitles))
	}
	s := p2.Subtitles[0]
	if s.Language != "en" {
		t.Errorf("Language = %q, want %q", s.Language, "en")
	}
	if s.URI != "subtitles/en/playlist.m3u8" {
		t.Errorf("URI = %q, want %q", s.URI, "subtitles/en/playlist.m3u8")
	}
}

func TestSerialize_ForcedSubtitle(t *testing.T) {
	p := &MasterPlaylist{
		Subtitles: []MediaTrack{
			{Type: "SUBTITLES", GroupID: "subs", Name: "English", Language: "en",
				URI: "subs.m3u8", Forced: true},
		},
		Variants: []Variant{
			{URI: "v.m3u8", Bandwidth: 1000000, SubtitleGroupID: "subs"},
		},
	}
	out, _ := Serialize(p)
	if !strings.Contains(out, "FORCED=YES") {
		t.Errorf("expected FORCED=YES in output:\n%s", out)
	}
	// Re-parse and confirm Forced field.
	p2, _ := Parse(out)
	if !p2.Subtitles[0].Forced {
		t.Error("re-parsed Forced = false, want true")
	}
}

func TestSerialize_RawLines_Direct(t *testing.T) {
	// Directly build a MasterPlaylist with Raw lines and verify they're preserved.
	p := &MasterPlaylist{
		Raw: []string{"#EXT-X-CUSTOM-TAG:value"},
		Variants: []Variant{
			{URI: "v.m3u8", Bandwidth: 1000000},
		},
	}
	out, err := Serialize(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#EXT-X-CUSTOM-TAG:value") {
		t.Errorf("expected raw line in output:\n%s", out)
	}
}

func TestSerialize_MediaTrack_EmptyTypeFallsBackToAudio(t *testing.T) {
	// Type == "" should emit TYPE=AUDIO in output.
	p := &MasterPlaylist{
		AudioTracks: []MediaTrack{
			{Type: "", GroupID: "aud", Name: "English", Language: "en",
				URI: "audio.m3u8", Default: true, AutoSelect: true},
		},
		Variants: []Variant{
			{URI: "v.m3u8", Bandwidth: 1000000, AudioGroupID: "aud"},
		},
	}
	out, err := Serialize(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "TYPE=AUDIO") {
		t.Errorf("expected TYPE=AUDIO in output when Type is empty:\n%s", out)
	}
}

func TestSerialize_HDCPLevel(t *testing.T) {
	// HDCPLevel should be emitted for a variant.
	p := &MasterPlaylist{
		Variants: []Variant{
			{URI: "v.m3u8", Bandwidth: 5000000, Width: 3840, Height: 2160,
				Codecs: "hvc1.1.2.L153.90", HDCPLevel: "TYPE-1"},
		},
	}
	out, err := Serialize(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "HDCP-LEVEL=TYPE-1") {
		t.Errorf("expected HDCP-LEVEL=TYPE-1 in output:\n%s", out)
	}
}
