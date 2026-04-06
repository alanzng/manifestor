package dash

import "errors"

// Sentinel errors returned by the dash package.
var (
	// ErrNoVariantsRemain is returned when all representations have been filtered out.
	ErrNoVariantsRemain = errors.New("dash: no representations remain after filtering")

	// ErrParseFailure is returned when the MPD content cannot be parsed.
	ErrParseFailure = errors.New("dash: failed to parse MPD")

	// ErrEmptyVariantList is returned when Build() is called with no representations added.
	ErrEmptyVariantList = errors.New("dash: no representations added before Build()")

	// ErrInvalidVariant is returned when a representation is missing a required field.
	ErrInvalidVariant = errors.New("dash: representation is missing a required field (ID or Bandwidth)")

	// ErrInvalidLanguageTag is returned when an AdaptationSet lang value is not a valid BCP-47 tag.
	ErrInvalidLanguageTag = errors.New("dash: lang is not a valid BCP-47 language tag")
)
