package manifest

// audioLabelByLangOption carries a language-keyed map of display names to
// apply to origin audio tracks in HLS/DASH.
type audioLabelByLangOption struct {
	sharedOption
	labels map[string]string
}

// WithAudioLabelByLanguage rewrites the display name of every ORIGIN audio
// track whose BCP-47 language tag (case-insensitive) appears in labels.
// HLS: sets the NAME attribute on #EXT-X-MEDIA:TYPE=AUDIO entries.
// DASH: sets the label on audio AdaptationSets.
//
// Subtitles, video sets, and injected audio tracks are not affected.
// Empty-string values in labels are ignored (original name preserved).
//
// Example:
//
//	manifest.Filter(content, manifest.WithAudioLabelByLanguage(map[string]string{
//	    "tg": "Tiếng gốc",
//	    "en": "English",
//	}))
func WithAudioLabelByLanguage(labels map[string]string) Option {
	// Copy defensively so callers mutating their map post-call don't
	// influence filter behavior.
	cp := make(map[string]string, len(labels))
	for k, v := range labels {
		cp[k] = v
	}
	return audioLabelByLangOption{labels: cp}
}
