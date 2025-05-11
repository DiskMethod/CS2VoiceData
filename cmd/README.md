# cs2voice

`cs2voice` is a single binary that provides multiple sub-commands for working with CS2 demo voice data.

## Current Sub-commands

- `extract` – Extract per-player voice data to audio files.

---

## Usage

```sh
cs2voice extract [options] <path-to-demo-file>
```

Or build from source:

```sh
go build
./cs2voice extract [options] <path-to-demo-file>
```

### Examples

```sh
# Extract all voice as FLAC to a custom directory
cs2voice extract -o ./voices --format flac /path/to/match.dem

# Extract only for a specific player, as 16-bit WAV
cs2voice extract --player 76561198000000000 --bit-depth 16 /path/to/match.dem
```

Audio files will be created in the specified output directory (default: current directory), named by player Steam ID.

---

## Options

- `<path-to-demo-file>`: Path to an unzipped CS2 demo file (`.dem`).
- `-o, --output-dir <dir>`: Output directory for extracted audio files (default: current directory)
- `--format <wav|flac|mp3|ogg>`: Output audio format (default: wav)
- `--sample-rate <rate>`: Set audio sample rate (default: 48000)
- `--bit-depth <bits>`: Set audio bit depth (default: 32)
- `--player <steamid>`: Extract voice for specific player(s) only (can be repeated)
- `--force`: Overwrite existing files
- `-v, --verbose`: Enable verbose logging
- `--help`: Show help message

---

## Requirements

- Go 1.23+
- System dependencies:
  - `libopus` (for Opus voice decoding)
- Go dependencies are listed in the root `go.mod`.

> **Note:**  
> This tool requires cgo. Make sure `CGO_ENABLED=1` is set in your environment when building or running.  
> If you use VS Code (with the Go extension), set `"go.toolsEnvVars": { "CGO_ENABLED": "1" }` in your workspace or user settings to avoid IDE errors and ensure proper code navigation.

**Install Opus dependencies:**

- **Ubuntu/Debian:**
  ```sh
  sudo apt-get install pkg-config libopus-dev libopusfile-dev
  ```
- **macOS (Homebrew):**
  ```sh
  brew install pkg-config opus opusfile
  ```

---

## Troubleshooting

- **No WAV files produced:**
  - Make sure your demo is not a Valve Matchmaking demo (these do not contain voice data).
  - Check that your system has the required Opus libraries installed.

- **Decoder errors:**
  - Ensure your demo file is unzipped and not corrupted.

---

## See Also

- [Project root README](../README.md)
