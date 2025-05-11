// Package decoder provides utilities for parsing and decoding Steam and CS2 voice data packets.
package decoder

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

// minimumLength is the smallest possible size of a valid voice data packet.
// This is based on the observed structure from reverse engineering the Steam voice codec.
// See: https://zhenyangli.me/posts/reversing-steam-voice-codec/
const (
	minimumLength = 18

	// PayloadTypeHeader is the expected value for the payload type byte that indicates Steam voice packet header
	PayloadTypeHeader = 0x0B

	// VoiceTypeOpusPLC is the value for the voiceType byte indicating Opus PLC encoded voice data
	VoiceTypeOpusPLC = 0x06

	// VoiceTypeSilence is the value for the voiceType byte indicating silence
	VoiceTypeSilence = 0x00
)

var (
	// ErrInsufficientData is returned when there is not enough data to parse a chunk.
	ErrInsufficientData   = errors.New("insufficient amount of data to chunk")
	// ErrInvalidVoicePacket is returned when a voice packet does not match the expected format.
	ErrInvalidVoicePacket = errors.New("invalid voice packet")
	// ErrMismatchChecksum is returned when a packet's checksum does not match the computed value.
	ErrMismatchChecksum   = errors.New("mismatching voice data checksum")
)

// Chunk represents a parsed voice data packet from a CS2 demo file.
type Chunk struct {
	SteamID    uint64
	SampleRate uint16
	Length     uint16
	Data       []byte
	Checksum   uint32
}

// DecodeChunk parses a raw voice data packet from a CS2 demo file.
//
// Packet structure (see blog for details):
// [u64 steamID][u8 payloadType=0x0B][u16 sampleRate][u8 voiceType][u16 length][voice data][u32 crc32]
// - steamID: Little-endian 64-bit Steam ID of the player
// - payloadType: Always 0x0B for Steam voice packets (see PayloadTypeHeader)
// - sampleRate: Audio sample rate (typically 24000, see reverse engineering)
// - voiceType: 0x06 for Opus PLC data, 0x00 for silence
// - length: Length of the following voice data
// - voice data: Opus PLC encoded data (if voiceType==0x06)
// - crc32: CRC32 checksum of all previous bytes
//
// For more details, see: https://zhenyangli.me/posts/reversing-steam-voice-codec/
// DecodeChunk parses a raw voice data packet from a CS2 demo file and returns a Chunk.
// Returns an error if the packet is invalid, incomplete, or fails checksum verification.
func DecodeChunk(b []byte) (*Chunk, error) {
	bLen := len(b)

	if bLen < minimumLength {
		return nil, fmt.Errorf("%w (received: %d bytes, expected at least %d bytes)", ErrInsufficientData, bLen, minimumLength)
	}

	chunk := &Chunk{}

	buf := bytes.NewBuffer(b)

	if err := binary.Read(buf, binary.LittleEndian, &chunk.SteamID); err != nil {
		return nil, err
	}

	var payloadType byte
	if err := binary.Read(buf, binary.LittleEndian, &payloadType); err != nil {
		return nil, err
	}

	// PayloadTypeHeader (0x0B) is always expected for Steam voice packets
	if payloadType != PayloadTypeHeader {
		return nil, fmt.Errorf("%w (received %x, expected %x)", ErrInvalidVoicePacket, payloadType, PayloadTypeHeader)
	}

	if err := binary.Read(buf, binary.LittleEndian, &chunk.SampleRate); err != nil {
		return nil, err
	}

	var voiceType byte
	if err := binary.Read(buf, binary.LittleEndian, &voiceType); err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.LittleEndian, &chunk.Length); err != nil {
		return nil, err
	}

	switch voiceType {
	case VoiceTypeOpusPLC:
		// Opus PLC encoded voice data
		remaining := buf.Len()
		chunkLen := int(chunk.Length)

		if remaining < chunkLen {
			return nil, fmt.Errorf("%w (received: %d bytes, expected at least %d bytes)", ErrInsufficientData, bLen, (bLen + (chunkLen - remaining)))
		}

		data := make([]byte, chunkLen)
		n, err := buf.Read(data)

		if err != nil {
			return nil, err
		}

		if n != chunkLen {
			return nil, fmt.Errorf("%w (expected to read %d bytes, but read %d bytes)", ErrInsufficientData, chunkLen, n)
		}

		chunk.Data = data
	case VoiceTypeSilence:
		// Silence frame (no data)
		// The length field is the number of silence frames
		// chunk.Data remains empty
	default:
		return nil, fmt.Errorf("%w (expected 0x6 or 0x0 voice data, received %x)", ErrInvalidVoicePacket, voiceType)
	}

	remaining := buf.Len()

	if remaining != 4 {
		return nil, fmt.Errorf("%w (has %d bytes remaining, expected 4 bytes remaining)", ErrInvalidVoicePacket, remaining)
	}

	if err := binary.Read(buf, binary.LittleEndian, &chunk.Checksum); err != nil {
		return nil, err
	}

	actualChecksum := crc32.ChecksumIEEE(b[0 : bLen-4])

	if chunk.Checksum != actualChecksum {
		return nil, fmt.Errorf("%w (received %x, expected %x)", ErrMismatchChecksum, chunk.Checksum, actualChecksum)
	}

	return chunk, nil
}
