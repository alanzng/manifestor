package hls_test

import (
	"fmt"

	"github.com/datvietvac-techhub/manifestor/hls"
)

func ExampleSliceMediaPlaylist() {
	raw := "#EXTM3U\n" +
		"#EXT-X-VERSION:6\n" +
		"#EXT-X-TARGETDURATION:4\n" +
		"#EXTINF:4.000,\nseg1.m4s\n" +
		"#EXTINF:4.000,\nseg2.m4s\n" +
		"#EXTINF:4.000,\nseg3.m4s\n" +
		"#EXT-X-ENDLIST\n"

	fmt.Print(hls.SliceMediaPlaylist(raw, 1))
	// Output:
	// #EXTM3U
	// #EXT-X-VERSION:6
	// #EXT-X-TARGETDURATION:4
	// #EXTINF:4.000,
	// seg1.m4s
	// #EXT-X-ENDLIST
}

func ExampleSliceMediaPlaylistByDuration() {
	raw := "#EXTM3U\n#EXT-X-TARGETDURATION:6\n" +
		"#EXTINF:6.0,\na.ts\n#EXTINF:6.0,\nb.ts\n#EXTINF:6.0,\nc.ts\n#EXT-X-ENDLIST\n"

	out, n := hls.SliceMediaPlaylistByDuration(raw, 12, 4)
	fmt.Println("kept", n, "segments")
	fmt.Print(out)
	// Output:
	// kept 2 segments
	// #EXTM3U
	// #EXT-X-TARGETDURATION:6
	// #EXTINF:6.0,
	// a.ts
	// #EXTINF:6.0,
	// b.ts
	// #EXT-X-ENDLIST
}

func ExampleAddTrialSuffix() {
	fmt.Println(hls.AddTrialSuffix("video_720p.m3u8"))
	// Output: video_720p_trial.m3u8
}
