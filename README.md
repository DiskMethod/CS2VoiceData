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

---

## Acknowledgements

- Forked and extended from [DandrewsDev/CS2VoiceData](https://github.com/DandrewsDev/CS2VoiceData).
- Thanks to [@DandrewsDev](https://github.com/DandrewsDev), [@rumblefrog](https://github.com/rumblefrog), [@markus-wa](https://github.com/markus-wa), and all contributors to the original project and libraries.
- Special thanks to [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) and [Reversing Steam Voice Codec blog post](https://zhenyangli.me/posts/reversing-steam-voice-codec/).

---
