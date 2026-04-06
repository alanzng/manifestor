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
// Returns ErrOrphanedGroupID if a variant's AudioGroupID has no matching EXT-X-MEDIA group.
func (b *MasterBuilder) Build() (string, error) {
	// TODO: implement
	panic("not implemented")
}
