package extract

import (
	"github.com/DiskMethod/cs2-voice-tools/internal/decoder"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

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
	defaultOpusSampleRate  = 48000
	// defaultNumChannels is the number of audio channels (mono audio).
	defaultNumChannels     = 1
	// defaultBitDepth is the bit depth for output WAV files.
	defaultBitDepth        = 32
	// intPCMMaxValue is the maximum integer value for PCM normalization.
	intPCMMaxValue         = 2147483647
)

// ExtractVoiceData parses a CS2 demo file and writes per-player WAV files containing voice data.
func ExtractVoiceData(demoPath string) error {
	voiceDataPerPlayer := map[string][][]byte{}

	file, err := os.Open(demoPath)
	if err != nil {
		return fmt.Errorf("failed to open demo file '%s': %w", demoPath, err)
	}
	defer file.Close()

	parser := dem.NewParser(file)
	var format string

	parser.RegisterNetMessageHandler(func(m *msgs2.CSVCMsg_VoiceData) {
		steamId := strconv.Itoa(int(m.GetXuid()))
		format = m.Audio.Format.String()
		voiceDataPerPlayer[steamId] = append(voiceDataPerPlayer[steamId], m.Audio.VoiceData)
	})

	err = parser.ParseToEnd()
	if errors.Is(err, dem.ErrCancelled) {
		log.Println("Parsing was cancelled.")
	} else if errors.Is(err, dem.ErrUnexpectedEndOfDemo) {
		log.Println("Demo file ended unexpectedly (may be corrupt).")
	} else if errors.Is(err, dem.ErrInvalidFileType) {
		log.Println("Invalid demo file type.")
	} else if err != nil {
		log.Printf("Unknown error parsing demo: %v", err)
	}

	for playerId, voiceData := range voiceDataPerPlayer {
		wavFilePath := fmt.Sprintf("%s.wav", playerId)
		if format == "VOICEDATA_FORMAT_OPUS" {
			err = opusToWav(voiceData, wavFilePath)
			if err != nil {
				log.Printf("failed to initialize OpusDecoder: %v", err)
				continue
			}
		} else if format == "VOICEDATA_FORMAT_STEAM" {
			err = convertAudioDataToWavFiles(voiceData, wavFilePath)
			if err != nil {
				log.Printf("failed to write wav for player %s: %v", playerId, err)
			}
		}
	}

	defer parser.Close()
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
func opusToWav(data [][]byte, wavName string) (err error) {
	opusDecoder, err := decoder.NewDecoder(defaultOpusSampleRate, defaultNumChannels)
	if err != nil {
		return
	}
	var pcmBuffer []int
	for _, d := range data {
		pcm, err := decoder.Decode(opusDecoder, d)
		if err != nil {
			log.Printf("failed to decode Opus data: %v", err)
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
		return
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
		return
	}
	return
}
