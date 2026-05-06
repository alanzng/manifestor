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
