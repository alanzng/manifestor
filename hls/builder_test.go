package hls

import (
	"errors"
	"strings"
	"testing"
)

// ---- Validation errors ----

func TestBuild_ErrEmptyVariantList(t *testing.T) {
	_, err := NewMasterBuilder().Build()
	if !errors.Is(err, ErrEmptyVariantList) {
		t.Errorf("got %v, want ErrEmptyVariantList", err)
	}
}

func TestBuild_ErrInvalidVariant_MissingURI(t *testing.T) {
	_, err := NewMasterBuilder().
		AddVariant(VariantParams{Bandwidth: 1000000}).
		Build()
	if !errors.Is(err, ErrInvalidVariant) {
		t.Errorf("got %v, want ErrInvalidVariant", err)
	}
}

func TestBuild_ErrInvalidVariant_MissingBandwidth(t *testing.T) {
	_, err := NewMasterBuilder().
		AddVariant(VariantParams{URI: "v.m3u8"}).
		Build()
	if !errors.Is(err, ErrInvalidVariant) {
		t.Errorf("got %v, want ErrInvalidVariant", err)
	}
}

func TestBuild_ErrOrphanedGroupID_Audio(t *testing.T) {
	_, err := NewMasterBuilder().
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 1000000, AudioGroupID: "audio-group"}).
		Build()
	if !errors.Is(err, ErrOrphanedGroupID) {
		t.Errorf("got %v, want ErrOrphanedGroupID", err)
	}
}

func TestBuild_ErrOrphanedGroupID_Subtitle(t *testing.T) {
	_, err := NewMasterBuilder().
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 1000000, SubtitleGroupID: "subs"}).
		Build()
	if !errors.Is(err, ErrOrphanedGroupID) {
		t.Errorf("got %v, want ErrOrphanedGroupID", err)
	}
}

// ---- Minimal valid build ----

func TestBuild_MinimalVariant(t *testing.T) {
	out, err := NewMasterBuilder().
		AddVariant(VariantParams{URI: "video.m3u8", Bandwidth: 5000000}).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "#EXTM3U\n") {
		t.Error("output does not start with #EXTM3U")
	}
	if !strings.Contains(out, "video.m3u8") {
		t.Error("output missing variant URI")
	}
	if !strings.Contains(out, "BANDWIDTH=5000000") {
		t.Error("output missing BANDWIDTH")
	}
}

// ---- Version ----

func TestBuild_DefaultVersionIs3(t *testing.T) {
	out, _ := NewMasterBuilder().
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 1000000}).
		Build()
	if !strings.Contains(out, "#EXT-X-VERSION:3\n") {
		t.Errorf("expected #EXT-X-VERSION:3, got:\n%s", out)
	}
}

func TestBuild_SetVersion(t *testing.T) {
	out, _ := NewMasterBuilder().
		SetVersion(6).
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 1000000}).
		Build()
	if !strings.Contains(out, "#EXT-X-VERSION:6\n") {
		t.Errorf("expected #EXT-X-VERSION:6, got:\n%s", out)
	}
}

// ---- Variant fields ----

func TestBuild_VariantAllFields(t *testing.T) {
	out, err := NewMasterBuilder().
		AddAudioTrack(AudioTrackParams{GroupID: "audio", Name: "English", Language: "en"}).
		AddSubtitleTrack(SubtitleTrackParams{GroupID: "subs", Name: "English", Language: "en", URI: "subs.m3u8"}).
		AddVariant(VariantParams{
			URI:              "hd.m3u8",
			Bandwidth:        8000000,
			AverageBandwidth: 7500000,
			Codecs:           "avc1.640028,mp4a.40.2",
			Width:            1920,
			Height:           1080,
			FrameRate:        29.97,
			AudioGroupID:     "audio",
			SubtitleGroupID:  "subs",
			HDCPLevel:        "TYPE-0",
		}).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"BANDWIDTH=8000000",
		"AVERAGE-BANDWIDTH=7500000",
		`CODECS="avc1.640028,mp4a.40.2"`,
		"RESOLUTION=1920x1080",
		"FRAME-RATE=29.970",
		`AUDIO="audio"`,
		`SUBTITLES="subs"`,
		"HDCP-LEVEL=TYPE-0",
		"hd.m3u8",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

// ---- Audio track ----

func TestBuild_AudioTrack(t *testing.T) {
	out, err := NewMasterBuilder().
		AddAudioTrack(AudioTrackParams{
			GroupID:    "audio",
			Name:       "English",
			Language:   "en",
			URI:        "audio-en.m3u8",
			Default:    true,
			AutoSelect: true,
		}).
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 1000000, AudioGroupID: "audio"}).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#EXT-X-MEDIA:") {
		t.Error("expected #EXT-X-MEDIA line")
	}
	if !strings.Contains(out, "TYPE=AUDIO") {
		t.Error("expected TYPE=AUDIO")
	}
	if !strings.Contains(out, "DEFAULT=YES") {
		t.Error("expected DEFAULT=YES")
	}
	if !strings.Contains(out, "AUTOSELECT=YES") {
		t.Error("expected AUTOSELECT=YES")
	}
}

// ---- Subtitle track ----

func TestBuild_SubtitleTrack_Forced(t *testing.T) {
	out, err := NewMasterBuilder().
		AddSubtitleTrack(SubtitleTrackParams{
			GroupID:  "subs",
			Name:     "English",
			Language: "en",
			URI:      "subs.m3u8",
			Forced:   true,
		}).
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 1000000, SubtitleGroupID: "subs"}).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "TYPE=SUBTITLES") {
		t.Error("expected TYPE=SUBTITLES")
	}
	if !strings.Contains(out, "FORCED=YES") {
		t.Error("expected FORCED=YES")
	}
}

// ---- I-frame stream ----

func TestBuild_IFrameStream(t *testing.T) {
	out, err := NewMasterBuilder().
		AddVariant(VariantParams{URI: "v.m3u8", Bandwidth: 5000000}).
		AddIFrameStream(IFrameParams{
			URI:       "iframe.m3u8",
			Bandwidth: 200000,
			Codecs:    "avc1.640028",
			Width:     1280,
			Height:    720,
		}).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#EXT-X-I-FRAME-STREAM-INF:") {
		t.Error("expected #EXT-X-I-FRAME-STREAM-INF line")
	}
	if !strings.Contains(out, "iframe.m3u8") {
		t.Error("expected iframe URI")
	}
}

// ---- Variant insertion order (BH-09) ----

func TestBuild_VariantInsertionOrder(t *testing.T) {
	uris := []string{"low.m3u8", "mid.m3u8", "high.m3u8"}
	b := NewMasterBuilder()
	for i, u := range uris {
		b.AddVariant(VariantParams{URI: u, Bandwidth: (i + 1) * 1000000})
	}
	out, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, err := Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	for i, want := range uris {
		if p.Variants[i].URI != want {
			t.Errorf("Variants[%d].URI = %q, want %q", i, p.Variants[i].URI, want)
		}
	}
}

// ---- Output is re-parseable ----

func TestBuild_OutputReparseable(t *testing.T) {
	out, err := NewMasterBuilder().
		AddAudioTrack(AudioTrackParams{GroupID: "audio", Name: "English", Language: "en", URI: "audio.m3u8", Default: true, AutoSelect: true}).
		AddVariant(VariantParams{URI: "hd.m3u8", Bandwidth: 5000000, Width: 1920, Height: 1080, AudioGroupID: "audio"}).
		AddVariant(VariantParams{URI: "sd.m3u8", Bandwidth: 1500000, Width: 1280, Height: 720, AudioGroupID: "audio"}).
		AddIFrameStream(IFrameParams{URI: "iframe-hd.m3u8", Bandwidth: 200000}).
		Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	p, err := Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(p.Variants) != 2 {
		t.Errorf("variants = %d, want 2", len(p.Variants))
	}
	if len(p.AudioTracks) != 1 {
		t.Errorf("audio tracks = %d, want 1", len(p.AudioTracks))
	}
	if len(p.IFrames) != 1 {
		t.Errorf("I-frames = %d, want 1", len(p.IFrames))
	}
}

// ---- Multiple variants with same audio group (valid) ----

func TestBuild_MultipleVariantsSameAudioGroup(t *testing.T) {
	_, err := NewMasterBuilder().
		AddAudioTrack(AudioTrackParams{GroupID: "audio", Name: "English", Language: "en"}).
		AddVariant(VariantParams{URI: "hd.m3u8", Bandwidth: 5000000, AudioGroupID: "audio"}).
		AddVariant(VariantParams{URI: "sd.m3u8", Bandwidth: 1500000, AudioGroupID: "audio"}).
		Build()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
