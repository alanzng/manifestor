package manifest

type clearAudioOption struct{ sharedOption }

// WithClearAudioTracks removes every audio track (HLS) or audio AdaptationSet
// (DASH) parsed from the source manifest before any inject options run.
// Use it together with WithHLSInjectAudioTrack or WithDASHInjectAdaptationSet
// to fully replace origin audio with injected audio.
func WithClearAudioTracks() Option { return clearAudioOption{} }
