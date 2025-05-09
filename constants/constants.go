// Package constants provides centralized protocol constants for the CS2 Voice Data extractor.
// These constants are shared across the main logic and decoder package to avoid magic numbers.
//
// See: https://zhenyangli.me/posts/reversing-steam-voice-codec/
// Package constants centralizes protocol and audio configuration constants for the CS2 voice data tool.
package constants

const (
	// Default sample rate for Steam voice data (reverse engineered, see blog)
	DefaultSteamSampleRate = 24000
	// Default sample rate for Opus voice data
	DefaultOpusSampleRate  = 48000
	// Number of channels for voice data
	DefaultNumChannels     = 1
	// Bit depth for WAV encoding
	DefaultBitDepth        = 32
	// Int PCM conversion multiplier (for float32 -> int)
	IntPCMMaxValue         = 2147483647
)
