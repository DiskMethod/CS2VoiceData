// Package decoder provides utilities for decoding Opus-encoded audio streams from Steam and CS2 voice data.
package decoder

import (
	"bytes"
	"encoding/binary"

	"gopkg.in/hraban/opus.v2"
)

const (
	// FrameSize is the number of samples per Opus frame for decoding.
	FrameSize = 480
)

// OpusDecoder wraps an opus.Decoder and tracks the current frame for audio processing.
type OpusDecoder struct {
	decoder *opus.Decoder

	currentFrame uint16
}

// NewOpusDecoder creates a new OpusDecoder with the specified sample rate and channel count.
func NewOpusDecoder(sampleRate, channels int) (*OpusDecoder, error) {
	decoder, err := opus.NewDecoder(sampleRate, channels)

	if err != nil {
		return nil, err
	}

	return &OpusDecoder{
		decoder:      decoder,
		currentFrame: 0,
	}, nil
}

// Decode decodes a slice of Opus-encoded bytes into PCM float32 samples.
func (d *OpusDecoder) Decode(b []byte) ([]float32, error) {
	buf := bytes.NewBuffer(b)

	output := make([]float32, 0, 1024)

	for buf.Len() != 0 {
		var chunkLen int16
		if err := binary.Read(buf, binary.LittleEndian, &chunkLen); err != nil {
			return nil, err
		}

		if chunkLen == -1 {
			d.currentFrame = 0
			break
		}

		var currentFrame uint16
		if err := binary.Read(buf, binary.LittleEndian, &currentFrame); err != nil {
			return nil, err
		}

		previousFrame := d.currentFrame

		chunk := make([]byte, chunkLen)
		n, err := buf.Read(chunk)
		if err != nil {
			return nil, err
		}

		if n != int(chunkLen) {
			return nil, ErrInvalidVoicePacket
		}

		if currentFrame >= previousFrame {
			if currentFrame == previousFrame {
				d.currentFrame = currentFrame + 1

				decoded, err := d.decodeSteamChunk(chunk)

				if err != nil {
					return nil, err
				}

				output = append(output, decoded...)
			} else {
				decoded, err := d.decodeLoss(currentFrame - previousFrame)

				if err != nil {
					return nil, err
				}

				output = append(output, decoded...)
			}
		}
	}

	return output, nil
}

func (d *OpusDecoder) decodeSteamChunk(b []byte) ([]float32, error) {
	o := make([]float32, FrameSize)

	n, err := d.decoder.DecodeFloat32(b, o)

	if err != nil {
		return nil, err
	}

	return o[:n], nil
}

func (d *OpusDecoder) decodeLoss(samples uint16) ([]float32, error) {
	loss := min(samples, 10)

	o := make([]float32, 0, FrameSize*loss)

	for i := 0; i < int(loss); i += 1 {
		t := make([]float32, FrameSize)

		if err := d.decoder.DecodePLCFloat32(t); err != nil {
			return nil, err
		}

		o = append(o, t...)
	}

	return o, nil
}

// NewDecoder returns a new opus.Decoder for the given sample rate and channel count.
func NewDecoder(sampleRate, channels int) (*opus.Decoder, error) {
	decoder, err := opus.NewDecoder(sampleRate, channels)
	return decoder, err
}

// Decode decodes Opus-encoded data using the provided opus.Decoder and returns PCM float32 samples.
func Decode(decoder *opus.Decoder, data []byte) ([]float32, error) {
	pcm := make([]float32, 1024)

	nlen, err := decoder.DecodeFloat32(data, pcm)
	if err != nil {
		return nil, err
	}

	return pcm[:nlen], nil
}
