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
	"flag"
	"fmt"
	"os"
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
	fs := flag.NewFlagSet("filter", flag.ExitOnError)
	url := fs.String("url", "", "upstream manifest URL")
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
	_ = fs.Parse(args)

	// TODO: implement
	_ = url
	_ = input
	_ = output
	_ = codec
	_ = maxRes
	_ = minRes
	_ = maxBw
	_ = minBw
	_ = fps
	_ = cdn
	_ = token
	fmt.Fprintln(os.Stderr, "filter: not implemented")
	os.Exit(1)
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	format := fs.String("format", "", "manifest format: hls|dash (required)")
	output := fs.String("output", "", "output file (default: stdout)")
	variants := fs.String("variants", "", "path to JSON spec file")
	version := fs.Int("version", 3, "HLS version (HLS only)")
	duration := fs.String("duration", "", "DASH presentation duration ISO 8601 (DASH only)")
	profile := fs.String("profile", "", "DASH profile: ondemand|live (DASH only)")
	cdn := fs.String("cdn", "", "CDN base URL applied after building")
	token := fs.String("token", "", "auth token appended to all URIs after building")
	_ = fs.Parse(args)

	// TODO: implement
	_ = format
	_ = output
	_ = variants
	_ = version
	_ = duration
	_ = profile
	_ = cdn
	_ = token
	fmt.Fprintln(os.Stderr, "build: not implemented")
	os.Exit(1)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP port to listen on")
	timeout := fs.Duration("timeout", 0, "upstream fetch timeout (default: 10s)")
	_ = fs.Parse(args)

	// TODO: implement
	_ = port
	_ = timeout
	fmt.Fprintln(os.Stderr, "serve: not implemented")
	os.Exit(1)
}
