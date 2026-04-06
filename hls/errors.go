package hls

import "errors"

// Sentinel errors returned by the HLS package.
var (
	// ErrNotMasterPlaylist is returned when the content is a valid HLS media
	// playlist but a master playlist was expected.
	ErrNotMasterPlaylist = errors.New("hls: content is a media playlist, not a master playlist")

	// ErrNoVariantsRemain is returned when all variants have been filtered out.
	ErrNoVariantsRemain = errors.New("hls: no variants remain after filtering")

	// ErrParseFailure is returned when the manifest content cannot be parsed.
	ErrParseFailure = errors.New("hls: failed to parse manifest")

	// ErrEmptyVariantList is returned when Build() is called with no variants added.
	ErrEmptyVariantList = errors.New("hls: no variants added before Build()")

	// ErrInvalidVariant is returned when a variant is missing a required field.
	ErrInvalidVariant = errors.New("hls: variant is missing a required field (URI or Bandwidth)")

	// ErrOrphanedGroupID is returned when a variant references an AudioGroupID or
	// SubtitleGroupID that has no matching #EXT-X-MEDIA entry.
	ErrOrphanedGroupID = errors.New("hls: variant references a group ID with no matching EXT-X-MEDIA entry")
)
