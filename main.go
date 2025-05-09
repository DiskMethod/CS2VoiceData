// Package main implements the CS2 voice data extraction tool.
// It parses CS2 demo files and extracts player voice communications into WAV files.
package main

import (
	"CS2VoiceData/constants"
	"CS2VoiceData/decoder"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/msgs2"
)

// main is the entry point for the CS2 voice data extraction tool.
// It parses command-line arguments, processes the demo file, and writes WAV files for each player's voice data.
func main() {
	// Use a command-line flag for the demo file path for Go idiomatic CLI behavior
	demoPath := flag.String("demo", "", "Path to the unzipped CS2 demo file")
	flag.Parse()

	if *demoPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: ./main -demo <path-to-demo-file>")
		os.Exit(1)
	}

	// Create a map of a users to voice data.
	// Each chunk of voice data is a slice of bytes, store all those slices in a grouped slice.
	var voiceDataPerPlayer = map[string][][]byte{}

	// The file path to an unzipped demo file.
	file, err := os.Open(*demoPath)
	if err != nil {
		log.Fatalf("Failed to open demo file '%s': %v", *demoPath, err)
	}
	defer file.Close()

	parser := dem.NewParser(file)
	var format string

	// Add a parser register for the VoiceData net message.
	parser.RegisterNetMessageHandler(func(m *msgs2.CSVCMsg_VoiceData) {
		// Get the users Steam ID 64.
		steamId := strconv.Itoa(int(m.GetXuid()))
		// Append voice data to map
		format = m.Audio.Format.String()
		voiceDataPerPlayer[steamId] = append(voiceDataPerPlayer[steamId], m.Audio.VoiceData)
	})

	// Parse the full demo file.
	err = parser.ParseToEnd()

	// For each users data, create a wav file containing their voice comms.
	for playerId, voiceData := range voiceDataPerPlayer {
		wavFilePath := fmt.Sprintf("%s.wav", playerId)
		if format == "VOICEDATA_FORMAT_OPUS" {
			err = opusToWav(voiceData, wavFilePath)
			if err != nil {
				fmt.Printf("failed to initialize OpusDecoder: %v\n", err)
				continue
			}

		} else if format == "VOICEDATA_FORMAT_STEAM" {
			convertAudioDataToWavFiles(voiceData, wavFilePath)
		}
	}

	defer parser.Close()
}

// convertAudioDataToWavFiles decodes Steam-format voice data payloads and writes them to a WAV file.
// It uses the Opus decoder for each chunk and encodes the PCM output as a WAV file.
func convertAudioDataToWavFiles(payloads [][]byte, fileName string) {
	// This sample rate can be set using data from the VoiceData net message.
	// But every demo processed has used 24000 and is single channel.
	voiceDecoder, err := decoder.NewOpusDecoder(constants.DefaultSteamSampleRate, constants.DefaultNumChannels)

	if err != nil {
		fmt.Printf("failed to initialize OpusDecoder: %v\n", err)
	}

	o := make([]int, 0, 1024)

	for _, payload := range payloads {
		c, err := decoder.DecodeChunk(payload)

		if err != nil {
			fmt.Printf("failed to initialize OpusDecoder: %v\n", err)
		}

		// Not silent frame
		if c != nil && len(c.Data) > 0 {
			pcm, err := voiceDecoder.Decode(c.Data)

			if err != nil {
				fmt.Printf("failed to initialize OpusDecoder: %v\n", err)
			}

			converted := make([]int, len(pcm))
			for i, v := range pcm {
				// Float32 buffer implementation is wrong in go-audio, so we have to convert to int before encoding
				converted[i] = int(v * constants.IntPCMMaxValue) // no error here, just conversion
			}

			o = append(o, converted...)
		}
	}

	outFile, err := os.Create(fileName)

	if err != nil {
		fmt.Printf("failed to initialize OpusDecoder: %v\n", err)
	}
	defer outFile.Close()

	// Encode new wav file, from decoded opus data.
	enc := wav.NewEncoder(outFile, constants.DefaultSteamSampleRate, constants.DefaultBitDepth, constants.DefaultNumChannels, 1)

	buf := &audio.IntBuffer{
		Data: o,
		Format: &audio.Format{
			SampleRate:  constants.DefaultSteamSampleRate,
			NumChannels: constants.DefaultNumChannels,
		},
	}

	// Write voice data to the file.
	if err := enc.Write(buf); err != nil {
		fmt.Printf("failed to write WAV data: %v\n", err)
	}

	enc.Close()
}

// opusToWav decodes Opus-format voice data and writes the result to a WAV file.
// Returns an error if decoding or file writing fails.
func opusToWav(data [][]byte, wavName string) (err error) {
	opusDecoder, err := decoder.NewDecoder(constants.DefaultOpusSampleRate, constants.DefaultNumChannels)
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
			pp[i] = int(p * constants.IntPCMMaxValue)
		}

		pcmBuffer = append(pcmBuffer, pp...)
	}

	file, err := os.Create(wavName)
	if err != nil {
		return
	}
	defer file.Close()

	enc := wav.NewEncoder(file, constants.DefaultOpusSampleRate, constants.DefaultBitDepth, constants.DefaultNumChannels, 1)
	defer enc.Close()

	buffer := &audio.IntBuffer{
		Data: pcmBuffer,
		Format: &audio.Format{
			SampleRate:  constants.DefaultOpusSampleRate,
			NumChannels: constants.DefaultNumChannels,
		},
	}

	err = enc.Write(buffer)
	if err != nil {
		return
	}

	return
}
