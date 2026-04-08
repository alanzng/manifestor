package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alanzng/manifestor/server"
)

func ctxGet(t *testing.T, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func ctxPost(t *testing.T, url, contentType string, body []byte) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return http.DefaultClient.Do(req)
}

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

	resp, err := ctxGet(t, srv.URL+"/filter")
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

	resp, err := ctxGet(t, srv.URL+"/filter?url="+upstream.URL+"/test.m3u8")
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

	resp, err := ctxGet(t, srv.URL+"/filter?url="+upstream.URL+"/test.m3u8&codec=h264")
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

	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
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

	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
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

	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ---- Filter param validation ----

func TestHandleFilterInvalidMaxRes(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://x.com/m.m3u8&max_res=bad")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFilterInvalidMinRes(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://x.com/m.m3u8&min_res=bad")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFilterInvalidMaxBw(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://x.com/m.m3u8&max_bw=notanint")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFilterInvalidMinBw(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://x.com/m.m3u8&min_bw=notanint")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFilterInvalidFps(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://x.com/m.m3u8&fps=notafloat")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFilterNoVariantsRemain(t *testing.T) {
	// HLS with only h265 — filtering for h264 leaves nothing.
	hevcOnly := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=3000000,CODECS="hvc1.1.2.L120.90",RESOLUTION=1920x1080
video-hevc.m3u8
`
	upstream := mockUpstream(hevcOnly)
	defer upstream.Close()

	srv := newTestServer()
	defer srv.Close()

	resp, err := ctxGet(t, srv.URL+"/filter?url="+upstream.URL+"/test.m3u8&codec=h264")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestHandleFilterFetchFailed(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	// Point to a non-existent server.
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://127.0.0.1:19999/no.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", resp.StatusCode)
	}
}

func TestHandleFilterWithAllParams(t *testing.T) {
	upstream := mockUpstream(sampleHLS)
	defer upstream.Close()
	srv := newTestServer()
	defer srv.Close()

	url := srv.URL + "/filter?url=" + upstream.URL + "/test.m3u8" +
		"&codec=h264&max_res=1920x1080&min_res=480x270&max_bw=9000000&min_bw=100000&fps=60" +
		"&cdn=https://cdn.example.com&token=secret&lang=en&origin=https://origin.example.com"
	resp, err := ctxGet(t, url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	// May return 200 or 422 depending on filters — just ensure no 5xx.
	if resp.StatusCode >= 500 {
		t.Errorf("unexpected server error: %d", resp.StatusCode)
	}
}

// ---- Build method/format validation ----

func TestHandleBuildMethodNotAllowed(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/build")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleBuildInvalidJSON(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", []byte("{invalid json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleBuildUnknownFormat(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{"format": "xml"})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleBuildHLSEmptyVariants(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format":   "hls",
		"variants": []interface{}{},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestHandleBuildHLSWithAudioAndIFrames(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format":  "hls",
		"version": 6,
		"audio_tracks": []map[string]interface{}{
			{"group_id": "audio-en", "name": "English", "language": "en", "default": true, "auto_select": true},
		},
		"iframes": []map[string]interface{}{
			{"uri": "iframe.m3u8", "bandwidth": 500000},
		},
		"variants": []map[string]interface{}{
			{"uri": "v.m3u8", "bandwidth": 3000000, "audio_group_id": "audio-en"},
		},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleBuildDASHWithSegmentTemplate(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format":   "dash",
		"profile":  "isoff-live",
		"duration": "PT4M0.00S",
		"adaptation_sets": []map[string]interface{}{
			{
				"content_type": "video",
				"mime_type":    "video/mp4",
				"segment_template": map[string]interface{}{
					"initialization": "$RepresentationID$/init.mp4",
					"media":          "$RepresentationID$/$Number$.m4s",
					"timescale":      90000,
					"duration":       270000,
					"start_number":   1,
				},
				"representations": []map[string]interface{}{
					{"id": "v1", "bandwidth": 3000000, "codecs": "avc1.640028"},
				},
			},
		},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleBuildDASHWithSegmentBase(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format":  "dash",
		"profile": "isoff-on-demand",
		"adaptation_sets": []map[string]interface{}{
			{
				"content_type": "video",
				"mime_type":    "video/mp4",
				"segment_base": map[string]interface{}{
					"index_range":    "0-819",
					"initialization": "0-499",
				},
				"representations": []map[string]interface{}{
					{"id": "v1", "bandwidth": 3000000},
				},
			},
		},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	if !strings.Contains(buf.String(), "<MPD") {
		t.Errorf("expected <MPD in output")
	}
}

func TestHandleBuildWithPostTransforms(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format": "hls",
		"variants": []map[string]interface{}{
			{"uri": "https://origin.example.com/v.m3u8", "bandwidth": 3000000},
		},
		"cdn":   "https://cdn.example.com",
		"token": "abc123",
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
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
	if !strings.Contains(result, "cdn.example.com") {
		t.Errorf("expected CDN rewrite in output, got:\n%s", result)
	}
}

func TestHandleFilterInvalidMaxResHeight(t *testing.T) {
	// Use a mock upstream so that the /filter handler can reach parseResolution.
	// The resolution check happens before the upstream fetch, so we can use any URL.
	srv := newTestServer()
	defer srv.Close()
	resp, err := ctxGet(t, srv.URL+"/filter?url=http://x.com/m.m3u8&max_res=1920xabc")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid height in max_res, got %d", resp.StatusCode)
	}
}

func TestHandleBuildHLSWithSubtitles(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format":  "hls",
		"version": 3,
		"subtitles": []map[string]interface{}{
			{
				"group_id": "subs",
				"name":     "English",
				"language": "en",
				"uri":      "en.vtt",
				"default":  true,
			},
		},
		"variants": []map[string]interface{}{
			{
				"uri":               "v.m3u8",
				"bandwidth":         3000000,
				"subtitle_group_id": "subs",
			},
		},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
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
	if !strings.Contains(result, "TYPE=SUBTITLES") {
		t.Errorf("expected TYPE=SUBTITLES in output, got:\n%s", result)
	}
}

func TestHandleBuildDASHEmptyAdaptationSets(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	// Send a DASH request with empty adaptation_sets — should get ErrEmptyVariantList → 422.
	body, _ := json.Marshal(map[string]interface{}{
		"format":          "dash",
		"adaptation_sets": []interface{}{},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for empty DASH adaptation_sets, got %d", resp.StatusCode)
	}
}

func TestHandleBuildDASHContentType(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format":  "dash",
		"profile": "isoff-on-demand",
		"adaptation_sets": []map[string]interface{}{
			{
				"content_type": "video",
				"mime_type":    "video/mp4",
				"representations": []map[string]interface{}{
					{"id": "v1", "bandwidth": 3000000},
				},
			},
		},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "dash+xml") {
		t.Errorf("expected dash+xml content type, got %q", ct)
	}
}

func TestParseResolution_InvalidWidth(t *testing.T) {
	// "ABCx720" — width is not a number.
	srv := newTestServer()
	defer srv.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1000000\nv.m3u8\n"))
	}))
	defer upstream.Close()

	resp, err := ctxGet(t, srv.URL+"/filter?url="+upstream.URL+"/p.m3u8&max_res=ABCx720")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid resolution width, got %d", resp.StatusCode)
	}
}

func TestHandleBuild_PostBuildWithToken(t *testing.T) {
	// Token on a build request triggers post-build transform.
	srv := newTestServer()
	defer srv.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"format": "hls",
		"token":  "mytoken",
		"variants": []map[string]interface{}{
			{"uri": "https://origin.example.com/v.m3u8", "bandwidth": 3000000},
		},
	})
	resp, err := ctxPost(t, srv.URL+"/build", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "mytoken") {
		t.Errorf("expected token in output, got:\n%s", b)
	}
}

func TestHandleFilter_GenericBadGateway(t *testing.T) {
	// Serve an unparseable manifest to trigger the generic 502 (not ErrFetchFailed).
	srv := newTestServer()
	defer srv.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return garbage — not valid HLS or DASH.
		w.Write([]byte("NOT_A_MANIFEST"))
	}))
	defer upstream.Close()

	resp, err := ctxGet(t, srv.URL+"/filter?url="+upstream.URL+"/p.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502 for unparseable manifest, got %d", resp.StatusCode)
	}
}
