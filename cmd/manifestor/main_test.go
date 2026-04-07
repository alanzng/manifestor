package main_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary into a temp dir.
	dir, err := os.MkdirTemp("", "manifestor-test-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "manifestor")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Determine the module root relative to this test file.
	_, testFile, _, _ := runtime.Caller(0)
	moduleRoot := filepath.Join(filepath.Dir(testFile), "..", "..")

	cmd := exec.Command("/usr/local/go/bin/go", "build", "-o", binaryPath, "./cmd/manifestor/")
	cmd.Dir = moduleRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}

	os.Exit(m.Run())
}

// testdataPath returns the absolute path to a testdata file.
func testdataPath(t *testing.T, parts ...string) string {
	t.Helper()
	_, testFile, _, _ := runtime.Caller(0)
	moduleRoot := filepath.Join(filepath.Dir(testFile), "..", "..")
	return filepath.Join(append([]string{moduleRoot, "testdata"}, parts...)...)
}

func run(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running binary: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestFilterNoArgs(t *testing.T) {
	_, _, code := run(t, "filter")
	if code == 0 {
		t.Error("expected non-zero exit code when no args provided to filter")
	}
}

func TestFilterBothURLAndInput(t *testing.T) {
	fixture := testdataPath(t, "hls", "bento4_mixed_codecs.m3u8")
	_, _, code := run(t, "filter", "--url", "http://example.com/test.m3u8", "--input", fixture)
	if code == 0 {
		t.Error("expected non-zero exit code when both --url and --input are set")
	}
}

func TestFilterInputCodecH264(t *testing.T) {
	fixture := testdataPath(t, "hls", "bento4_mixed_codecs.m3u8")
	stdout, _, code := run(t, "filter", "--input", fixture, "--codec", "h264")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "avc1") {
		t.Errorf("expected output to contain 'avc1', got:\n%s", stdout)
	}
	if strings.Contains(stdout, "hvc1") {
		t.Errorf("expected output to NOT contain 'hvc1', got:\n%s", stdout)
	}
}

func TestFilterInputCodecH265(t *testing.T) {
	fixture := testdataPath(t, "hls", "bento4_mixed_codecs.m3u8")
	stdout, _, code := run(t, "filter", "--input", fixture, "--codec", "h265")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "hvc1") {
		t.Errorf("expected output to contain 'hvc1', got:\n%s", stdout)
	}
}

func TestFilterFromURL(t *testing.T) {
	fixture := testdataPath(t, "hls", "bento4_mixed_codecs.m3u8")
	content, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write(content)
	}))
	defer ts.Close()

	stdout, _, code := run(t, "filter", "--url", ts.URL+"/test.m3u8", "--codec", "h264")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "#EXTM3U") {
		t.Errorf("expected output to start with #EXTM3U, got:\n%s", stdout)
	}
}

func TestBuildHLS(t *testing.T) {
	fixture := testdataPath(t, "build", "hls_simple.json")
	stdout, _, code := run(t, "build", "--format", "hls", "--variants", fixture)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "#EXTM3U") {
		t.Errorf("expected output to start with #EXTM3U, got:\n%s", stdout)
	}
}

func TestBuildDASH(t *testing.T) {
	fixture := testdataPath(t, "build", "dash_simple.json")
	stdout, _, code := run(t, "build", "--format", "dash", "--variants", fixture)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "<MPD") {
		t.Errorf("expected output to contain '<MPD', got:\n%s", stdout)
	}
}

func TestServeRespondsToFilter(t *testing.T) {
	// Start a mock upstream server serving an HLS manifest.
	fixture := testdataPath(t, "hls", "bento4_mixed_codecs.m3u8")
	content, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write(content)
	}))
	defer upstream.Close()

	// Start manifestor serve on a free port.
	port := "19876"
	cmd := exec.Command(binaryPath, "serve", "--port", port)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal("failed to start server:", err)
	}
	defer cmd.Process.Kill()

	// Wait for server to be ready.
	var resp *http.Response
	for i := 0; i < 20; i++ {
		resp, err = http.Get("http://localhost:" + port + "/filter?url=" + upstream.URL + "/test.m3u8")
		if err == nil {
			break
		}
		// brief wait
		cmd.Process.Signal(nil) // no-op, just checking alive
		// sleep ~50ms
		for j := 0; j < 500000; j++ {
		}
	}
	if err != nil {
		t.Fatalf("could not connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
