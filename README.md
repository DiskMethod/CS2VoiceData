# CS2VoiceData Modular CLI Suite

A modular suite of CLI tools for extracting, compressing, transcribing, and analyzing player voice data from CS2 demo files.

**Forked from:** [DandrewsDev/CS2VoiceData](https://github.com/DandrewsDev/CS2VoiceData)

---

## Features

- Modular CLI tools for each stage of CS2 voice data processing:
  - Extraction (`cs2voice-extract`)
  - Compression (`cs2voice-compress` - planned)
  - Transcription (`cs2voice-transcribe` - planned)
  - Analysis (`cs2voice-analyze` - planned)
  - Unified pipeline (`cs2voice-pipeline` - planned)

---

## Installation

- Requires Go 1.23+ and dependencies listed in `go.mod`.
- Some tools may require additional system libraries (e.g., libopus, ffmpeg). See each tool's README for details.

---

## Documentation

**Usage instructions and CLI documentation for each tool are located in their respective directories under `cmd/`.**

- Example: See [`cmd/cs2voice-extract/README.md`](cmd/cs2voice-extract/README.md) for extraction tool usage.

---

## Acknowledgements

- Forked and extended from [DandrewsDev/CS2VoiceData](https://github.com/DandrewsDev/CS2VoiceData).
- Thanks to [@DandrewsDev](https://github.com/DandrewsDev), [@rumblefrog](https://github.com/rumblefrog), [@markus-wa](https://github.com/markus-wa), and all contributors to the original project and libraries.
- Special thanks to [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) and [Reversing Steam Voice Codec blog post](https://zhenyangli.me/posts/reversing-steam-voice-codec/).

---

## License

MIT
