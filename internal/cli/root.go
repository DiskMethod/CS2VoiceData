// Package cli wires up the Cobra command structure for the cs2voice CLI utilities.
package cli

import (
    "os"

    "github.com/DiskMethod/cs2-voice-tools/internal/logutil"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "cs2voice",
    Short: "Suite of CS2 voice utilities",
    Long:  `cs2voice is a single binary providing sub-commands to extract, transcribe, and analyze player voice data from CS2 demos.`,
}

func init() {
    // Global / persistent flags.
    rootCmd.PersistentFlags().BoolVarP(&logutil.Verbose, "verbose", "v", false, "enable verbose output")

    // Sub-commands
    rootCmd.AddCommand(NewExtractCmd())
}

// Execute is the main entry point for running the CLI.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
