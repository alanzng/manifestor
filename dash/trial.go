package dash

import (
	"fmt"
	"regexp"
)

// trialMPDDurationRe matches the mediaPresentationDuration attribute on the
// root MPD element. Matches single OR double-quoted values; uses negated
// character classes so each match stops at the first closing quote.
var trialMPDDurationRe = regexp.MustCompile(`mediaPresentationDuration\s*=\s*"[^"]*"|mediaPresentationDuration\s*=\s*'[^']*'`)

// PatchPresentationDuration rewrites every mediaPresentationDuration attribute
// occurrence to mediaPresentationDuration="PT{seconds}S" and returns the
// patched XML. If the attribute is absent the input is returned unchanged.
//
// This is a string-level rewrite — segments, timeline, and the rest of the
// document are untouched. DASH players honour the root duration and stop
// playback there, which is sufficient for trial generation.
//
// Caller is responsible for validating seconds; passing 0 or negative values
// produces a syntactically valid but semantically degenerate MPD.
//
// Pure function — safe for concurrent use.
func PatchPresentationDuration(raw string, seconds int) string {
	repl := fmt.Sprintf(`mediaPresentationDuration="PT%dS"`, seconds)
	return trialMPDDurationRe.ReplaceAllString(raw, repl)
}
