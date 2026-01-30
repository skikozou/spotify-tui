package ui

import (
	"context"
	"time"

	"spotify-tui/internal/spotify"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	spotifysdk "github.com/zmb3/spotify/v2"
)

type FocusPanel int

const (
	FocusSidebar FocusPanel = iota
	FocusMain
)

type Model struct {
	ctx    context.Context
	client *spotify.Client

	// UI State
	width  int
	height int
	focus  FocusPanel

	// Sidebar
	playlists     list.Model
	selectedIndex int

	// Main Panel
	tracks             []spotifysdk.PlaylistTrack
	trackIndex         int
	currentPlaylistURI spotifysdk.URI
	searchMode         bool
	searchQuery        string
	searchResults      []spotifysdk.FullTrack
	searchIndex        int

	// Player State
	currentTrack *spotifysdk.CurrentlyPlaying
	isPlaying    bool
	progress     time.Duration
	duration     time.Duration
	lastUpdate   time.Time
	shuffle      bool
	repeatState  string

	// Error
	err string
}

type tickMsg time.Time
type playbackMsg *spotifysdk.CurrentlyPlaying
type playlistsMsg []spotifysdk.SimplePlaylist
type tracksMsg struct {
	tracks      []spotifysdk.PlaylistTrack
	playlistURI spotifysdk.URI
}
type searchResultsMsg []spotifysdk.FullTrack
type errorMsg string

func NewModel(ctx context.Context, client *spotify.Client) Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0) // アイテム間のスペースを0に

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)

	return Model{
		ctx:         ctx,
		client:      client,
		focus:       FocusSidebar,
		playlists:   l,
		lastUpdate:  time.Now(),
		repeatState: "off",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchPlaylists(),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) fetchPlaylists() tea.Cmd {
	return func() tea.Msg {
		playlists, err := m.client.UserPlaylists(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return playlistsMsg(playlists)
	}
}

func (m Model) fetchCurrentPlayback() tea.Cmd {
	return func() tea.Msg {
		playing, err := m.client.CurrentlyPlaying(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return playbackMsg(playing)
	}
}

func (m Model) fetchPlaylistTracks(playlistID spotifysdk.ID) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.client.PlaylistTracks(m.ctx, playlistID)
		if err != nil {
			return errorMsg(err.Error())
		}
		// プレイリストのURIを構築
		playlistURI := spotifysdk.URI("spotify:playlist:" + string(playlistID))
		return tracksMsg{
			tracks:      tracks,
			playlistURI: playlistURI,
		}
	}
}

func (m Model) playTrackInPlaylist(offset int) tea.Cmd {
	return func() tea.Msg {
		if m.currentPlaylistURI == "" {
			return errorMsg("No playlist context")
		}
		if err := m.client.PlayTrackInContext(m.ctx, m.currentPlaylistURI, offset); err != nil {
			return errorMsg(err.Error())
		}
		return nil
	}
}

func (m Model) playTrackAlone(uri spotifysdk.URI) tea.Cmd {
	return func() tea.Msg {
		if err := m.client.PlayTrackAlone(m.ctx, uri); err != nil {
			return errorMsg(err.Error())
		}
		return nil
	}
}

func (m Model) togglePlayPause() tea.Cmd {
	return func() tea.Msg {
		var err error
		if m.isPlaying {
			err = m.client.Pause(m.ctx)
		} else {
			err = m.client.Play(m.ctx)
		}
		if err != nil {
			return errorMsg(err.Error())
		}
		return nil
	}
}

func (m Model) nextTrack() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.Next(m.ctx); err != nil {
			return errorMsg(err.Error())
		}
		return nil
	}
}

func (m Model) previousTrack() tea.Cmd {
	return func() tea.Msg {
		if err := m.client.Previous(m.ctx); err != nil {
			return errorMsg(err.Error())
		}
		return nil
	}
}

func (m Model) performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		if query == "" {
			return searchResultsMsg([]spotifysdk.FullTrack{})
		}

		results, err := m.client.Search(m.ctx, query)
		if err != nil {
			return errorMsg(err.Error())
		}
		return searchResultsMsg(results)
	}
}
