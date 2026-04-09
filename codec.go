// Package manifestor provides shared types for HLS and DASH manifest processing.
package manifestor

import "strings"

// Codec identifies a video codec family used for filtering variants/representations.
type Codec string

const (
	// H264 matches avc1.* and avc3.* codec strings.
	H264 Codec = "h264"
	// H265 matches hvc1.* and hev1.* codec strings.
	H265 Codec = "h265"
	// VP9 matches vp09.* and bare "vp9" codec strings.
	VP9 Codec = "vp9"
	// AV1 matches av01.* codec strings.
	AV1 Codec = "av1"
)

// ParseCodec converts a string to a Codec, returning an error if the value is
// not a recognised codec family. Matching is case-insensitive.
func ParseCodec(s string) (Codec, error) {
	switch Codec(strings.ToLower(strings.TrimSpace(s))) {
	case H264:
		return H264, nil
	case H265:
		return H265, nil
	case VP9:
		return VP9, nil
	case AV1:
		return AV1, nil
	}
	return "", &InvalidCodecError{Value: s}
}

// MatchesCodec reports whether the CODECS attribute value (comma-separated list
// of RFC 6381 codec strings) contains a codec belonging to this family.
func (c Codec) MatchesCodec(codecsField string) bool {
	for _, raw := range strings.Split(codecsField, ",") {
		s := strings.ToLower(strings.TrimSpace(raw))
		switch c {
		case H264:
			if strings.HasPrefix(s, "avc1.") || strings.HasPrefix(s, "avc3.") {
				return true
			}
		case H265:
			if strings.HasPrefix(s, "hvc1.") || strings.HasPrefix(s, "hev1.") {
				return true
			}
		case VP9:
			if strings.HasPrefix(s, "vp09.") || s == "vp9" {
				return true
			}
		case AV1:
			if strings.HasPrefix(s, "av01.") {
				return true
			}
		}
	}
	return false
}

// InvalidCodecError is returned by ParseCodec when the input is not a known codec family.
type InvalidCodecError struct {
	Value string
}

func (e *InvalidCodecError) Error() string {
	return "unknown codec: " + e.Value + " (accepted: h264, h265, vp9, av1)"
}
