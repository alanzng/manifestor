package dash

import "testing"

func BenchmarkPatchPresentationDuration(b *testing.B) {
	in := `<?xml version="1.0"?><MPD mediaPresentationDuration="PT4M20S" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011"><Period><AdaptationSet/></Period></MPD>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PatchPresentationDuration(in, 60)
	}
}
