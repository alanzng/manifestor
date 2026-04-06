// Package hls provides parsing, filtering, building, and serialization of
// HLS (HTTP Live Streaming) Master Playlists.
package hls

// MasterPlaylist represents a parsed HLS Master Playlist (#EXTM3U).
type MasterPlaylist struct {
	Version     int
	Variants    []Variant
	AudioTracks []MediaTrack
	Subtitles   []MediaTrack
	IFrames     []IFrameStream
	// Raw holds unrecognized lines for pass-through serialization.
	Raw []string
}

// Variant represents a single #EXT-X-STREAM-INF entry and its URI.
type Variant struct {
	URI              string
	Bandwidth        int
	AverageBandwidth int
	Codecs           string
	Width            int
	Height           int
	FrameRate        float64
	AudioGroupID     string
	SubtitleGroupID  string
	HDCPLevel        string
}

// MediaTrack represents a #EXT-X-MEDIA entry (audio, subtitles, closed-captions).
type MediaTrack struct {
	Type       string // AUDIO | SUBTITLES | CLOSED-CAPTIONS
	GroupID    string
	Name       string
	Language   string
	URI        string
	Default    bool
	AutoSelect bool
	Forced     bool
}

// IFrameStream represents a #EXT-X-I-FRAME-STREAM-INF entry.
type IFrameStream struct {
	URI       string
	Bandwidth int
	Codecs    string
	Width     int
	Height    int
}

// VariantParams holds the parameters for building a video variant.
type VariantParams struct {
	URI              string // required
	Bandwidth        int    // required
	AverageBandwidth int
	Codecs           string
	Width            int
	Height           int
	FrameRate        float64
	AudioGroupID     string
	SubtitleGroupID  string
	HDCPLevel        string
}

// AudioTrackParams holds the parameters for building an #EXT-X-MEDIA AUDIO entry.
type AudioTrackParams struct {
	GroupID    string // required
	Name       string // required
	Language   string
	URI        string
	Default    bool
	AutoSelect bool
	Forced     bool
}

// SubtitleTrackParams holds the parameters for building an #EXT-X-MEDIA SUBTITLES entry.
type SubtitleTrackParams struct {
	GroupID  string // required
	Name     string // required
	Language string
	URI      string // required
	Default  bool
	Forced   bool
}

// IFrameParams holds the parameters for building an #EXT-X-I-FRAME-STREAM-INF entry.
type IFrameParams struct {
	URI       string // required
	Bandwidth int    // required
	Codecs    string
	Width     int
	Height    int
}
