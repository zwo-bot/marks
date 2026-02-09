# marks

A bookmark manager for rofi and other launchers. Access and search your browser bookmarks quickly and efficiently.

## Features

- Supports Firefox and Chrome/Chromium bookmarks
- **Linux and macOS** support
- Automatic browser profile detection
- Interactive configuration TUI (`marks configure`)
- Favicon support
- Fast SQLite-based caching (pure Go, no CGO required)
- Clean, single-line display with title and URL
- Cross-compilation support (no CGO dependency)

## Installation

### From source

```bash
go install github.com/zwo-bot/marks@latest
```

### Build from repository

```bash
git clone https://github.com/zwo-bot/marks.git
cd marks
make build.local
```

The binary will be at `build/marks`.

### Cross-compilation

```bash
# Linux
make build.linux.amd64
make build.linux.arm64

# macOS
make build.darwin
make build.darwin.arm64
```

## Usage

### Quick Start

Run the interactive configuration to detect and select your browser profiles:

```bash
marks configure
```

### With rofi (Linux)

```bash
rofi -show bookmarks -show-icons -modi 'bookmarks: marks rofi'
```

### Show bookmarks

```bash
marks show                    # JSON output
marks show -f text            # Text output
marks show -d=false           # Without deduplication
```

### Update bookmark database

```bash
marks update
```

### List available plugins

```bash
marks list-plugins
```

## Interactive Configuration (`marks configure`)

The `marks configure` command provides an interactive TUI to set up your browser profiles:

1. Scans for installed browsers (Firefox, Chrome, Chromium)
2. Shows detected profiles in a list
3. Use **↑/↓** to navigate, **Space** to toggle selection
4. Press **Enter** to save, **q** to quit
5. Config is saved to the appropriate location for your OS

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| ↑/↓ or k/j | Navigate |
| Space | Toggle selection |
| a | Select all |
| n | Deselect all |
| Enter | Save configuration |
| q / Ctrl+C | Quit |

## Configuration

### Configuration File Location

The program looks for the configuration file in the following order:
1. Custom path specified with `--config` flag
2. Current directory (`./config.json`)
3. System config directory:
   - **Linux:** `~/.config/marks/config.json`
   - **macOS:** `~/Library/Application Support/marks/config.json`

### Data Storage

The bookmark database is stored at:
- **Linux:** `~/.local/share/marks/bookmarks.db`
- **macOS:** `~/Library/Application Support/marks/bookmarks.db`

Favicon cache is stored at:
- **Linux:** `~/.cache/marks-favicons/`
- **macOS:** `~/Library/Caches/marks-favicons/`

### Example Configuration

```json
{
    "plugins": {
        "firefox": {
            "profile_path": "~/.mozilla/firefox/xxxxxxxx.default-release"
        },
        "chrome": {
            "profile_path": "~/.config/google-chrome/Default/Bookmarks"
        }
    },
    "defaultBrowser": "firefox"
}
```

### Finding Your Profile Path

#### Firefox
1. Open Firefox and navigate to `about:profiles`
2. Look for the profile marked as "Default"
3. Copy the "Root Directory" path

**Default locations:**
- **Linux:** `~/.mozilla/firefox/<profile>/`
- **Linux (Snap):** `~/snap/firefox/common/.mozilla/firefox/<profile>/`
- **Linux (Flatpak):** `~/.var/app/org.mozilla.firefox/.mozilla/firefox/<profile>/`
- **macOS:** `~/Library/Application Support/Firefox/Profiles/<profile>/`

#### Chrome / Chromium

**Default locations:**
- **Linux (Chrome):** `~/.config/google-chrome/Default/Bookmarks`
- **Linux (Chromium):** `~/.config/chromium/Default/Bookmarks`
- **macOS (Chrome):** `~/Library/Application Support/Google/Chrome/Default/Bookmarks`
- **macOS (Chromium):** `~/Library/Application Support/Chromium/Default/Bookmarks`

Note: The application will attempt to automatically find these paths. Use `marks configure` for the easiest setup, or manually configure if auto-detection doesn't work for your setup.

## Building

```bash
make build.local     # Local platform
make build.linux     # Linux amd64
make build.darwin    # macOS amd64
```

## License

See [LICENSE](LICENSE) for details.
