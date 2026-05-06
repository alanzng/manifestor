package dash

import (
	"strings"
	"testing"
)

func TestPatchPresentationDuration_DoubleQuoted(t *testing.T) {
	in := `<?xml version="1.0"?><MPD mediaPresentationDuration="PT4M20S" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011"></MPD>`
	out := PatchPresentationDuration(in, 60)
	if !strings.Contains(out, `mediaPresentationDuration="PT60S"`) {
		t.Errorf("expected PT60S, got: %s", out)
	}
	if strings.Contains(out, "PT4M20S") {
		t.Errorf("old duration not replaced: %s", out)
	}
}

func TestPatchPresentationDuration_SingleQuoted(t *testing.T) {
	in := `<MPD mediaPresentationDuration='PT2H'></MPD>`
	out := PatchPresentationDuration(in, 30)
	if !strings.Contains(out, `mediaPresentationDuration="PT30S"`) {
		t.Errorf("expected PT30S, got: %s", out)
	}
}

func TestPatchPresentationDuration_AttributeAbsent(t *testing.T) {
	in := `<MPD profiles="urn:mpeg:dash:profile:isoff-live:2011"></MPD>`
	out := PatchPresentationDuration(in, 60)
	if out != in {
		t.Errorf("expected no-op when attribute absent, got: %s", out)
	}
}

func TestPatchPresentationDuration_PreservesSurroundingMarkup(t *testing.T) {
	in := `<MPD type="static" mediaPresentationDuration="PT4M" minBufferTime="PT1.5S"><Period/></MPD>`
	out := PatchPresentationDuration(in, 60)
	for _, want := range []string{`type="static"`, `minBufferTime="PT1.5S"`, `<Period/>`} {
		if !strings.Contains(out, want) {
			t.Errorf("surrounding markup mangled, missing %q: %s", want, out)
		}
	}
}

func TestPatchPresentationDuration_ZeroOrNegativeSeconds(t *testing.T) {
	// Match legacy vo-playlist behavior: still rewrites; caller is expected
	// to validate trial duration before calling.
	in := `<MPD mediaPresentationDuration="PT4M"></MPD>`
	out := PatchPresentationDuration(in, 0)
	if !strings.Contains(out, `mediaPresentationDuration="PT0S"`) {
		t.Errorf("expected PT0S for seconds=0, got: %s", out)
	}
}
