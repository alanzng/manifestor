// Package server provides an HTTP proxy server that filters and builds
// HLS and DASH manifests on the fly.
package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	manifestor "github.com/alanzng/manifestor"
	"github.com/alanzng/manifestor/dash"
	"github.com/alanzng/manifestor/hls"
	"github.com/alanzng/manifestor/manifest"
)

// Config holds the configuration for the HTTP server.
type Config struct {
	// Addr is the TCP address to listen on (e.g. ":8080").
	Addr string
	// FetchTimeout is the timeout for fetching upstream manifests.
	// Defaults to 10 seconds.
	FetchTimeout time.Duration
}

// Server is the HTTP proxy server.
type Server struct {
	cfg    Config
	mux    *http.ServeMux
	client *http.Client
}

// New creates a new Server with the given configuration.
func New(cfg Config) *Server {
	if cfg.FetchTimeout == 0 {
		cfg.FetchTimeout = 10 * time.Second
	}
	s := &Server{
		cfg:    cfg,
		mux:    http.NewServeMux(),
		client: &http.Client{Timeout: cfg.FetchTimeout},
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.cfg.Addr, s)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/filter", s.handleFilter)
	s.mux.HandleFunc("/build", s.handleBuild)
}

// handleFilter handles GET /filter requests.
func (s *Server) handleFilter(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	upstreamURL := q.Get("url")
	if upstreamURL == "" {
		http.Error(w, "missing required query param: url", http.StatusBadRequest)
		return
	}

	var opts []manifest.Option

	if codec := q.Get("codec"); codec != "" {
		c, err := manifestor.ParseCodec(codec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		opts = append(opts, manifest.WithCodec(c))
	}
	if maxRes := q.Get("max_res"); maxRes != "" {
		r, err := manifestor.ParseResolution(maxRes)
		if err != nil {
			http.Error(w, "invalid max_res: "+err.Error(), http.StatusBadRequest)
			return
		}
		opts = append(opts, manifest.WithMaxResolution(r))
	}
	if minRes := q.Get("min_res"); minRes != "" {
		r, err := manifestor.ParseResolution(minRes)
		if err != nil {
			http.Error(w, "invalid min_res: "+err.Error(), http.StatusBadRequest)
			return
		}
		opts = append(opts, manifest.WithMinResolution(r))
	}
	if maxBw := q.Get("max_bw"); maxBw != "" {
		v, err := strconv.Atoi(maxBw)
		if err != nil {
			http.Error(w, "invalid max_bw", http.StatusBadRequest)
			return
		}
		opts = append(opts, manifest.WithMaxBandwidth(v))
	}
	if minBw := q.Get("min_bw"); minBw != "" {
		v, err := strconv.Atoi(minBw)
		if err != nil {
			http.Error(w, "invalid min_bw", http.StatusBadRequest)
			return
		}
		opts = append(opts, manifest.WithMinBandwidth(v))
	}
	if fps := q.Get("fps"); fps != "" {
		v, err := strconv.ParseFloat(fps, 64)
		if err != nil {
			http.Error(w, "invalid fps", http.StatusBadRequest)
			return
		}
		opts = append(opts, manifest.WithMaxFrameRate(v))
	}
	if cdn := q.Get("cdn"); cdn != "" {
		opts = append(opts, manifest.WithCDNBaseURL(cdn))
	}
	if token := q.Get("token"); token != "" {
		opts = append(opts, manifest.WithAuthToken(token))
	}
	if lang := q.Get("lang"); lang != "" {
		opts = append(opts, manifest.WithAudioLanguage(lang))
	}
	if origin := q.Get("origin"); origin != "" {
		opts = append(opts, manifest.WithAbsoluteURIs(origin))
	} else {
		if upstreamOrigin := deriveOriginFromUpstreamURL(upstreamURL); upstreamOrigin != "" {
			opts = append(opts, manifest.WithAbsoluteURIs(upstreamOrigin))
		}
	}

	if raw := q.Get("inject_audio"); raw != "" {
		var tracks []injectTrackJSON
		if err := json.Unmarshal([]byte(raw), &tracks); err != nil {
			http.Error(w, "invalid inject_audio: "+err.Error(), http.StatusBadRequest)
			return
		}
		for _, t := range tracks {
			opts = append(opts, injectTrackToOptions(t, false)...)
		}
	}
	if raw := q.Get("inject_subtitle"); raw != "" {
		var tracks []injectTrackJSON
		if err := json.Unmarshal([]byte(raw), &tracks); err != nil {
			http.Error(w, "invalid inject_subtitle: "+err.Error(), http.StatusBadRequest)
			return
		}
		for _, t := range tracks {
			opts = append(opts, injectTrackToOptions(t, true)...)
		}
	}

	result, err := manifest.FilterFromURL(upstreamURL, opts...)
	if err != nil {
		if errors.Is(err, manifest.ErrNoVariantsRemain) {
			http.Error(w, "no variants remain after filtering", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, manifest.ErrFetchFailed) {
			http.Error(w, "failed to fetch upstream manifest", http.StatusBadGateway)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	contentType := detectContentType(result)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(result))
}

// handleBuild handles POST /build requests.
func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req buildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Format == "" {
		http.Error(w, "missing required field: format", http.StatusBadRequest)
		return
	}

	var (
		result string
		err    error
	)

	switch strings.ToLower(req.Format) {
	case "hls":
		result, err = buildHLS(&req)
	case "dash":
		result, err = buildDASH(&req)
	default:
		http.Error(w, "unknown format: "+req.Format, http.StatusBadRequest)
		return
	}

	if err != nil {
		if errors.Is(err, hls.ErrEmptyVariantList) ||
			errors.Is(err, hls.ErrInvalidVariant) ||
			errors.Is(err, dash.ErrEmptyVariantList) ||
			errors.Is(err, dash.ErrInvalidVariant) {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
			http.Error(w, "post-build transform error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	contentType := detectContentType(result)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(result))
}

// detectContentType returns the correct Content-Type for the manifest.
func detectContentType(content string) string {
	if strings.HasPrefix(strings.TrimSpace(content), "#EXTM3U") {
		return "application/vnd.apple.mpegurl"
	}
	return "application/dash+xml"
}

func deriveOriginFromUpstreamURL(upstreamURL string) string {
	u, err := url.Parse(upstreamURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}

	originPath := u.Path
	if originPath == "" {
		originPath = "/"
	} else if idx := strings.LastIndex(originPath, "/"); idx >= 0 {
		originPath = originPath[:idx+1]
	}

	u.Path = originPath
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func buildHLS(req *buildRequest) (string, error) {
	b := hls.NewMasterBuilder()
	if req.Version > 0 {
		b.SetVersion(req.Version)
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
	for _, sub := range req.Subtitles {
		b.AddSubtitleTrack(hls.SubtitleTrackParams{
			GroupID:  sub.GroupID,
			Name:     sub.Name,
			Language: sub.Language,
			URI:      sub.URI,
			Default:  sub.Default,
			Forced:   sub.Forced,
		})
	}
	for _, v := range req.Variants {
		b.AddVariant(hls.VariantParams{
			URI:              v.URI,
			Bandwidth:        v.Bandwidth,
			AverageBandwidth: v.AverageBandwidth,
			Codecs:           v.Codecs,
			Width:            v.Width,
			Height:           v.Height,
			FrameRate:        v.FrameRate,
			AudioGroupID:     v.AudioGroupID,
			SubtitleGroupID:  v.SubtitleGroupID,
			HDCPLevel:        v.HDCPLevel,
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

// JSON schema types for build request.

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
	IndexRange          string `json:"index_range"`
	Initialization      string `json:"initialization"`
	InitializationRange string `json:"initialization_range"`
}

// injectTrackJSON is the unified schema for inject_audio / inject_subtitle query params.
// DASH fields: mime_type, content_type, lang, name, representations.
// HLS fields:  group_id, language, name, uri, default, auto_select, forced.
type injectTrackJSON struct {
	// Shared
	Lang string `json:"lang"`
	Name string `json:"name"`
	// DASH
	MimeType        string               `json:"mime_type"`
	ContentType     string               `json:"content_type"`
	Representations []representationJSON `json:"representations"`
	// HLS
	GroupID    string `json:"group_id"`
	Language   string `json:"language"` // alias for Lang in HLS context
	URI        string `json:"uri"`
	Default    bool   `json:"default"`
	AutoSelect bool   `json:"auto_select"`
	Forced     bool   `json:"forced"`
}

// injectTrackToOptions converts an injectTrackJSON to the appropriate manifest options.
// isSubtitle controls whether HLS subtitle or audio track options are produced.
func injectTrackToOptions(t injectTrackJSON, isSubtitle bool) []manifest.Option {
	var opts []manifest.Option

	lang := t.Lang
	if lang == "" {
		lang = t.Language
	}

	// DASH: inject as AdaptationSet when mime_type or representations are set.
	if t.MimeType != "" || len(t.Representations) > 0 {
		asp := dash.AdaptationSetParams{
			MimeType:    t.MimeType,
			ContentType: t.ContentType,
			Lang:        lang,
			Name:        t.Name,
		}
		if isSubtitle && asp.ContentType == "" {
			asp.ContentType = "text"
		}
		for _, r := range t.Representations {
			rp := dash.RepresentationParams{
				ID:        r.ID,
				Bandwidth: r.Bandwidth,
				Codecs:    r.Codecs,
				Width:     r.Width,
				Height:    r.Height,
				FrameRate: r.FrameRate,
				MimeType:  r.MimeType,
				BaseURL:   r.BaseURL,
			}
			asp.Representations = append(asp.Representations, rp)
		}
		opts = append(opts, manifest.WithDASHInjectAdaptationSet(asp))
	}

	// HLS: inject when uri or group_id is set.
	if t.URI != "" || t.GroupID != "" {
		if isSubtitle {
			opts = append(opts, manifest.WithHLSInjectSubtitle(hls.SubtitleTrackParams{
				GroupID:  t.GroupID,
				Name:     t.Name,
				Language: lang,
				URI:      t.URI,
				Default:  t.Default,
				Forced:   t.Forced,
			}))
		} else {
			opts = append(opts, manifest.WithHLSInjectAudioTrack(hls.AudioTrackParams{
				GroupID:    t.GroupID,
				Name:       t.Name,
				Language:   lang,
				URI:        t.URI,
				Default:    t.Default,
				AutoSelect: t.AutoSelect,
				Forced:     t.Forced,
			}))
		}
	}

	return opts
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
	BaseURL      string `json:"base_url"`
}
