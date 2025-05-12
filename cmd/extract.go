/*
Copyright 2025 Lucas Chagas <lucas.w.chagas@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/DiskMethod/cs2-voice-tools/internal/extract"
	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [flags] <demo-file>",
	Short: "Extract voice data from a CS2 demo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		demoPath := args[0]
		if err := extract.ExtractVoiceData(demoPath, Opts.AbsOutputDir); err != nil {
			return err
		}
		fmt.Printf("Voice data extraction complete. Files saved to: %s\n", Opts.AbsOutputDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// TODO: Add command-specific flags here (e.g., output-dir, format, etc.)
}
