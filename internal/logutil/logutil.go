// Package logutil provides helpers for conditional verbose logging throughout the cs2-voice-tools project.
package logutil

import "log"

// Verbose controls whether verbose logging is enabled.
var Verbose bool

// VLog logs a message if Verbose is true.
func VLog(format string, args ...any) {
	if Verbose {
		log.Printf(format, args...)
	}
}
