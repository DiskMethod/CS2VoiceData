// Package main provides the entry point for the cs2voice-extract CLI tool.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DiskMethod/cs2-voice-tools/internal/extract"
	"github.com/DiskMethod/cs2-voice-tools/internal/logutil"
)

// main is the entry point for the cs2voice-extract command-line tool.
func main() {
	flag.BoolVar(&logutil.Verbose, "verbose", false, "enable verbose output")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <demo-file>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
	demoPath := flag.Arg(0)
	err := extract.ExtractVoiceData(demoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting voice data: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Voice data extraction complete.")
}
