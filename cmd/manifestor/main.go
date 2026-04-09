// Command manifestor is the CLI tool for parsing, filtering, building, and
// transforming HLS and DASH manifests.
//
// Usage:
//
//	manifestor filter [flags]   — filter a manifest
//	manifestor build  [flags]   — build a manifest from a JSON spec
//	manifestor serve  [flags]   — run the HTTP proxy server
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	manifestor "github.com/alanzng/manifestor"

	"github.com/alanzng/manifestor/dash"
	"github.com/alanzng/manifestor/hls"
	"github.com/alanzng/manifestor/manifest"
	"github.com/alanzng/manifestor/server"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "filter":
		runFilter(os.Args[2:])
	case "build":
		runBuild(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: manifestor <command> [flags]

Commands:
  filter   Fetch or read a manifest, apply filters, and write the result
  build    Build a manifest from a JSON spec file
  serve    Run the HTTP proxy server

Run 'manifestor <command> --help' for command-specific flags.`)
}

func runFilter(args []string) {
	if err := filterManifest(args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// filterManifest implements the filter subcommand logic. It writes the filtered
// manifest to stdout (or to the --output file). Returns an error instead of
// calling os.Exit so that it can be tested without spawning a subprocess.
func filterManifest(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("filter", flag.ContinueOnError)
	urlFlag := fs.String("url", "", "upstream manifest URL")
	input := fs.String("input", "", "local manifest file path")
	output := fs.String("output", "", "output file (default: stdout)")
	codec := fs.String("codec", "", "codec filter: h264|h265|vp9|av1")
	maxRes := fs.String("max-res", "", "max resolution e.g. 1920x1080")
	minRes := fs.String("min-res", "", "min resolution e.g. 854x480")
	maxBw := fs.Int("max-bw", 0, "max bandwidth in bits/s")
	minBw := fs.Int("min-bw", 0, "min bandwidth in bits/s")
	fps := fs.Float64("fps", 0, "max frame rate")
	cdn := fs.String("cdn", "", "CDN base URL")
	token := fs.String("token", "", "auth token appended to URIs")
	lang := fs.String("lang", "", "audio language filter (BCP-47)")
	origin := fs.String("origin", "", "resolve relative URIs against this origin")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate: exactly one of --url or --input.
	if *urlFlag == "" && *input == "" {
		return fmt.Errorf("one of --url or --input is required")
	}
	if *urlFlag != "" && *input != "" {
		return fmt.Errorf("--url and --input are mutually exclusive")
	}

	// Build options.
	var opts []manifest.Option
	if *codec != "" {
		c, err := manifestor.ParseCodec(*codec)
		if err != nil {
			return err
		}
		opts = append(opts, manifest.WithCodec(c))
	}
	if *maxRes != "" {
		r, err := manifestor.ParseResolution(*maxRes)
		if err != nil {
			return fmt.Errorf("invalid --max-res %q: %w", *maxRes, err)
		}
		opts = append(opts, manifest.WithMaxResolution(r))
	}
	if *minRes != "" {
		r, err := manifestor.ParseResolution(*minRes)
		if err != nil {
			return fmt.Errorf("invalid --min-res %q: %w", *minRes, err)
		}
		opts = append(opts, manifest.WithMinResolution(r))
	}
	if *maxBw != 0 {
		opts = append(opts, manifest.WithMaxBandwidth(*maxBw))
	}
	if *minBw != 0 {
		opts = append(opts, manifest.WithMinBandwidth(*minBw))
	}
	if *fps != 0 {
		opts = append(opts, manifest.WithMaxFrameRate(*fps))
	}
	if *cdn != "" {
		opts = append(opts, manifest.WithCDNBaseURL(*cdn))
	}
	if *token != "" {
		opts = append(opts, manifest.WithAuthToken(*token))
	}
	if *lang != "" {
		opts = append(opts, manifest.WithAudioLanguage(*lang))
	}
	if *origin != "" {
		opts = append(opts, manifest.WithAbsoluteURIs(*origin))
	}

	// Load and filter.
	var result string
	var err error
	if *urlFlag != "" {
		result, err = manifest.FilterFromURL(*urlFlag, opts...)
	} else {
		result, err = manifest.FilterFromFile(*input, opts...)
	}
	if err != nil {
		if errors.Is(err, manifest.ErrNoVariantsRemain) {
			return fmt.Errorf("no variants remain after filtering — try relaxing your filter options")
		}
		return err
	}

	// Write output.
	if *output != "" {
		return os.WriteFile(*output, []byte(result), 0o644)
	}
	_, err = fmt.Fprint(stdout, result)
	return err
}

func runBuild(args []string) {
	if err := buildManifest(args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// buildManifest implements the build subcommand logic. It writes the built
// manifest to stdout (or to the --output file). Returns an error instead of
// calling os.Exit so that it can be tested without spawning a subprocess.
func buildManifest(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	format := fs.String("format", "", "manifest format: hls|dash (required)")
	output := fs.String("output", "", "output file (default: stdout)")
	variants := fs.String("variants", "", "path to JSON spec file")
	version := fs.Int("version", 0, "HLS version (HLS only, default 3)")
	duration := fs.String("duration", "", "DASH presentation duration ISO 8601 (DASH only)")
	profile := fs.String("profile", "", "DASH profile: ondemand|live (DASH only)")
	cdn := fs.String("cdn", "", "CDN base URL applied after building")
	token := fs.String("token", "", "auth token appended to all URIs after building")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *format == "" {
		return fmt.Errorf("--format is required (hls or dash)")
	}
	if *variants == "" {
		return fmt.Errorf("--variants JSON file is required")
	}

	data, err := os.ReadFile(*variants)
	if err != nil {
		return fmt.Errorf("reading variants file: %w", err)
	}

	var req buildRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("parsing variants JSON: %w", err)
	}

	// CLI flags override JSON fields.
	if *duration != "" {
		req.Duration = *duration
	}
	if *profile != "" {
		req.Profile = *profile
	}
	if *cdn != "" {
		req.CDN = *cdn
	}
	if *token != "" {
		req.Token = *token
	}

	var result string
	switch strings.ToLower(*format) {
	case "hls":
		result, err = buildHLS(&req, *version)
	case "dash":
		result, err = buildDASH(&req)
	default:
		return fmt.Errorf("unknown format %q (use hls or dash)", *format)
	}
	if err != nil {
		return fmt.Errorf("building manifest: %w", err)
	}

	// Post-build transforms.
	if req.CDN != "" || req.Token != "" {
		var postOpts []manifest.Option
		if req.CDN != "" {
			postOpts = append(postOpts, manifest.WithCDNBaseURL(req.CDN))
		}
		if req.Token != "" {
			postOpts = append(postOpts, manifest.WithAuthToken(req.Token))
		}
		result, err = manifest.Filter(result, postOpts...)
		if err != nil {
			return fmt.Errorf("applying post-build transforms: %w", err)
		}
	}

	if *output != "" {
		return os.WriteFile(*output, []byte(result), 0o644)
	}
	_, err = fmt.Fprint(stdout, result)
	return err
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP port to listen on")
	timeout := fs.Duration("timeout", 0, "upstream fetch timeout (default: 10s)")
	_ = fs.Parse(args)

	fmt.Fprintf(os.Stderr, "Listening on :%d\n", *port)
	s := server.New(server.Config{
		Addr:         fmt.Sprintf(":%d", *port),
		FetchTimeout: *timeout,
	})
	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// buildHLS builds an HLS master playlist from a buildRequest.
func buildHLS(req *buildRequest, versionFlag int) (string, error) {
	b := hls.NewMasterBuilder()
	v := req.Version
	if versionFlag > 0 {
		v = versionFlag
	}
	if v > 0 {
		b.SetVersion(v)
	}

	for _, a := range req.AudioTracks {
		b.AddAudioTrack(hls.AudioTrackParams{
			GroupID:    a.GroupID,
			Name:       a.Name,
			Language:   a.Language,
			URI:        a.URI,
			Default:    a.Default,
			AutoSelect: a.AutoSelect,
			Forced:     a.Forced,
		})
	}
	for _, s := range req.Subtitles {
		b.AddSubtitleTrack(hls.SubtitleTrackParams{
			GroupID:  s.GroupID,
			Name:     s.Name,
			Language: s.Language,
			URI:      s.URI,
			Default:  s.Default,
			Forced:   s.Forced,
		})
	}
	for _, vr := range req.Variants {
		b.AddVariant(hls.VariantParams{
			URI:              vr.URI,
			Bandwidth:        vr.Bandwidth,
			AverageBandwidth: vr.AverageBandwidth,
			Codecs:           vr.Codecs,
			Width:            vr.Width,
			Height:           vr.Height,
			FrameRate:        vr.FrameRate,
			AudioGroupID:     vr.AudioGroupID,
			SubtitleGroupID:  vr.SubtitleGroupID,
			HDCPLevel:        vr.HDCPLevel,
		})
	}
	for _, f := range req.IFrames {
		b.AddIFrameStream(hls.IFrameParams{
			URI:       f.URI,
			Bandwidth: f.Bandwidth,
			Codecs:    f.Codecs,
			Width:     f.Width,
			Height:    f.Height,
		})
	}

	return b.Build()
}

// buildDASH builds a DASH MPD from a buildRequest.
func buildDASH(req *buildRequest) (string, error) {
	cfg := dash.MPDConfig{
		Profile:       req.Profile,
		Duration:      req.Duration,
		MinBufferTime: req.MinBufferTime,
	}
	b := dash.NewMPDBuilder(cfg)

	for _, as := range req.AdaptationSets {
		asp := dash.AdaptationSetParams{
			ContentType: as.ContentType,
			MimeType:    as.MimeType,
			Lang:        as.Lang,
		}
		if as.SegmentTemplate != nil {
			asp.SegmentTemplate = &dash.SegmentTemplateParams{
				Initialization: as.SegmentTemplate.Initialization,
				Media:          as.SegmentTemplate.Media,
				Timescale:      as.SegmentTemplate.Timescale,
				Duration:       as.SegmentTemplate.Duration,
				StartNumber:    as.SegmentTemplate.StartNumber,
			}
		}
		if as.SegmentBase != nil {
			asp.SegmentBase = &dash.SegmentBaseParams{
				IndexRange:     as.SegmentBase.IndexRange,
				Initialization: as.SegmentBase.Initialization,
			}
		}
		for _, r := range as.Representations {
			asp.Representations = append(asp.Representations, dash.RepresentationParams{
				ID:           r.ID,
				Bandwidth:    r.Bandwidth,
				Codecs:       r.Codecs,
				Width:        r.Width,
				Height:       r.Height,
				FrameRate:    r.FrameRate,
				MimeType:     r.MimeType,
				StartWithSAP: r.StartWithSAP,
			})
		}
		b.AddAdaptationSet(asp)
	}

	return b.Build()
}

// writeOutput writes s to the given file path, or stdout if path is empty.
func writeOutput(path, s string) error {
	if path == "" {
		_, err := fmt.Fprint(os.Stdout, s)
		return err
	}
	return os.WriteFile(path, []byte(s), 0o644)
}

// JSON schema structs for build request (also used by server package via shared types file).

type buildRequest struct {
	Format string `json:"format"` // "hls" | "dash"
	// HLS fields
	Version     int                 `json:"version"`
	Variants    []variantJSON       `json:"variants"`
	AudioTracks []audioTrackJSON    `json:"audio_tracks"`
	Subtitles   []subtitleTrackJSON `json:"subtitles"`
	IFrames     []iframeJSON        `json:"iframes"`
	// DASH fields
	Profile        string              `json:"profile"`
	Duration       string              `json:"duration"`
	MinBufferTime  string              `json:"min_buffer_time"`
	AdaptationSets []adaptationSetJSON `json:"adaptation_sets"`
	// Post-build transforms
	CDN   string `json:"cdn"`
	Token string `json:"token"`
}

type variantJSON struct {
	URI              string  `json:"uri"`
	Bandwidth        int     `json:"bandwidth"`
	AverageBandwidth int     `json:"average_bandwidth"`
	Codecs           string  `json:"codecs"`
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	FrameRate        float64 `json:"frame_rate"`
	AudioGroupID     string  `json:"audio_group_id"`
	SubtitleGroupID  string  `json:"subtitle_group_id"`
	HDCPLevel        string  `json:"hdcp_level"`
}

type audioTrackJSON struct {
	GroupID    string `json:"group_id"`
	Name       string `json:"name"`
	Language   string `json:"language"`
	URI        string `json:"uri"`
	Default    bool   `json:"default"`
	AutoSelect bool   `json:"auto_select"`
	Forced     bool   `json:"forced"`
}

type subtitleTrackJSON struct {
	GroupID  string `json:"group_id"`
	Name     string `json:"name"`
	Language string `json:"language"`
	URI      string `json:"uri"`
	Default  bool   `json:"default"`
	Forced   bool   `json:"forced"`
}

type iframeJSON struct {
	URI       string `json:"uri"`
	Bandwidth int    `json:"bandwidth"`
	Codecs    string `json:"codecs"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type adaptationSetJSON struct {
	ContentType     string               `json:"content_type"`
	MimeType        string               `json:"mime_type"`
	Lang            string               `json:"lang"`
	SegmentTemplate *segmentTemplateJSON `json:"segment_template"`
	SegmentBase     *segmentBaseJSON     `json:"segment_base"`
	Representations []representationJSON `json:"representations"`
}

type segmentTemplateJSON struct {
	Initialization string `json:"initialization"`
	Media          string `json:"media"`
	Timescale      int    `json:"timescale"`
	Duration       int    `json:"duration"`
	StartNumber    int    `json:"start_number"`
}

type segmentBaseJSON struct {
	IndexRange     string `json:"index_range"`
	Initialization string `json:"initialization"`
}

type representationJSON struct {
	ID           string `json:"id"`
	Bandwidth    int    `json:"bandwidth"`
	Codecs       string `json:"codecs"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FrameRate    string `json:"frame_rate"`
	MimeType     string `json:"mime_type"`
	StartWithSAP int    `json:"start_with_sap"`
}
