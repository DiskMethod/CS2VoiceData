package cli

import (
    "fmt"

    "github.com/DiskMethod/cs2-voice-tools/internal/extract"
    "github.com/spf13/cobra"
)

// NewExtractCmd returns the Cobra command that handles voice extraction.
func NewExtractCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "extract [flags] <demo-file>",
        Short: "Extract voice data from a CS2 demo",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            demoPath := args[0]
            if err := extract.ExtractVoiceData(demoPath); err != nil {
                return err
            }
            fmt.Println("Voice data extraction complete.")
            return nil
        },
    }
    return cmd
}
