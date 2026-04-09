package manifestor

import "testing"

func TestParseCodec(t *testing.T) {
	tests := []struct {
		input string
		want  Codec
		err   bool
	}{
		{"h264", H264, false},
		{"H264", H264, false},
		{" h265 ", H265, false},
		{"vp9", VP9, false},
		{"av1", AV1, false},
		{"hevc", "", true},
		{"", "", true},
		{"H.264", "", true},
	}
	for _, tt := range tests {
		got, err := ParseCodec(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("ParseCodec(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseCodec(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseCodec(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCodec_MatchesCodec(t *testing.T) {
	tests := []struct {
		codec Codec
		field string
		match bool
	}{
		{H264, "avc1.640028", true},
		{H264, "avc3.640028", true},
		{H264, "hvc1.1.2.L120", false},
		{H265, "hvc1.1.2.L120.90", true},
		{H265, "hev1.1.2.L120.90", true},
		{H265, "avc1.640028", false},
		{VP9, "vp09.00.10.08", true},
		{VP9, "vp9", true},
		{VP9, "avc1.640028", false},
		{AV1, "av01.0.00M.08", true},
		{AV1, "avc1.640028", false},
		{H264, "avc1.640028,mp4a.40.2", true},
		{H264, "hvc1.1.2.L120.90,mp4a.40.2", false},
	}
	for _, tt := range tests {
		got := tt.codec.MatchesCodec(tt.field)
		if got != tt.match {
			t.Errorf("%s.MatchesCodec(%q) = %v, want %v", tt.codec, tt.field, got, tt.match)
		}
	}
}

func TestInvalidCodecError(t *testing.T) {
	_, err := ParseCodec("bad")
	if err == nil {
		t.Fatal("expected error")
	}
	e, ok := err.(*InvalidCodecError)
	if !ok {
		t.Fatalf("expected *InvalidCodecError, got %T", err)
	}
	if e.Value != "bad" {
		t.Errorf("Value = %q, want bad", e.Value)
	}
}
