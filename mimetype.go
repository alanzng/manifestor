package manifestor

// MimeType identifies a media MIME type used for filtering representations.
type MimeType string

const (
	MimeVideoMP4  MimeType = "video/mp4"
	MimeVideoWebM MimeType = "video/webm"
	MimeAudioMP4  MimeType = "audio/mp4"
	MimeAudioWebM MimeType = "audio/webm"
	MimeTextVTT   MimeType = "text/vtt"
	MimeTextTTML  MimeType = "application/ttml+xml"
)
