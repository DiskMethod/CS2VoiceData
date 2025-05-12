// Package extract provides functions to extract and convert voice data from CS2 demo files into audio files.
package extract

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// Common error types for the extraction process
var (
	// ErrNoVoiceData is returned when no voice data is found in the demo
	ErrNoVoiceData = errors.New("no voice data found in demo")

	// ErrInvalidFormat is returned when an unsupported format is specified
	ErrInvalidFormat = errors.New("invalid audio format")

	// ErrFFMPEGNotFound is returned when ffmpeg is not available for conversion
	ErrFFMPEGNotFound = errors.New("ffmpeg not found")

	// ErrOutputDirNotWritable is returned when the output directory cannot be written to
	ErrOutputDirNotWritable = errors.New("output directory is not writable")

	// supportedFormats is the list of audio formats supported by this tool
	supportedFormats = []string{"wav", "mp3", "ogg", "flac", "aac", "m4a"}
)

// ExtractOptions contains all configuration options for the voice data extraction process.
type ExtractOptions struct {
	// DemoPath is the path to the CS2 demo file
	DemoPath string

	// OutputDir is the directory where extracted audio files will be saved
	OutputDir string

	// ForceOverwrite determines whether existing files should be overwritten
	ForceOverwrite bool

	// PlayerIDs is an optional slice of SteamID64s to filter by
	// If empty, all players' voice data will be extracted
	PlayerIDs []string

	// Format specifies the output audio format (wav, mp3, ogg, etc.)
	Format string
}

// validateFormat checks if the given format is supported.
// Returns nil if valid, or an error with suggestions otherwise.
func validateFormat(format string) error {
	for _, f := range supportedFormats {
		if f == format {
			return nil
		}
	}
	return fmt.Errorf("%w: '%s' (supported formats: %s)",
		ErrInvalidFormat, format, strings.Join(supportedFormats, ", "))
}

// sanitizeFilename removes or replaces characters that are unsafe for filenames across platforms.
// This ensures generated filenames are valid on various operating systems.
func sanitizeFilename(name string) string {
	// Replace unsafe characters with underscores
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := re.ReplaceAllString(name, "_")

	// Trim leading/trailing periods and spaces which can cause issues
	sanitized = strings.Trim(sanitized, " .")

	// If the sanitization process results in an empty string, provide a fallback
	if sanitized == "" {
		return "player"
	}

	return sanitized
}

// checkOutputDirectory verifies that the output directory exists and is writable.
// If the directory doesn't exist, it attempts to create it.
func checkOutputDirectory(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to access output directory: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("output path exists but is not a directory: %s", dir)
	}

	// Check if it's writable by creating and immediately removing a test file
	testFile := filepath.Join(dir, ".cs2voice-write-test")
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		return fmt.Errorf("output directory is not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// ExtractVoiceData parses a CS2 demo file and writes per-player audio files containing voice data.
// Uses the provided options to configure the extraction process.
func ExtractVoiceData(opts ExtractOptions) error {
	// Validate required fields
	if opts.DemoPath == "" {
		return fmt.Errorf("demo path is required")
	}

	if opts.OutputDir == "" {
		// Default to current directory if not specified
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.OutputDir = cwd
	}

	// Default to WAV if no format specified
	if opts.Format == "" {
		opts.Format = "wav"
	} else {
		// Validate format
		opts.Format = strings.ToLower(opts.Format)
		if err := validateFormat(opts.Format); err != nil {
			return err
		}
	}

	// Convert playerIDs slice to a map for O(1) lookups
	playerFilter := make(map[string]bool)
	for _, id := range opts.PlayerIDs {
		playerFilter[id] = true
	}

	// Track which requested players were found
	foundPlayers := make(map[string]bool)
	voiceDataPerPlayer := map[string][][]byte{}

	slog.Debug("Opening demo file", "path", opts.DemoPath)
	file, err := os.Open(opts.DemoPath)
	if err != nil {
		return fmt.Errorf("failed to open demo file '%s': %w", opts.DemoPath, err)
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

	// Check if no voice data was found
	if len(voiceDataPerPlayer) == 0 {
		return ErrNoVoiceData
	}

	// Check if the output directory exists and is writable
	if err := checkOutputDirectory(opts.OutputDir); err != nil {
		return fmt.Errorf("output directory issue: %w", err)
	}

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

		// Sanitize the player ID for filename safety
		safePlayerId := sanitizeFilename(playerId)

		// Set up paths
		var tempWavPath, finalOutputPath string

		// For WAV format, optimize by writing directly to the final path
		if opts.Format == "wav" {
			// Write directly to the output directory, skipping the temporary file
			finalOutputPath = filepath.Join(opts.OutputDir, fmt.Sprintf("%s.wav", safePlayerId))
			tempWavPath = finalOutputPath // Both point to the same location
		} else {
			// For other formats, use the temporary directory for WAV files
			tempWavPath = filepath.Join(tempDir, fmt.Sprintf("%s.wav", safePlayerId))
			finalOutputPath = filepath.Join(opts.OutputDir, fmt.Sprintf("%s.%s", safePlayerId, opts.Format))
		}

		// Check if file already exists and respect ForceOverwrite flag
		if _, err := os.Stat(finalOutputPath); err == nil && !opts.ForceOverwrite {
			slog.Warn("File already exists, skipping", "path", finalOutputPath)
			continue
		} else if !os.IsNotExist(err) && err != nil {
			// Some other error occurred checking the file
			slog.Error("Failed to check file existence", "path", finalOutputPath, "error", err)
			continue
		}

		var err error
		// Generate the WAV file (either temporary or final for WAV format)
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

		// For WAV format, optimize by writing directly to the final path
		if opts.Format == "wav" {
			// Since we know the output format is WAV, skip the temporary file.
			// Write directly to the output directory
			finalOutputPath = filepath.Join(opts.OutputDir, fmt.Sprintf("%s.wav", safePlayerId))

			// For direct WAV output, overwrite tempWavPath to point to our final destination
			tempWavPath = finalOutputPath

			// The remaining code will now write directly to the final location
			// And we'll skip the conversion step since we continue below

			// After the generate step completes, we're done - no need for conversion
			slog.Debug("Audio file created successfully", "player", playerId, "path", finalOutputPath)
			continue
		}

		// Convert to the desired format if needed
		err = convertAudioToFormat(tempWavPath, finalOutputPath, opts.Format)
		if err != nil {
			slog.Error("Failed to convert audio format", "player", playerId, "format", opts.Format, "error", err)
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

	slog.Debug("Extraction complete",
		"demo", opts.DemoPath,
		"outputDir", opts.OutputDir,
		"format", opts.Format)
	return nil
}

// convertAudioToFormat uses ffmpeg to convert a WAV file to the specified format
// Takes source WAV path, destination path, and format as parameters
func convertAudioToFormat(wavPath string, outputPath string, format string) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("%w: %v", ErrFFMPEGNotFound, err)
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
