package hls

import (
	"os"
	"strings"
	"testing"
)

func TestAddTrialSuffix(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"playlist.m3u8", "playlist_trial.m3u8"},
		{"video_720p.m3u8", "video_720p_trial.m3u8"},
		{"noext", "noext_trial"},
		{"", "_trial"},
	}
	for _, tc := range cases {
		if got := AddTrialSuffix(tc.in); got != tc.want {
			t.Errorf("AddTrialSuffix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAddTrialSuffix_QueryStringDocumentedBehavior(t *testing.T) {
	// path.Ext treats the query as part of the extension. This test pins the
	// current behavior so any future change is explicit. If a real call site
	// needs query-aware splitting, that's a follow-up — see Out of Scope.
	got := AddTrialSuffix("audio.m3u8?token=abc")
	want := "audio_trial.m3u8?token=abc"
	if got != want {
		t.Errorf("query-string behavior changed: got %q, want %q", got, want)
	}
}

// shared test helpers — used by later tests
func mustReadFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}
func contains(haystack, needle string) bool { return strings.Contains(haystack, needle) }
func count(haystack, needle string) int     { return strings.Count(haystack, needle) }

func TestSliceMediaPlaylist_KeepsFirstNSegments(t *testing.T) {
	raw := mustReadFile(t, "../testdata/hls/media_short.m3u8")
	out := SliceMediaPlaylist(raw, 2)

	for _, want := range []string{
		"#EXTM3U",
		"#EXT-X-VERSION:6",
		"#EXT-X-TARGETDURATION:4",
		`#EXT-X-MAP:URI="init.mp4"`,
	} {
		if !contains(out, want) {
			t.Errorf("missing header tag %q in:\n%s", want, out)
		}
	}
	if !contains(out, "seg-00001.m4s") || !contains(out, "seg-00002.m4s") {
		t.Errorf("expected first two segments kept, got:\n%s", out)
	}
	if contains(out, "seg-00003.m4s") {
		t.Errorf("expected seg-00003.m4s dropped, got:\n%s", out)
	}
	if count(out, "#EXT-X-ENDLIST") != 1 {
		t.Errorf("expected exactly one #EXT-X-ENDLIST, got %d", count(out, "#EXT-X-ENDLIST"))
	}
}

func TestSliceMediaPlaylist_GoldenFile(t *testing.T) {
	raw := mustReadFile(t, "../testdata/hls/media_short.m3u8")
	want := mustReadFile(t, "../testdata/hls/media_short_trial2.golden.m3u8")
	got := SliceMediaPlaylist(raw, 2)
	if got != want {
		t.Errorf("output drift vs golden file.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSliceMediaPlaylist_HandlesCRLF(t *testing.T) {
	raw := strings.ReplaceAll(mustReadFile(t, "../testdata/hls/media_short.m3u8"), "\n", "\r\n")
	out := SliceMediaPlaylist(raw, 1)
	if !contains(out, "seg-00001.m4s") {
		t.Errorf("CRLF input not handled: %s", out)
	}
	if contains(out, "\r") {
		t.Errorf("output should be LF-only, got CR in:\n%q", out)
	}
}

func TestSliceMediaPlaylist_ZeroOrNegativeNYieldsNoSegments(t *testing.T) {
	raw := mustReadFile(t, "../testdata/hls/media_short.m3u8")
	for _, n := range []int{0, -3} {
		out := SliceMediaPlaylist(raw, n)
		if contains(out, "#EXTINF") {
			t.Errorf("expected no segments for n=%d, got:\n%s", n, out)
		}
		if !contains(out, "#EXT-X-ENDLIST") {
			t.Errorf("expected ENDLIST appended even for n=%d", n)
		}
	}
}

func TestSliceMediaPlaylist_EmptyInput(t *testing.T) {
	out := SliceMediaPlaylist("", 5)
	if out != "#EXT-X-ENDLIST\n" {
		t.Errorf("empty input: got %q, want %q", out, "#EXT-X-ENDLIST\n")
	}
}
