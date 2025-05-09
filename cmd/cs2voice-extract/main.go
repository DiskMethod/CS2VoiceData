package main

import (
	"fmt"
	"os"
	"CS2VoiceData/internal/extract"
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
