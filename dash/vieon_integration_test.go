package dash

// Integration tests that reproduce the Vieon VOD production use case:
//
//   Input:  vieon_vod.mpd  — AVC1 + HVC1 video, one Tajik audio track, relative BaseURLs
//   Output: H.265 only, max 720p, absolute CDN URLs, injected dubbed audio and subtitle

import (
	"strings"
	"testing"
)

const vieonCDNBase = "https://vod-bp.vieon.vn/fb3ae865ebf27eec47466c132d33f30d/1775555787000/ott-vod-202603/vod/2026/03/12/bffa9046-2fe5-4b01-888e-ed9d91ce035e/"

func TestVieon_Filter_H265Only(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")
	out, err := Filter(content, WithCodec("h265"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	for _, p := range m.Periods {
		for _, as := range p.AdaptationSets {
			for _, r := range as.Representations {
				if r.Codecs != "" && !strings.HasPrefix(r.Codecs, "hvc1.") && !strings.HasPrefix(r.Codecs, "mp4a.") {
					t.Errorf("non-H265 codec survived: %q (rep %s)", r.Codecs, r.ID)
				}
			}
		}
	}
}

func TestVieon_Filter_MaxResolution720p(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")
	out, err := Filter(content, WithCodec("h265"), WithMaxResolution(1280, 720))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	var videoReps []Representation
	for _, p := range m.Periods {
		for _, as := range p.AdaptationSets {
			if strings.HasPrefix(as.MimeType, "video/") {
				videoReps = append(videoReps, as.Representations...)
			}
		}
	}
	// hvc1-1 (854x480) and hvc1-3 (1280x720) survive; hvc1-2 (1920x1080) is dropped.
	if len(videoReps) != 2 {
		t.Fatalf("video representations = %d, want 2", len(videoReps))
	}
	ids := map[string]bool{}
	for _, r := range videoReps {
		ids[r.ID] = true
	}
	if !ids["video-hvc1-1"] {
		t.Error("expected video-hvc1-1 to survive")
	}
	if !ids["video-hvc1-3"] {
		t.Error("expected video-hvc1-3 to survive")
	}
	if ids["video-hvc1-2"] {
		t.Error("video-hvc1-2 (1920x1080) should have been dropped")
	}
}

func TestVieon_Filter_AbsoluteBaseURLs(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	for _, p := range m.Periods {
		for _, as := range p.AdaptationSets {
			for _, r := range as.Representations {
				if r.BaseURL == "" {
					continue
				}
				if !strings.HasPrefix(r.BaseURL, "https://") {
					t.Errorf("rep %s BaseURL not absolute: %q", r.ID, r.BaseURL)
				}
				if !strings.Contains(r.BaseURL, "vod-bp.vieon.vn") {
					t.Errorf("rep %s BaseURL missing CDN host: %q", r.ID, r.BaseURL)
				}
			}
		}
	}
}

func TestVieon_Filter_VideoBaseURLContent(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	// Verify hvc1-1 BaseURL is correctly resolved.
	as := m.Periods[0].AdaptationSets[0]
	var hvc1 *Representation
	for i := range as.Representations {
		if as.Representations[i].ID == "video-hvc1-1" {
			hvc1 = &as.Representations[i]
			break
		}
	}
	if hvc1 == nil {
		t.Fatal("video-hvc1-1 not found")
	}
	want := vieonCDNBase + "media-video-hvc1-1.mp4"
	if hvc1.BaseURL != want {
		t.Errorf("BaseURL = %q\nwant    %q", hvc1.BaseURL, want)
	}
}

func TestVieon_InjectDubbedAudio(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")

	const dubbedCDNBase = "https://vod-bp.vieon.vn/3ad1dcfca2aeaf6a13118a3fed017be9/1775555787000/ott-vod-202603/vod/2026/03/24/5f5bbfae-3d8a-4654-8b69-de5dbe22e518/"

	dubbedAudio := AdaptationSetParams{
		MimeType: "audio/mp4",
		Lang:     "tm",
		Name:     "Thuyết Minh",
		Representations: []RepresentationParams{
			{
				ID:        "5f5bbfae-3d8a-4654-8b69-de5dbe22e518_tm_196728",
				Bandwidth: 196728,
				Codecs:    "mp4a.40.2",
				BaseURL:   dubbedCDNBase + "media-audio-tg-mp4a.mp4",
				AudioChannelConfiguration: &AudioChannelConfiguration{
					SchemeIDURI: "urn:mpeg:dash:23003:3:audio_channel_configuration:2011",
					Value:       "2",
				},
			},
		},
	}

	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonCDNBase),
		WithInjectAdaptationSet(dubbedAudio),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	// Find injected dubbed audio.
	var tm *AdaptationSet
	for i := range m.Periods[0].AdaptationSets {
		if m.Periods[0].AdaptationSets[i].Lang == "tm" {
			tm = &m.Periods[0].AdaptationSets[i]
			break
		}
	}
	if tm == nil {
		t.Fatal("injected dubbed audio (lang=tm) not found")
	}
	if tm.Name != "Thuyết Minh" {
		t.Errorf("Name = %q, want Thuyết Minh", tm.Name)
	}
	if len(tm.Representations) != 1 {
		t.Fatalf("Representations = %d, want 1", len(tm.Representations))
	}
	r := tm.Representations[0]
	if r.AudioChannelConfiguration == nil {
		t.Fatal("AudioChannelConfiguration is nil on injected audio rep")
	}
	if r.AudioChannelConfiguration.Value != "2" {
		t.Errorf("AudioChannelConfiguration.Value = %q, want 2", r.AudioChannelConfiguration.Value)
	}
	if !strings.Contains(r.BaseURL, "5f5bbfae-3d8a-4654-8b69-de5dbe22e518") {
		t.Errorf("dubbed audio BaseURL = %q", r.BaseURL)
	}
}

func TestVieon_InjectSubtitle(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")

	subtitle := AdaptationSetParams{
		ContentType: "text",
		MimeType:    "text/vtt",
		Lang:        "vi",
		Roles:       []Role{{SchemeIDURI: "urn:mpeg:dash:role:2011", Value: "subtitle"}},
		Representations: []RepresentationParams{
			{
				ID:        "subtitles/vi",
				Bandwidth: 16,
				BaseURL:   "https://static2.vieon.vn/vieplay-image/subtitle/2026/03/18/rkbpx4p1_climax_raw_master_2026_s01_ep01a_v2.vtt",
			},
		},
	}

	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonCDNBase),
		WithInjectAdaptationSet(subtitle),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	// Subtitle AdaptationSet must be present.
	var sub *AdaptationSet
	for i := range m.Periods[0].AdaptationSets {
		if m.Periods[0].AdaptationSets[i].Lang == "vi" {
			sub = &m.Periods[0].AdaptationSets[i]
			break
		}
	}
	if sub == nil {
		t.Fatal("injected subtitle (lang=vi) not found")
	}
	if sub.MimeType != "text/vtt" {
		t.Errorf("subtitle MimeType = %q, want text/vtt", sub.MimeType)
	}
	if len(sub.Roles) != 1 || sub.Roles[0].Value != "subtitle" {
		t.Errorf("subtitle Role = %+v", sub.Roles)
	}
	if len(sub.Representations) != 1 {
		t.Fatalf("subtitle Representations = %d, want 1", len(sub.Representations))
	}
	if !strings.Contains(sub.Representations[0].BaseURL, "vieon.vn") {
		t.Errorf("subtitle BaseURL = %q", sub.Representations[0].BaseURL)
	}
}

func TestVieon_FullPipeline_AdaptationSetCount(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")

	const dubbedCDNBase = "https://vod-bp.vieon.vn/3ad1dcfca2aeaf6a13118a3fed017be9/1775555787000/ott-vod-202603/vod/2026/03/24/5f5bbfae-3d8a-4654-8b69-de5dbe22e518/"

	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonCDNBase),
		WithInjectAdaptationSet(AdaptationSetParams{
			MimeType: "audio/mp4",
			Lang:     "tm",
			Name:     "Thuyết Minh",
			Representations: []RepresentationParams{
				{ID: "tm-audio", Bandwidth: 196728, Codecs: "mp4a.40.2",
					BaseURL: dubbedCDNBase + "media-audio-tg-mp4a.mp4"},
			},
		}),
		WithInjectAdaptationSet(AdaptationSetParams{
			ContentType: "text",
			MimeType:    "text/vtt",
			Lang:        "vi",
			Roles:       []Role{{SchemeIDURI: "urn:mpeg:dash:role:2011", Value: "subtitle"}},
			Representations: []RepresentationParams{
				{ID: "subtitles/vi", Bandwidth: 16,
					BaseURL: "https://static2.vieon.vn/vieplay-image/subtitle/2026/03/18/rkbpx4p1_climax_raw_master_2026_s01_ep01a_v2.vtt"},
			},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	// Expected: 1 H265 video AdaptationSet + 1 original audio (tg) + 1 dubbed (tm) + 1 subtitle (vi) = 4
	got := len(m.Periods[0].AdaptationSets)
	if got != 4 {
		t.Errorf("AdaptationSets = %d, want 4", got)
		for i, as := range m.Periods[0].AdaptationSets {
			t.Logf("  [%d] mimeType=%s lang=%s reps=%d", i, as.MimeType, as.Lang, len(as.Representations))
		}
	}
}

func TestVieon_OriginalAudioPreserved(t *testing.T) {
	content := mustReadFixture(t, "../testdata/dash/vieon_vod.mpd")
	out, err := Filter(content,
		WithCodec("h265"),
		WithMaxResolution(1280, 720),
		WithAbsoluteURIs(vieonCDNBase),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := Parse(out)

	var tg *AdaptationSet
	for i := range m.Periods[0].AdaptationSets {
		if m.Periods[0].AdaptationSets[i].Lang == "tg" {
			tg = &m.Periods[0].AdaptationSets[i]
			break
		}
	}
	if tg == nil {
		t.Fatal("original audio (lang=tg) not found after filter")
	}
	if len(tg.Representations) != 1 {
		t.Fatalf("original audio Representations = %d, want 1", len(tg.Representations))
	}
	r := tg.Representations[0]
	if r.ID != "audio-tg-mp4a" {
		t.Errorf("ID = %q, want audio-tg-mp4a", r.ID)
	}
	// BaseURL must now be absolute.
	if !strings.HasPrefix(r.BaseURL, "https://") {
		t.Errorf("audio BaseURL not absolute: %q", r.BaseURL)
	}
	if r.AudioChannelConfiguration == nil {
		t.Fatal("AudioChannelConfiguration is nil on original audio")
	}
}
