// Package extract provides functions to extract and convert voice data from CS2 demo files into audio files.
package extract

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DiskMethod/cs2-voice-tools/internal/decoder"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/msgs2"
)

// Default audio parameters for decoding CS2 demo voice data.
const (
	// defaultSteamSampleRate is the sample rate (Hz) for Steam-format voice data.
	defaultSteamSampleRate = 24000
	// defaultOpusSampleRate is the sample rate (Hz) for Opus-format voice data.
	defaultOpusSampleRate = 48000
	// defaultNumChannels is the number of audio channels (mono audio).
	defaultNumChannels = 1
	// defaultBitDepth is the bit depth for output WAV files.
	defaultBitDepth = 32
	// intPCMMaxValue is the maximum integer value for PCM normalization.
	intPCMMaxValue = 2147483647
)

// ExtractVoiceData parses a CS2 demo file and writes per-player audio files containing voice data.
// The outputDir parameter specifies where to save the extracted files.
// When forceOverwrite is false, the function will not overwrite existing files.
// If playerIDs is provided, only extracts voice data for those specific players.
// The format parameter specifies the desired output audio format (wav, mp3, etc.).
func ExtractVoiceData(demoPath, outputDir string, forceOverwrite bool, playerIDs []string, format string) error {
	// Convert playerIDs slice to a map for O(1) lookups
	playerFilter := make(map[string]bool)
	for _, id := range playerIDs {
		playerFilter[id] = true
	}

	// Track which requested players were found
	foundPlayers := make(map[string]bool)
	voiceDataPerPlayer := map[string][][]byte{}

	slog.Debug("Opening demo file", "path", demoPath)
	file, err := os.Open(demoPath)
	if err != nil {
		return fmt.Errorf("failed to open demo file '%s': %w", demoPath, err)
	}
	defer file.Close()

	parser := dem.NewParser(file)
	var voiceDataFormat string

	parser.RegisterNetMessageHandler(func(m *msgs2.CSVCMsg_VoiceData) {
		steamId := strconv.Itoa(int(m.GetXuid()))
		voiceDataFormat = m.Audio.Format.String()
		voiceDataPerPlayer[steamId] = append(voiceDataPerPlayer[steamId], m.Audio.VoiceData)
	})

	err = parser.ParseToEnd()
	if err != nil {
		if errors.Is(err, dem.ErrCancelled) {
			return fmt.Errorf("parsing was cancelled: %w", err)
		} else if errors.Is(err, dem.ErrUnexpectedEndOfDemo) {
			return fmt.Errorf("demo file ended unexpectedly (may be corrupt): %w", err)
		} else if errors.Is(err, dem.ErrInvalidFileType) {
			return fmt.Errorf("invalid demo file type: %w", err)
		}
		return fmt.Errorf("unknown error parsing demo: %w", err)
	}

	slog.Debug("Found players with voice data", "count", len(voiceDataPerPlayer))

	// Create a temporary directory for intermediate WAV files
	tempDir, err := os.MkdirTemp("", "cs2voice-tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	// Ensure temporary directory cleanup on function exit
	defer os.RemoveAll(tempDir)

	slog.Debug("Created temporary directory for processing", "path", tempDir)

	for playerId, voiceData := range voiceDataPerPlayer {
		// Apply player filter if provided
		if len(playerFilter) > 0 && !playerFilter[playerId] {
			slog.Debug("Skipping player (not in filter)", "player", playerId)
			continue
		}

		// Mark this player as found if it was in the filter
		if playerFilter[playerId] {
			foundPlayers[playerId] = true
		}

		// Set up temporary and final paths
		tempWavPath := filepath.Join(tempDir, fmt.Sprintf("%s.wav", playerId))
		finalOutputPath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", playerId, format))

		// Check if file already exists and respect forceOverwrite flag
		if _, err := os.Stat(finalOutputPath); err == nil && !forceOverwrite {
			slog.Warn("File already exists, skipping", "path", finalOutputPath)
			continue
		} else if !os.IsNotExist(err) && err != nil {
			// Some other error occurred checking the file
			slog.Error("Failed to check file existence", "path", finalOutputPath, "error", err)
			continue
		}

		var err error
		// Generate the WAV file in the temporary directory
		if voiceDataFormat == "VOICEDATA_FORMAT_OPUS" {
			err = opusToWav(voiceData, tempWavPath)
			if err != nil {
				slog.Error("Failed to initialize OpusDecoder", "error", err)
				continue
			}
		} else if voiceDataFormat == "VOICEDATA_FORMAT_STEAM" {
			err = convertAudioDataToWavFiles(voiceData, tempWavPath)
			if err != nil {
				slog.Error("Failed to write WAV file", "player", playerId, "error", err)
				continue
			}
		} else {
			slog.Warn("Unknown voice data format", "format", voiceDataFormat)
			continue
		}

		// For WAV format, copy from temp to output directory
		if format == "wav" {
			// Read the temporary file
			wavData, err := os.ReadFile(tempWavPath)
			if err != nil {
				slog.Error("Failed to read temporary WAV file", "path", tempWavPath, "error", err)
				continue
			}

			// Write to the final location
			err = os.WriteFile(finalOutputPath, wavData, 0644)
			if err != nil {
				slog.Error("Failed to write WAV file to final location", "path", finalOutputPath, "error", err)
				continue
			}

			slog.Debug("Audio file created successfully", "player", playerId, "path", finalOutputPath)
			continue
		}

		// Convert to the desired format if needed
		err = convertAudioToFormat(tempWavPath, finalOutputPath, format)
		if err != nil {
			slog.Error("Failed to convert audio format", "player", playerId, "format", format, "error", err)
			continue
		}

		slog.Debug("Audio file created successfully", "player", playerId, "path", finalOutputPath)
	}

	defer parser.Close()

	// Log information about player filter results
	if len(playerFilter) > 0 {
		slog.Debug("Player filter results", "requested", len(playerFilter), "found", len(foundPlayers))

		// Check if any requested players were not found
		if len(foundPlayers) < len(playerFilter) {
			for id := range playerFilter {
				if !foundPlayers[id] {
					slog.Warn("Requested player not found in demo", "player", id)
				}
			}
		}
	}

	slog.Debug("Extraction complete for demo file", "path", demoPath)
	return nil
}

// convertAudioToFormat uses ffmpeg to convert a WAV file to the specified format
// Takes source WAV path, destination path, and format as parameters
func convertAudioToFormat(wavPath string, outputPath string, format string) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Build the ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", wavPath,           // Input file
		"-y",                    // Overwrite output file
		"-loglevel", "error",    // Only show errors
		"-hide_banner",          // Hide the banner
		outputPath)              // Output file

	// Capture stderr for error reporting
	var stderr strings.Builder
	cmd.Stderr = &stderr

	// Run the command
	slog.Debug("Converting audio", "from", wavPath, "to", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w: %s", err, stderr.String())
	}

	return nil
}

// convertAudioDataToWavFiles decodes Steam-format voice data payloads and writes them to a WAV file.
// It uses the Opus decoder for each chunk and encodes the PCM output as a WAV file. Returns an error if any operation fails.
func convertAudioDataToWavFiles(payloads [][]byte, fileName string) error {
	voiceDecoder, err := decoder.NewOpusDecoder(defaultSteamSampleRate, defaultNumChannels)
	if err != nil {
		return fmt.Errorf("failed to initialize OpusDecoder: %w", err)
	}
	o := make([]int, 0, 1024)
	for _, payload := range payloads {
		c, err := decoder.DecodeChunk(payload)
		if err != nil {
			return fmt.Errorf("failed to decode chunk: %w", err)
		}
		if c != nil && len(c.Data) > 0 {
			pcm, err := voiceDecoder.Decode(c.Data)
			if err != nil {
				return fmt.Errorf("failed to decode Opus frame: %w", err)
			}
			converted := make([]int, len(pcm))
			for i, v := range pcm {
				converted[i] = int(v * intPCMMaxValue)
			}
			o = append(o, converted...)
		}
	}
	outFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create wav file: %w", err)
	}
	defer outFile.Close()
	enc := wav.NewEncoder(outFile, defaultSteamSampleRate, defaultBitDepth, defaultNumChannels, 1)
	buf := &audio.IntBuffer{
		Data: o,
		Format: &audio.Format{
			SampleRate:  defaultSteamSampleRate,
			NumChannels: defaultNumChannels,
		},
	}
	if err := enc.Write(buf); err != nil {
		return fmt.Errorf("failed to write WAV data: %w", err)
	}
	enc.Close()
	return nil
}

// opusToWav decodes Opus-format voice data and writes the result to a WAV file.
// Returns an error if decoding or file writing fails.
func opusToWav(data [][]byte, wavName string) error {
	opusDecoder, err := decoder.NewDecoder(defaultOpusSampleRate, defaultNumChannels)
	if err != nil {
		return fmt.Errorf("failed to initialize OpusDecoder: %w", err)
	}
	var pcmBuffer []int
	for _, d := range data {
		pcm, err := decoder.Decode(opusDecoder, d)
		if err != nil {
			slog.Warn("Failed to decode Opus data", "error", err)
			continue
		}
		pp := make([]int, len(pcm))
		for i, p := range pcm {
			pp[i] = int(p * intPCMMaxValue)
		}
		pcmBuffer = append(pcmBuffer, pp...)
	}
	file, err := os.Create(wavName)
	if err != nil {
		return fmt.Errorf("failed to create wav file: %w", err)
	}
	defer file.Close()
	enc := wav.NewEncoder(file, defaultOpusSampleRate, defaultBitDepth, defaultNumChannels, 1)
	defer enc.Close()
	buffer := &audio.IntBuffer{
		Data: pcmBuffer,
		Format: &audio.Format{
			SampleRate:  defaultOpusSampleRate,
			NumChannels: defaultNumChannels,
		},
	}
	err = enc.Write(buffer)
	if err != nil {
		return fmt.Errorf("failed to write WAV data: %w", err)
	}
	return nil
}
