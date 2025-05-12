/*
Copyright 2025 Lucas Chagas <lucas.w.chagas@gmail.com>
*/
package cmd

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Options holds global configuration for all commands
type Options struct {
	// Verbose enables detailed debug logging
	Verbose bool

	// OutputDir is the directory where output files will be saved
	OutputDir string

	// AbsOutputDir is the resolved absolute path of OutputDir
	// This is computed during command execution
	AbsOutputDir string
}

// Opts is the global options instance used by all commands
var Opts Options

// Verbose controls whether debug logging is enabled globally (for backward compatibility)
var Verbose bool

// syncVerbose keeps the legacy Verbose variable in sync with Opts.Verbose
func syncVerbose() {
	Verbose = Opts.Verbose
}

// resolveOutputDir ensures the output directory is ready for use
// It resolves the absolute path and creates the directory if needed
func resolveOutputDir() error {
	// Use current directory if no output dir specified
	if Opts.OutputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		Opts.AbsOutputDir = cwd
		return nil
	}

	// If path isn't absolute, make it absolute
	if !filepath.IsAbs(Opts.OutputDir) {
		absPath, err := filepath.Abs(Opts.OutputDir)
		if err != nil {
			return err
		}
		Opts.AbsOutputDir = absPath
	} else {
		Opts.AbsOutputDir = Opts.OutputDir
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(Opts.AbsOutputDir, 0755); err != nil {
		return err
	}

	return nil
}

// Default logger that other packages can import
var Logger *slog.Logger

// SetLogOutput sets the output writer for the logger
// Useful for testing or redirecting logs
func SetLogOutput(w io.Writer) {
	// Sync before using Opts
	syncVerbose()

	level := slog.LevelInfo
	if Opts.Verbose {
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
		// Sync the legacy Verbose variable with the options struct
		syncVerbose()

		// Set up logging based on verbose flag
		logLevel := slog.LevelInfo
		if Opts.Verbose {
			logLevel = slog.LevelDebug
		}

		// Configure the global logger with text handler
		handlerOpts := &slog.HandlerOptions{
			Level: logLevel,
		}
		Logger = slog.New(slog.NewTextHandler(os.Stderr, handlerOpts))

		// Replace the default logger
		slog.SetDefault(Logger)

		// Resolve and prepare output directory
		if err := resolveOutputDir(); err != nil {
			slog.Error("Failed to set up output directory", "error", err)
			os.Exit(1)
		}
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
	// For backward compatibility: we register Verbose directly but also sync with Opts.Verbose
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose output")

	// In the future, register directly with the Options struct for new flags
	rootCmd.PersistentFlags().StringVarP(&Opts.OutputDir, "output-dir", "o", "", "directory to save output files (default: current directory)")

	// Keep options in sync
	Opts.Verbose = Verbose
}
