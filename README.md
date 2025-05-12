# cs2-voice-tools

A modular suite of CLI tools for extracting, transcribing, and analyzing player voice data from CS2 demo files. (Compression is now handled directly by the extraction tool.)

**Forked from:** [DandrewsDev/CS2VoiceData](https://github.com/DandrewsDev/CS2VoiceData)

---

## Features

- Modular CLI tools for each stage of CS2 voice data processing:
  - Extraction (`cs2voice extract`): Extracts per-player voice data from CS2 demos with support for direct output to various formats (WAV, FLAC, etc.) and user-selectable audio quality (sample rate, bit depth, etc.)
  - Transcription (`cs2voice transcribe` - planned)
  - Analysis (`cs2voice analyze` - planned)
  - Unified pipeline (`cs2voice pipeline` - planned)

---

## Installation

- Requires Go 1.23+ and dependencies listed in `go.mod`.
- Some tools may require additional system libraries (e.g., libopus, ffmpeg). See each tool's README for details.

## Usage

### Global Flags

All commands support the following global flags:

- `-v, --verbose`: Enable verbose logging (shows additional debug information)
- `-o, --output-dir`: Directory to save output files (default: current directory)
- `-f, --force`: Force overwrite existing files (default: skip existing files)

### Extract Command Flags

The `extract` command supports these additional flags:

- `-p, --players`: Filter to specific players by SteamID64 (comma-separated list)
- `-t, --format`: Output audio format (wav, mp3, ogg, flac, aac, m4a - default: wav)

> **Note**: Using formats other than WAV requires ffmpeg to be installed on your system

Examples:

```bash
# Run extraction with verbose logging
cs2voice extract --verbose my-demo.dem

# Save extracted files to a specific directory
cs2voice extract --output-dir /path/to/output my-demo.dem

# Short forms work too
cs2voice extract -v -o /path/to/output my-demo.dem

# Directory will be created if it doesn't exist
cs2voice extract -o ./new-directory my-demo.dem

# Force overwrite existing files
cs2voice extract --force my-demo.dem

# Extract voice for specific players only
cs2voice extract --players 76561198123456789,76561198987654321 my-demo.dem

# Extract voice in MP3 format
cs2voice extract --format mp3 my-demo.dem

# Extract voice in FLAC format (lossless compression)
cs2voice extract --format flac my-demo.dem

# Combine multiple flags
cs2voice extract -v -o ./output -f -p 76561198123456789 -t mp3 my-demo.dem
```

---

## Acknowledgements

- Forked and extended from [DandrewsDev/CS2VoiceData](https://github.com/DandrewsDev/CS2VoiceData).
- Thanks to [@DandrewsDev](https://github.com/DandrewsDev), [@rumblefrog](https://github.com/rumblefrog), [@markus-wa](https://github.com/markus-wa), and all contributors to the original project and libraries.
- Special thanks to [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) and [Reversing Steam Voice Codec blog post](https://zhenyangli.me/posts/reversing-steam-voice-codec/).

---
