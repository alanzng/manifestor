package hls

import (
	"strconv"
	"strings"
)

// Parse parses an HLS Master Playlist from a raw string and returns a
// MasterPlaylist. It returns ErrNotMasterPlaylist if the content is a valid
// media playlist and ErrParseFailure if the content cannot be parsed at all.
func Parse(content string) (*MasterPlaylist, error) {
	lines := splitLines(content)
	if len(lines) == 0 {
		return nil, ErrParseFailure
	}

	// First non-empty line must be #EXTM3U.
	firstLine := ""
	for _, l := range lines {
		if l != "" {
			firstLine = l
			break
		}
	}
	if !strings.HasPrefix(firstLine, "#EXTM3U") {
		return nil, ErrParseFailure
	}

	// Detect media playlists before doing a full parse.
	for _, l := range lines {
		if strings.HasPrefix(l, "#EXT-X-TARGETDURATION") {
			return nil, ErrNotMasterPlaylist
		}
	}

	p := &MasterPlaylist{}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		switch {
		case line == "" || line == "#EXTM3U":
			// skip

		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			v, err := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-VERSION:"))
			if err == nil {
				p.Version = v
			}

		case strings.HasPrefix(line, "#EXT-X-MEDIA:"):
			attrs := parseAttrs(strings.TrimPrefix(line, "#EXT-X-MEDIA:"))
			track := parseMediaTrack(attrs)
			switch strings.ToUpper(track.Type) {
			case "AUDIO":
				p.AudioTracks = append(p.AudioTracks, track)
			case "SUBTITLES":
				p.Subtitles = append(p.Subtitles, track)
			default:
				// CLOSED-CAPTIONS and unknown types stored as subtitles-like
				p.Subtitles = append(p.Subtitles, track)
			}

		case strings.HasPrefix(line, "#EXT-X-STREAM-INF:"):
			attrs := parseAttrs(strings.TrimPrefix(line, "#EXT-X-STREAM-INF:"))
			v := parseVariant(attrs)
			// URI is on the next non-blank line.
			i++
			for i < len(lines) && lines[i] == "" {
				i++
			}
			if i < len(lines) {
				v.URI = lines[i]
			}
			p.Variants = append(p.Variants, v)

		case strings.HasPrefix(line, "#EXT-X-I-FRAME-STREAM-INF:"):
			attrs := parseAttrs(strings.TrimPrefix(line, "#EXT-X-I-FRAME-STREAM-INF:"))
			p.IFrames = append(p.IFrames, parseIFrameStream(attrs))

		case strings.HasPrefix(line, "#"):
			// Unknown tag — preserve for pass-through (P-05).
			p.Raw = append(p.Raw, line)
		}
	}

	return p, nil
}

// splitLines splits content on newlines and trims carriage returns (CRLF support).
func splitLines(content string) []string {
	raw := strings.Split(content, "\n")
	out := make([]string, len(raw))
	for i, l := range raw {
		out[i] = strings.TrimRight(l, "\r")
	}
	return out
}

// parseAttrs parses an HLS attribute list string into a key→value map.
//
// Format: KEY=VALUE,KEY="QUOTED VALUE",KEY=123
// Commas inside double-quoted values are not treated as delimiters.
func parseAttrs(s string) map[string]string {
	attrs := make(map[string]string)
	for len(s) > 0 {
		// Find key (everything up to '=').
		eq := strings.IndexByte(s, '=')
		if eq < 0 {
			break
		}
		key := strings.TrimSpace(s[:eq])
		s = s[eq+1:]

		// Read value.
		var value string
		if len(s) > 0 && s[0] == '"' {
			// Quoted value — scan to closing quote.
			end := strings.IndexByte(s[1:], '"')
			if end < 0 {
				// Unterminated quote — take rest of string.
				value = s[1:]
				s = ""
			} else {
				value = s[1 : end+1]
				s = s[end+2:] // skip closing quote
			}
		} else {
			// Unquoted value — read until comma.
			comma := strings.IndexByte(s, ',')
			if comma < 0 {
				value = s
				s = ""
			} else {
				value = s[:comma]
				s = s[comma:]
			}
		}

		attrs[key] = value

		// Skip the separating comma.
		if len(s) > 0 && s[0] == ',' {
			s = s[1:]
		}
	}
	return attrs
}

// parseVariant converts an EXT-X-STREAM-INF attribute map into a Variant.
func parseVariant(attrs map[string]string) Variant {
	v := Variant{}
	if bw, ok := attrs["BANDWIDTH"]; ok {
		v.Bandwidth, _ = strconv.Atoi(bw)
	}
	if abw, ok := attrs["AVERAGE-BANDWIDTH"]; ok {
		v.AverageBandwidth, _ = strconv.Atoi(abw)
	}
	if c, ok := attrs["CODECS"]; ok {
		v.Codecs = c
	}
	if res, ok := attrs["RESOLUTION"]; ok {
		v.Width, v.Height = parseResolution(res)
	}
	if fr, ok := attrs["FRAME-RATE"]; ok {
		v.FrameRate, _ = strconv.ParseFloat(fr, 64)
	}
	if a, ok := attrs["AUDIO"]; ok {
		v.AudioGroupID = a
	}
	if sub, ok := attrs["SUBTITLES"]; ok {
		v.SubtitleGroupID = sub
	}
	if hdcp, ok := attrs["HDCP-LEVEL"]; ok {
		v.HDCPLevel = hdcp
	}
	return v
}

// parseMediaTrack converts an EXT-X-MEDIA attribute map into a MediaTrack.
func parseMediaTrack(attrs map[string]string) MediaTrack {
	t := MediaTrack{}
	if typ, ok := attrs["TYPE"]; ok {
		t.Type = typ
	}
	if gid, ok := attrs["GROUP-ID"]; ok {
		t.GroupID = gid
	}
	if name, ok := attrs["NAME"]; ok {
		t.Name = name
	}
	if lang, ok := attrs["LANGUAGE"]; ok {
		t.Language = lang
	}
	if uri, ok := attrs["URI"]; ok {
		t.URI = uri
	}
	if d, ok := attrs["DEFAULT"]; ok {
		t.Default = strings.EqualFold(d, "YES")
	}
	if a, ok := attrs["AUTOSELECT"]; ok {
		t.AutoSelect = strings.EqualFold(a, "YES")
	}
	if f, ok := attrs["FORCED"]; ok {
		t.Forced = strings.EqualFold(f, "YES")
	}
	return t
}

// parseIFrameStream converts an EXT-X-I-FRAME-STREAM-INF attribute map into an IFrameStream.
func parseIFrameStream(attrs map[string]string) IFrameStream {
	f := IFrameStream{}
	if uri, ok := attrs["URI"]; ok {
		f.URI = uri
	}
	if bw, ok := attrs["BANDWIDTH"]; ok {
		f.Bandwidth, _ = strconv.Atoi(bw)
	}
	if c, ok := attrs["CODECS"]; ok {
		f.Codecs = c
	}
	if res, ok := attrs["RESOLUTION"]; ok {
		f.Width, f.Height = parseResolution(res)
	}
	return f
}

// parseResolution parses a "WxH" resolution string into width and height.
func parseResolution(s string) (w, h int) {
	parts := strings.SplitN(s, "x", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	w, _ = strconv.Atoi(parts[0])
	h, _ = strconv.Atoi(parts[1])
	return w, h
}
