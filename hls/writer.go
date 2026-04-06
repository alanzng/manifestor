package hls

import (
	"fmt"
	"strings"
)

// Serialize serializes a MasterPlaylist back to a valid HLS m3u8 string.
// The output preserves version, media tracks, variants, I-frame streams, and
// any unrecognized lines stored in Raw (requirement P-05).
func Serialize(p *MasterPlaylist) (string, error) {
	var sb strings.Builder

	sb.WriteString("#EXTM3U\n")

	if p.Version > 0 {
		fmt.Fprintf(&sb, "#EXT-X-VERSION:%d\n", p.Version)
	}

	// Media tracks: AUDIO first, then SUBTITLES / CLOSED-CAPTIONS.
	for _, t := range p.AudioTracks {
		writeMediaTrack(&sb, t)
	}
	for _, t := range p.Subtitles {
		writeMediaTrack(&sb, t)
	}

	// Variants.
	for _, v := range p.Variants {
		writeVariant(&sb, v)
	}

	// I-frame streams.
	for _, f := range p.IFrames {
		writeIFrameStream(&sb, f)
	}

	// Pass-through unknown lines.
	for _, raw := range p.Raw {
		sb.WriteString(raw)
		sb.WriteByte('\n')
	}

	return sb.String(), nil
}

// writeMediaTrack emits a single #EXT-X-MEDIA line.
func writeMediaTrack(sb *strings.Builder, t MediaTrack) {
	sb.WriteString("#EXT-X-MEDIA:")

	typ := t.Type
	if typ == "" {
		typ = "AUDIO"
	}
	fmt.Fprintf(sb, "TYPE=%s", typ)
	fmt.Fprintf(sb, ",GROUP-ID=%q", t.GroupID)
	fmt.Fprintf(sb, ",NAME=%q", t.Name)

	if t.Language != "" {
		fmt.Fprintf(sb, ",LANGUAGE=%q", t.Language)
	}
	sb.WriteString(",DEFAULT=")
	if t.Default {
		sb.WriteString("YES")
	} else {
		sb.WriteString("NO")
	}
	sb.WriteString(",AUTOSELECT=")
	if t.AutoSelect {
		sb.WriteString("YES")
	} else {
		sb.WriteString("NO")
	}
	if t.Forced {
		sb.WriteString(",FORCED=YES")
	}
	if t.URI != "" {
		fmt.Fprintf(sb, ",URI=%q", t.URI)
	}

	sb.WriteByte('\n')
}

// writeVariant emits a #EXT-X-STREAM-INF line followed by the variant URI.
func writeVariant(sb *strings.Builder, v Variant) {
	fmt.Fprintf(sb, "#EXT-X-STREAM-INF:BANDWIDTH=%d", v.Bandwidth)

	if v.AverageBandwidth > 0 {
		fmt.Fprintf(sb, ",AVERAGE-BANDWIDTH=%d", v.AverageBandwidth)
	}
	if v.Codecs != "" {
		fmt.Fprintf(sb, ",CODECS=%q", v.Codecs)
	}
	if v.Width > 0 && v.Height > 0 {
		fmt.Fprintf(sb, ",RESOLUTION=%dx%d", v.Width, v.Height)
	}
	if v.FrameRate > 0 {
		fmt.Fprintf(sb, ",FRAME-RATE=%.3f", v.FrameRate)
	}
	if v.AudioGroupID != "" {
		fmt.Fprintf(sb, ",AUDIO=%q", v.AudioGroupID)
	}
	if v.SubtitleGroupID != "" {
		fmt.Fprintf(sb, ",SUBTITLES=%q", v.SubtitleGroupID)
	}
	if v.HDCPLevel != "" {
		fmt.Fprintf(sb, ",HDCP-LEVEL=%s", v.HDCPLevel)
	}

	sb.WriteByte('\n')
	sb.WriteString(v.URI)
	sb.WriteByte('\n')
}

// writeIFrameStream emits a single #EXT-X-I-FRAME-STREAM-INF line.
func writeIFrameStream(sb *strings.Builder, f IFrameStream) {
	fmt.Fprintf(sb, "#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=%d", f.Bandwidth)

	if f.Codecs != "" {
		fmt.Fprintf(sb, ",CODECS=%q", f.Codecs)
	}
	if f.Width > 0 && f.Height > 0 {
		fmt.Fprintf(sb, ",RESOLUTION=%dx%d", f.Width, f.Height)
	}
	fmt.Fprintf(sb, ",URI=%q", f.URI)

	sb.WriteByte('\n')
}

// quoteAttr wraps a string in double quotes, matching HLS attribute list syntax.
// Used internally; exported for use by the builder.
func quoteAttr(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
