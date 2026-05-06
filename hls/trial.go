package hls

import (
	"path"
	"regexp"
	"strconv"
	"strings"
)

// AddTrialSuffix transforms "foo.m3u8" into "foo_trial.m3u8". It splits on
// the final extension as reported by path.Ext; if there is no extension the
// suffix is appended directly. URIs containing a query string have the suffix
// inserted before whatever path.Ext considers the extension — callers passing
// signed URLs should strip the query first if they need the suffix to land
// before the path's true extension.
func AddTrialSuffix(uri string) string {
	ext := path.Ext(uri)
	return strings.TrimSuffix(uri, ext) + "_trial" + ext
}

// SliceMediaPlaylist trims a raw HLS media playlist to the first n #EXTINF
// segments and appends a single #EXT-X-ENDLIST line. Header tags (including
// #EXT-X-MAP) are preserved verbatim; any pre-existing #EXT-X-ENDLIST in the
// input is dropped before the new one is appended.
//
// Intended for VOD playlists. Live and event playlists are out of scope —
// re-emitting #EXT-X-ENDLIST on a live origin produces a non-conformant
// trial manifest.
//
// Limitation (matches the legacy vo-playlist behavior): standalone tags that
// appear *between* segments without an attached #EXTINF block (e.g. an
// #EXT-X-KEY rotation, or an unpaired #EXT-X-DATERANGE) are dropped. This is
// safe for unencrypted on-demand HLS origins. Revisit if EXT-X-KEY rotation
// is ever introduced.
//
// Input is expected to be UTF-8 without BOM. CRLF line endings are tolerated;
// output is always LF.
//
// If n <= 0 the result has zero segments but still includes the header and
// the trailing #EXT-X-ENDLIST.
//
// Pure function — safe for concurrent use.
func SliceMediaPlaylist(raw string, n int) string {
	if n < 0 {
		n = 0
	}
	lines := strings.Split(raw, "\n")
	var sb strings.Builder
	inHeader := true
	taken := 0
	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.HasPrefix(line, "#EXTINF:") {
			inHeader = false
			if taken >= n {
				break
			}
			sb.WriteString(line)
			sb.WriteByte('\n')
			advance, took := appendSegmentBody(&sb, lines, i+1)
			i = advance
			if took {
				taken++
			}
			continue
		}
		if inHeader {
			appendHeaderLine(&sb, line)
		}
	}
	sb.WriteString("#EXT-X-ENDLIST\n")
	return sb.String()
}

// appendSegmentBody walks lines starting at start, copying intermediate tags
// and the URI line for one segment to sb. Returns the index of the URI line
// (so the caller's outer loop resumes after it) and whether a URI was emitted.
func appendSegmentBody(sb *strings.Builder, lines []string, start int) (int, bool) {
	for j := start; j < len(lines); j++ {
		next := strings.TrimRight(lines[j], "\r")
		if next == "" {
			continue
		}
		if strings.HasPrefix(next, "#EXTINF:") {
			return j - 1, false
		}
		sb.WriteString(next)
		sb.WriteByte('\n')
		if !strings.HasPrefix(next, "#") {
			return j, true
		}
	}
	return len(lines) - 1, false
}

// appendHeaderLine writes one header (pre-segment) line, dropping blanks and
// any pre-existing #EXT-X-ENDLIST (re-emitted at the end by SliceMediaPlaylist).
func appendHeaderLine(sb *strings.Builder, line string) {
	if line == "" || strings.HasPrefix(line, "#EXT-X-ENDLIST") {
		return
	}
	sb.WriteString(line)
	sb.WriteByte('\n')
}

var trialTargetDurationRe = regexp.MustCompile(`(?m)^#EXT-X-TARGETDURATION:(\d+)`)

// SliceMediaPlaylistByDuration trims raw to the first ⌊trialDuration / target⌋
// segments, where target is read from the playlist's #EXT-X-TARGETDURATION or
// from defaultTarget if that tag is absent. If both are zero or negative,
// target falls back to 4 (matching legacy vo-playlist behavior).
//
// Returns the sliced playlist and the segment count actually requested.
//
// Pure function — safe for concurrent use.
func SliceMediaPlaylistByDuration(raw string, trialDuration, defaultTarget int) (string, int) {
	target := defaultTarget
	if v := extractTargetDuration(raw); v > 0 {
		target = v
	}
	if target <= 0 {
		target = 4
	}
	n := trialDuration / target
	return SliceMediaPlaylist(raw, n), n
}

func extractTargetDuration(raw string) int {
	m := trialTargetDurationRe.FindStringSubmatch(raw)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.Atoi(m[1])
	return v
}
