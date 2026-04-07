package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alanzng/manifestor/server"
)

const sampleHLS = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=3000000,CODECS="avc1.640028",RESOLUTION=1280x720
video-720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1500000,CODECS="avc1.4d401e",RESOLUTION=854x480
video-480p.m3u8
`

func newTestServer() *httptest.Server {
	s := server.New(server.Config{})
	return httptest.NewServer(s)
}

func mockUpstream(content string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte(content))
	}))
}

func TestHandleFilterNoURL(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/filter")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFilterSuccess(t *testing.T) {
	upstream := mockUpstream(sampleHLS)
	defer upstream.Close()

	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/filter?url=" + upstream.URL + "/test.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "mpegurl") {
		t.Errorf("expected mpegurl content type, got %q", ct)
	}
}

func TestHandleFilterCodecH264(t *testing.T) {
	// HLS with mixed codecs.
	mixed := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=3000000,CODECS="avc1.640028",RESOLUTION=1920x1080
video-avc.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=3500000,CODECS="hvc1.1.2.L120.90",RESOLUTION=1280x720
video-hevc.m3u8
`
	upstream := mockUpstream(mixed)
	defer upstream.Close()

	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/filter?url=" + upstream.URL + "/test.m3u8&codec=h264")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	body := buf.String()
	if !strings.Contains(body, "avc1") {
		t.Errorf("expected avc1 in response, got:\n%s", body)
	}
	if strings.Contains(body, "hvc1") {
		t.Errorf("expected hvc1 to be filtered out, got:\n%s", body)
	}
}

func TestHandleBuildHLS(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal(map[string]interface{}{
		"format":  "hls",
		"version": 3,
		"variants": []map[string]interface{}{
			{
				"uri":       "video-720p.m3u8",
				"bandwidth": 3000000,
				"codecs":    "avc1.640028",
				"width":     1280,
				"height":    720,
			},
		},
	})

	resp, err := http.Post(srv.URL+"/build", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	result := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(result), "#EXTM3U") {
		t.Errorf("expected #EXTM3U, got:\n%s", result)
	}
}

func TestHandleBuildDASH(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal(map[string]interface{}{
		"format":   "dash",
		"profile":  "isoff-on-demand",
		"duration": "PT4M0.00S",
		"adaptation_sets": []map[string]interface{}{
			{
				"content_type": "video",
				"mime_type":    "video/mp4",
				"representations": []map[string]interface{}{
					{
						"id":        "v1",
						"bandwidth": 5000000,
						"codecs":    "avc1.640028",
						"width":     1920,
						"height":    1080,
					},
				},
			},
		},
	})

	resp, err := http.Post(srv.URL+"/build", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	result := buf.String()
	if !strings.Contains(result, "<MPD") {
		t.Errorf("expected <MPD in response, got:\n%s", result)
	}
}

func TestHandleBuildMissingFormat(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal(map[string]interface{}{
		"variants": []map[string]interface{}{
			{"uri": "v.m3u8", "bandwidth": 1000000},
		},
	})

	resp, err := http.Post(srv.URL+"/build", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
