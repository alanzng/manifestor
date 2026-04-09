package manifestor

import (
	"errors"
	"strconv"
	"strings"
)

// Resolution represents a width × height pair in pixels.
type Resolution struct {
	Width  int
	Height int
}

// Common resolution presets.
var (
	Res360p  = Resolution{640, 360}
	Res480p  = Resolution{854, 480}
	Res720p  = Resolution{1280, 720}
	Res1080p = Resolution{1920, 1080}
	Res1440p = Resolution{2560, 1440}
	Res4K    = Resolution{3840, 2160}
)

// ParseResolution parses a "WxH" string into a Resolution.
func ParseResolution(s string) (Resolution, error) {
	parts := strings.SplitN(s, "x", 2)
	if len(parts) != 2 {
		return Resolution{}, errors.New("expected WxH format")
	}
	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return Resolution{}, errors.New("invalid width")
	}
	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return Resolution{}, errors.New("invalid height")
	}
	return Resolution{Width: w, Height: h}, nil
}

// String returns the resolution in "WxH" format.
func (r Resolution) String() string {
	return strconv.Itoa(r.Width) + "x" + strconv.Itoa(r.Height)
}
