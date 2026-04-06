package hls

// MasterBuilder builds an HLS Master Playlist from scratch.
// Use NewMasterBuilder to create one.
type MasterBuilder struct {
	version   int
	variants  []VariantParams
	audio     []AudioTrackParams
	subtitles []SubtitleTrackParams
	iframes   []IFrameParams
}

// NewMasterBuilder returns a new MasterBuilder with defaults applied.
func NewMasterBuilder() *MasterBuilder {
	return &MasterBuilder{version: 3}
}

// SetVersion sets the #EXT-X-VERSION value. Defaults to 3.
func (b *MasterBuilder) SetVersion(n int) *MasterBuilder {
	b.version = n
	return b
}

// AddVariant appends a video variant stream to the playlist.
func (b *MasterBuilder) AddVariant(p VariantParams) *MasterBuilder {
	b.variants = append(b.variants, p)
	return b
}

// AddAudioTrack appends an #EXT-X-MEDIA TYPE=AUDIO entry.
func (b *MasterBuilder) AddAudioTrack(p AudioTrackParams) *MasterBuilder {
	b.audio = append(b.audio, p)
	return b
}

// AddSubtitleTrack appends an #EXT-X-MEDIA TYPE=SUBTITLES entry.
func (b *MasterBuilder) AddSubtitleTrack(p SubtitleTrackParams) *MasterBuilder {
	b.subtitles = append(b.subtitles, p)
	return b
}

// AddIFrameStream appends an #EXT-X-I-FRAME-STREAM-INF entry.
func (b *MasterBuilder) AddIFrameStream(p IFrameParams) *MasterBuilder {
	b.iframes = append(b.iframes, p)
	return b
}

// Build validates the configuration and serializes the playlist to a valid
// HLS m3u8 string. Variants are emitted in the order they were added.
//
// Returns ErrEmptyVariantList if no variants were added.
// Returns ErrInvalidVariant if any variant is missing URI or Bandwidth.
// Returns ErrOrphanedGroupID if a variant's AudioGroupID or SubtitleGroupID
// has no matching EXT-X-MEDIA entry.
func (b *MasterBuilder) Build() (string, error) {
	if len(b.variants) == 0 {
		return "", ErrEmptyVariantList
	}

	// Validate variants.
	for _, v := range b.variants {
		if v.URI == "" || v.Bandwidth == 0 {
			return "", ErrInvalidVariant
		}
	}

	// Build group-ID sets for orphan detection.
	audioGroups := make(map[string]bool, len(b.audio))
	for _, a := range b.audio {
		audioGroups[a.GroupID] = true
	}
	subtitleGroups := make(map[string]bool, len(b.subtitles))
	for _, s := range b.subtitles {
		subtitleGroups[s.GroupID] = true
	}

	for _, v := range b.variants {
		if v.AudioGroupID != "" && !audioGroups[v.AudioGroupID] {
			return "", ErrOrphanedGroupID
		}
		if v.SubtitleGroupID != "" && !subtitleGroups[v.SubtitleGroupID] {
			return "", ErrOrphanedGroupID
		}
	}

	// Convert params to types and delegate to Serialize.
	p := &MasterPlaylist{Version: b.version}

	for _, a := range b.audio {
		p.AudioTracks = append(p.AudioTracks, MediaTrack{
			Type:       "AUDIO",
			GroupID:    a.GroupID,
			Name:       a.Name,
			Language:   a.Language,
			URI:        a.URI,
			Default:    a.Default,
			AutoSelect: a.AutoSelect,
			Forced:     a.Forced,
		})
	}

	for _, s := range b.subtitles {
		p.Subtitles = append(p.Subtitles, MediaTrack{
			Type:     "SUBTITLES",
			GroupID:  s.GroupID,
			Name:     s.Name,
			Language: s.Language,
			URI:      s.URI,
			Default:  s.Default,
			Forced:   s.Forced,
		})
	}

	for _, v := range b.variants {
		p.Variants = append(p.Variants, Variant(v))
	}

	for _, f := range b.iframes {
		p.IFrames = append(p.IFrames, IFrameStream(f))
	}

	return Serialize(p)
}
