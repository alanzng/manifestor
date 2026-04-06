package hls

// Integration tests that reproduce the Vieon VOD production use case for HLS:
//
//   Input:  vieon_vod.m3u8  — AVC1 + HVC1 video, one Tajik audio, relative URIs
//   Output: H.265 only, max 720p, absolute CDN URIs, SUBTITLES= on variants,
//           injected dubbed audio and Vietnamese subtitle track

import (
	"strings"
	"testing"
)

const vieonHLSCDNBase = "https://vod-bp.vieon.vn/56714cc3c2fc1068f083ae040a56621d/1775572699000/ott-vod-202603/vod/2026/03/12/bffa9046-2fe5-4b01-888e-ed9d91ce035e/"

func TestVieon_HLS_Filter_H265Only(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content, WithCodec("h265"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !strings.HasPrefix(v.Codecs, "hvc1.") {
			t.Errorf("non-H265 variant survived: %s codecs=%q", v.URI, v.Codecs)
		}
	}
}

func TestVieon_HLS_Filter_MaxResolution720p(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content, WithCodec("h265"), WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	// hvc1-1 (854x480) and hvc1-3 (1280x720) survive; hvc1-2 (1920x1080) dropped.
	if len(p.Variants) != 2 {
		t.Fatalf("Variants = %d, want 2", len(p.Variants))
	}
	uris := map[string]bool{}
	for _, v := range p.Variants {
		uris[v.URI] = true
	}
	if !uris["video-hvc1-1.m3u8"] {
		t.Error("expected video-hvc1-1.m3u8 to survive")
	}
	if !uris["video-hvc1-3.m3u8"] {
		t.Error("expected video-hvc1-3.m3u8 to survive")
	}
}

func TestVieon_HLS_Filter_AbsoluteVariantURIs(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if !strings.HasPrefix(v.URI, "https://") {
			t.Errorf("variant URI not absolute: %q", v.URI)
		}
		if !strings.Contains(v.URI, "vod-bp.vieon.vn") {
			t.Errorf("variant URI missing CDN host: %q", v.URI)
		}
	}
}

func TestVieon_HLS_Filter_VariantURIContent(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)

	want := vieonHLSCDNBase + "video-hvc1-1.m3u8"
	found := false
	for _, v := range p.Variants {
		if v.URI == want {
			found = true
		}
	}
	if !found {
		t.Errorf("expected variant URI %q not found", want)
	}
}

func TestVieon_HLS_IFrameURIsRewritten(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)

	if len(p.IFrames) == 0 {
		t.Fatal("no I-frame streams in output")
	}
	for _, f := range p.IFrames {
		if !strings.HasPrefix(f.URI, "https://") {
			t.Errorf("I-frame URI not absolute: %q", f.URI)
		}
	}
}

func TestVieon_HLS_AudioTrackURIRewritten(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)

	if len(p.AudioTracks) == 0 {
		t.Fatal("no audio tracks in output")
	}
	for _, a := range p.AudioTracks {
		if a.URI != "" && !strings.HasPrefix(a.URI, "https://") {
			t.Errorf("audio track URI not absolute: %q", a.URI)
		}
	}
}

func TestVieon_HLS_WithVariantSubtitleGroup(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithVariantSubtitleGroup("subs"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)
	for _, v := range p.Variants {
		if v.SubtitleGroupID != "subs" {
			t.Errorf("variant %s SubtitleGroupID = %q, want subs", v.URI, v.SubtitleGroupID)
		}
	}
	if !strings.Contains(out, `SUBTITLES="subs"`) {
		t.Error("output missing SUBTITLES=\"subs\" attribute")
	}
}

func TestVieon_HLS_InjectDubbedAudio(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	const dubbedCDNBase = "https://vod-bp.vieon.vn/dc07150ba0e3ee475dc7e18731b13906/1775572699000/ott-vod-202603/vod/2026/03/24/5f5bbfae-3d8a-4654-8b69-de5dbe22e518/"

	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
		WithInjectAudioTrack(AudioTrackParams{
			GroupID:  "audio/mp4a",
			Name:     "Thuyết Minh",
			Language: "tm",
			URI:      dubbedCDNBase + "audio-tg-mp4a.m3u8",
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)

	var tm *MediaTrack
	for i := range p.AudioTracks {
		if p.AudioTracks[i].Language == "tm" {
			tm = &p.AudioTracks[i]
			break
		}
	}
	if tm == nil {
		t.Fatal("injected dubbed audio (lang=tm) not found")
	}
	if tm.Name != "Thuyết Minh" {
		t.Errorf("Name = %q, want Thuyết Minh", tm.Name)
	}
	if !strings.Contains(tm.URI, "5f5bbfae-3d8a-4654-8b69-de5dbe22e518") {
		t.Errorf("dubbed audio URI = %q", tm.URI)
	}
}

func TestVieon_HLS_InjectSubtitle(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
		WithVariantSubtitleGroup("subs"),
		WithInjectSubtitle(SubtitleTrackParams{
			GroupID:  "subs",
			Name:     "Tiếng Việt",
			Language: "vi",
			URI:      "https://playlist-free.vieon.vn/playlist/subtitle/3ff5715f-4dd8-45ed-aa59-8145c5a6019e/vi.m3u8",
			Default:  true,
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)

	var vi *MediaTrack
	for i := range p.Subtitles {
		if p.Subtitles[i].Language == "vi" {
			vi = &p.Subtitles[i]
			break
		}
	}
	if vi == nil {
		t.Fatal("injected subtitle (lang=vi) not found")
	}
	if vi.Name != "Tiếng Việt" {
		t.Errorf("Name = %q, want Tiếng Việt", vi.Name)
	}
	if !strings.Contains(vi.URI, "vieon.vn") {
		t.Errorf("subtitle URI = %q", vi.URI)
	}
}

func TestVieon_HLS_FullPipeline(t *testing.T) {
	content := mustReadFixture(t, "../testdata/hls/vieon_vod.m3u8")
	const dubbedCDNBase = "https://vod-bp.vieon.vn/dc07150ba0e3ee475dc7e18731b13906/1775572699000/ott-vod-202603/vod/2026/03/24/5f5bbfae-3d8a-4654-8b69-de5dbe22e518/"

	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonHLSCDNBase),
		WithVariantSubtitleGroup("subs"),
		WithInjectSubtitle(SubtitleTrackParams{
			GroupID:  "subs",
			Name:     "Tiếng Việt",
			Language: "vi",
			URI:      "https://playlist-free.vieon.vn/playlist/subtitle/3ff5715f-4dd8-45ed-aa59-8145c5a6019e/vi.m3u8",
			Default:  true,
		}),
		WithInjectAudioTrack(AudioTrackParams{
			GroupID:  "audio/mp4a",
			Name:     "Thuyết Minh",
			Language: "tm",
			URI:      dubbedCDNBase + "audio-tg-mp4a.m3u8",
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, _ := Parse(out)

	// 2 H265 variants (480p, 720p).
	if len(p.Variants) != 2 {
		t.Errorf("Variants = %d, want 2", len(p.Variants))
	}
	// 2 audio tracks: original tg + injected tm.
	if len(p.AudioTracks) != 2 {
		t.Errorf("AudioTracks = %d, want 2", len(p.AudioTracks))
	}
	// 1 subtitle track: vi.
	if len(p.Subtitles) != 1 {
		t.Errorf("Subtitles = %d, want 1", len(p.Subtitles))
	}
	// 2 I-frame streams (one per H265 variant).
	if len(p.IFrames) != 2 {
		t.Errorf("IFrames = %d, want 2", len(p.IFrames))
	}
	// All variant URIs absolute.
	for _, v := range p.Variants {
		if !strings.HasPrefix(v.URI, "https://") {
			t.Errorf("variant URI not absolute: %q", v.URI)
		}
		if v.SubtitleGroupID != "subs" {
			t.Errorf("variant missing SUBTITLES group: %q", v.URI)
		}
	}
	// All I-frame URIs absolute.
	for _, f := range p.IFrames {
		if !strings.HasPrefix(f.URI, "https://") {
			t.Errorf("I-frame URI not absolute: %q", f.URI)
		}
	}
}
