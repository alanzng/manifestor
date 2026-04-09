package manifestor

import "testing"

func TestParseResolution(t *testing.T) {
	tests := []struct {
		input string
		want  Resolution
		err   bool
	}{
		{"1920x1080", Res1080p, false},
		{"1280x720", Res720p, false},
		{"3840x2160", Res4K, false},
		{"100x200", Resolution{100, 200}, false},
		{"1920", Resolution{}, true},
		{"abcx1080", Resolution{}, true},
		{"1920xabc", Resolution{}, true},
	}
	for _, tt := range tests {
		got, err := ParseResolution(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("ParseResolution(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseResolution(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseResolution(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestResolution_String(t *testing.T) {
	if s := Res1080p.String(); s != "1920x1080" {
		t.Errorf("Res1080p.String() = %q, want 1920x1080", s)
	}
}
