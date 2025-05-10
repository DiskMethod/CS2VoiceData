# cs2voice-extract

Extracts per-player voice data from CS2 demo files and saves each player's audio as a separate WAV file.

---

## Features

- Parses CS2 demo files (`.dem`) and extracts voice chat audio.
- Outputs one WAV file per player found in the demo.
- Designed for modular use in larger CS2 voice data processing pipelines.

---

## Usage

```sh
go run ./cmd/cs2voice-extract <path-to-demo-file>
```
Or, after building:
```sh
go build -o cs2voice-extract ./cmd/cs2voice-extract
./cs2voice-extract <path-to-demo-file>
```

**Example:**
```sh
./cs2voice-extract /path/to/match.dem
```

WAV files will be created in the current directory, named by player Steam ID.

---

## Options

- `<path-to-demo-file>`: Path to an unzipped CS2 demo file (`.dem`).

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

## License

MIT

---

## See Also

- [Project root README](../../README.md)
