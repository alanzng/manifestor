package hls

import (
	"os"
	"testing"
)

func BenchmarkSliceMediaPlaylist(b *testing.B) {
	raw, err := os.ReadFile("../testdata/hls/media_short.m3u8")
	if err != nil {
		b.Fatal(err)
	}
	s := string(raw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SliceMediaPlaylist(s, 2)
	}
}

func BenchmarkSliceMediaPlaylistByDuration(b *testing.B) {
	raw, err := os.ReadFile("../testdata/hls/media_short.m3u8")
	if err != nil {
		b.Fatal(err)
	}
	s := string(raw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SliceMediaPlaylistByDuration(s, 8, 4)
	}
}

func BenchmarkAddTrialSuffix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AddTrialSuffix("video_720p.m3u8")
	}
}
