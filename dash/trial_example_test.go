package dash_test

import (
	"fmt"

	"github.com/datvietvac-techhub/manifestor/dash"
)

func ExamplePatchPresentationDuration() {
	in := `<MPD mediaPresentationDuration="PT4M20S"></MPD>`
	fmt.Println(dash.PatchPresentationDuration(in, 60))
	// Output: <MPD mediaPresentationDuration="PT60S"></MPD>
}
