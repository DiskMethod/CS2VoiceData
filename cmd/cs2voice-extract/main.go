package main

import (
	"fmt"
	"os"
	"github.com/DiskMethod/cs2-voice-tools/internal/extract"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cs2voice-extract <demo-file>")
		os.Exit(1)
	}
	demoPath := os.Args[1]
	err := extract.ExtractVoiceData(demoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting voice data: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Voice data extraction complete.")
}
