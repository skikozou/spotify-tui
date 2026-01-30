# Spotify TUI

A Terminal User Interface (TUI) client for Spotify written in Go using Bubbletea.

## Features

- 🎵 Browse and play your playlists
- 💚 Liked Songs support
- 🔍 Track search functionality
- ⏯️ Full playback control (Play/Pause, Next, Previous)
- 🔀 Shuffle and repeat modes (synced with Spotify)
- 📊 Real-time progress bar with smooth updates
- 🎨 Clean, Spotify-themed interface
- ⌨️ Keyboard-driven navigation
- 👤 User profile display
- ♫ Now playing indicator with playlist/album name

## Requirements

- Go 1.22 or higher
- Spotify Premium account (required for playback control)
- Spotify Developer credentials

## Setup

### 1. Get Spotify API Credentials

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new application
3. Set the Redirect URI to: `http://localhost:8080/callback`
4. Copy your Client ID and Client Secret

### 2. Install and Run

```bash
# Clone the repository
git clone https://github.com/yourusername/spotify-tui
cd spotify-tui

# Install dependencies
go mod download

# Build
go build -o spotify-tui ./cmd/spotify-tui

# Run
./spotify-tui
```

On first run, you'll be prompted to enter your Client ID and Client Secret. These will be saved to `~/.config/spotify-tui/config.json`.

## Usage

### Keybindings

#### Global
- `q` - Quit
- `Space` - Play/Pause
- `n` - Next track
- `p` - Previous track
- `s` - Toggle shuffle
- `r` - Cycle repeat mode (off → context → track)
- `/` - Search mode
- `Tab` - Switch between sidebar and main panel

#### Navigation
- `↑/↓` or `j/k` - Move selection
- `Enter` - Select playlist or play track
- `Esc` - Exit search mode

### Layout

```
┌──────────┬─────────────────────────────────────────┐
│          │                                         │
│ LEFT     │           RIGHT PANEL                   │
│ SIDEBAR  │           (Track list)                  │
│ (30%)    │              (70%)                      │
│          │  Song Title                             │
│ Playlists│  Artist Name (gray)                    │
│          │                                         │
├──────────┼─────────────────────────────────────────┤
│ USER     │ NOW PLAYING BAR                         │
│ Name     │ ♫ Track - Artist | Playlist Name        │
│ Plan     │ [████████░░] 2:34 / 4:12               │
│ Followers│ 🔀 ▶ 🔁   [Space] Pause | [n] Next     │
└──────────┴─────────────────────────────────────────┘
```

## Architecture

```
spotify-tui/
├── cmd/
│   └── spotify-tui/
│       └── main.go           # Entry point
├── internal/
│   ├── auth/
│   │   └── auth.go           # OAuth authentication
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── spotify/
│   │   └── client.go         # Spotify API wrapper
│   └── ui/
│       ├── model.go          # Bubbletea model
│       ├── update.go         # Update logic
│       ├── view.go           # View rendering
│       └── delegate.go       # Custom list delegate
├── go.mod
└── README.md
```

## Technology Stack

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [zmb3/spotify](https://github.com/zmb3/spotify) - Spotify Web API client

## Limitations

- Requires Spotify Premium for playback control
- Device switching is not yet available
- Volume control not yet implemented

## Future Enhancements

- [ ] Device selection
- [ ] Volume control
- [ ] Queue management
- [ ] Lyrics display
- [ ] Album/Artist browsing

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
