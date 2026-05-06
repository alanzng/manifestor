package hls

import (
	"path"
	"strings"
)

// AddTrialSuffix transforms "foo.m3u8" into "foo_trial.m3u8". It splits on
// the final extension as reported by path.Ext; if there is no extension the
// suffix is appended directly. URIs containing a query string have the suffix
// inserted before whatever path.Ext considers the extension — callers passing
// signed URLs should strip the query first if they need the suffix to land
// before the path's true extension.
func AddTrialSuffix(uri string) string {
	ext := path.Ext(uri)
	return strings.TrimSuffix(uri, ext) + "_trial" + ext
}
