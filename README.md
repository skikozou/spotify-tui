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
- 📋 Queue display with playback support
- 🔊 Active device display

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
- `Tab` - Cycle focus (Sidebar → Main → Queue)
- `Shift+Tab` - Reverse cycle focus

#### Navigation
- `↑/↓` or `j/k` - Move selection
- `Enter` - Select playlist, play track, or play from queue
- `Esc` - Exit search mode

### Layout

```
┌──────────┬──────────────┬──────────┐
│          │              │          │
│ SIDEBAR  │    MAIN      │  QUEUE   │
│  (30%)   │    (40%)     │  (30%)   │
│          │              │          │
│ Playlists│  Track list  │ Up next  │
│          │              │          │
├──────────┼──────────────┼──────────┤
│ USER     │ NOW PLAYING  │ DEVICE   │
│ Name     │ ♫ Track      │ Name     │
│ Plan     │ [████░░]     │ Type     │
│ Followers│ 🔀 ▶ 🔁      │ Volume   │
└──────────┴──────────────┴──────────┘
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
│       ├── delegate.go       # Custom list delegates
│       └── layout.go         # Layout calculations
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
- Volume control (adjustment) not yet implemented

## Future Enhancements

- [ ] Device selection
- [ ] Volume control
- [ ] Lyrics display
- [ ] Album/Artist browsing

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
