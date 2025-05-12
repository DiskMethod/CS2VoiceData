/*
Copyright 2025 Lucas Chagas <lucas.w.chagas@gmail.com>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/DiskMethod/cs2-voice-tools/internal/extract"
	"github.com/spf13/cobra"
)

var (
	// playerFilter is a comma-separated list of SteamID64s to filter by
	playerFilter string

	// steamID64Regex is the regular expression for validating SteamID64 format
	// SteamID64 should be a 17-digit number starting with 7656
	steamID64Regex = regexp.MustCompile(`^7656\d{13}$`)
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [flags] <demo-file>",
	Short: "Extract voice data from a CS2 demo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		demoPath := args[0]

		// Parse player filter if provided
		var playerIDs []string
		var invalidIDs []string

		if playerFilter != "" {
			// Split the comma-separated list and trim whitespace
			for _, id := range strings.Split(playerFilter, ",") {
				// Trim whitespace and ensure non-empty
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}

				// Validate SteamID64 format
				if !steamID64Regex.MatchString(id) {
					slog.Warn("Invalid SteamID64 format, skipping", "id", id)
					invalidIDs = append(invalidIDs, id)
					continue
				}

				playerIDs = append(playerIDs, id)
			}

			// Warn if no valid IDs were provided
			if len(playerIDs) == 0 && len(invalidIDs) > 0 {
				return fmt.Errorf("no valid SteamID64s provided, received: %s", strings.Join(invalidIDs, ", "))
			}
		}

		if err := extract.ExtractVoiceData(demoPath, Opts.AbsOutputDir, Opts.ForceOverwrite, playerIDs); err != nil {
			return err
		}

		msg := fmt.Sprintf("Voice data extraction complete. Files saved to: %s", Opts.AbsOutputDir)
		if len(playerIDs) > 0 {
			msg += fmt.Sprintf(" (filtered to %d players)", len(playerIDs))
		}
		fmt.Println(msg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// Add command-specific flags
	extractCmd.Flags().StringVarP(&playerFilter, "players", "p", "", "filter to specific players by steamID64 (comma-separated list)")
}
