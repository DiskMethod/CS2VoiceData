/*
Copyright 2025 Lucas Chagas <lucas.w.chagas@gmail.com>
*/
package cmd

import (
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Verbose controls whether debug logging is enabled globally
var Verbose bool

// Default logger that other packages can import
var Logger *slog.Logger

// SetLogOutput sets the output writer for the logger
// Useful for testing or redirecting logs
func SetLogOutput(w io.Writer) {
	level := slog.LevelInfo
	if Verbose {
		level = slog.LevelDebug
	}

	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}
	Logger = slog.New(slog.NewTextHandler(w, handlerOpts))
	slog.SetDefault(Logger)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "cs2-voice-tools",
	Aliases: []string{"cs2voice"},
	Short:   "Suite of CS2 voice utilities",
	Long: `cs2-voice-tools is a single binary that provides sub-commands to
extract, transcribe, and analyse player voice data from CS2 demo files.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging based on verbose flag
		logLevel := slog.LevelInfo
		if Verbose {
			logLevel = slog.LevelDebug
		}

		// Configure the global logger with text handler
		handlerOpts := &slog.HandlerOptions{
			Level: logLevel,
		}
		Logger = slog.New(slog.NewTextHandler(os.Stderr, handlerOpts))

		// Replace the default logger
		slog.SetDefault(Logger)
	},
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
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose output")
}
