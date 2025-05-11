/*
Copyright 2025 Lucas Chagas <lucas.w.chagas@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/DiskMethod/cs2-voice-tools/internal/logutil"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "cs2-voice-tools",
	Aliases: []string{"cs2voice"},
	Short:   "Suite of CS2 voice utilities",
	Long: `cs2-voice-tools is a single binary that provides sub-commands to
extract, transcribe, and analyse player voice data from CS2 demo files.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global / persistent flags.
	rootCmd.PersistentFlags().BoolVarP(&logutil.Verbose, "verbose", "v", false, "enable verbose output")
}
