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
	tracks              []spotifysdk.PlaylistTrack
	trackList           list.Model
	currentPlaylistURI  spotifysdk.URI
	currentPlaylistName string
	playingPlaylistName string
	isLikedSongs        bool
	loadingTracks       bool
	searchMode          bool
	searchQuery         string
	searchResults       []spotifysdk.FullTrack
	searchIndex         int

	// Player State
	currentTrack    *spotifysdk.PlayerState
	playingTrackURI string
	isPlaying       bool
	progress        time.Duration
	duration        time.Duration
	lastUpdate      time.Time
	shuffle         bool
	repeatState     string

	// Queue
	queue []spotifysdk.FullTrack

	// Devices
	devices      []spotifysdk.PlayerDevice
	activeDevice *spotifysdk.PlayerDevice
	volume       int

	// User
	user *spotifysdk.PrivateUser

	// Error
	err string
}

type tickMsg time.Time
type playbackMsg *spotifysdk.PlayerState
type playlistsMsg []spotifysdk.SimplePlaylist
type tracksMsg struct {
	tracks      []spotifysdk.PlaylistTrack
	playlistURI spotifysdk.URI
}
type savedTracksMsg []spotifysdk.SavedTrack
type searchResultsMsg []spotifysdk.FullTrack
type userMsg *spotifysdk.PrivateUser
type queueMsg *spotifysdk.Queue
type devicesMsg []spotifysdk.PlayerDevice
type errorMsg string

func NewModel(ctx context.Context, client *spotify.Client) Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0) // アイテム間のスペースを0に

	playlistList := list.New([]list.Item{}, delegate, 0, 0)
	playlistList.SetShowHelp(false)
	playlistList.SetFilteringEnabled(false)
	playlistList.SetShowStatusBar(false)
	playlistList.SetShowTitle(false)

	trackDelegate := NewTrackDelegate()
	trackList := list.New([]list.Item{}, trackDelegate, 0, 0)
	trackList.SetShowHelp(false)
	trackList.SetFilteringEnabled(false)
	trackList.SetShowStatusBar(false)
	trackList.SetShowTitle(false)

	return Model{
		ctx:         ctx,
		client:      client,
		focus:       FocusSidebar,
		playlists:   playlistList,
		trackList:   trackList,
		lastUpdate:  time.Now(),
		repeatState: "off",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchPlaylists(),
		m.fetchUser(),
		m.fetchQueue(),
		m.fetchDevices(),
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
		state, err := m.client.PlayerState(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return playbackMsg(state)
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

func (m Model) fetchSavedTracks() tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.client.SavedTracks(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return savedTracksMsg(tracks)
	}
}

type playStartedMsg string

func (m Model) playTrackInPlaylist(offset int) tea.Cmd {
	playlistName := m.currentPlaylistName
	return func() tea.Msg {
		// Liked Songsの場合はURIリストで再生
		if m.isLikedSongs {
			uris := make([]spotifysdk.URI, len(m.tracks))
			for i, track := range m.tracks {
				uris[i] = track.Track.URI
			}
			if err := m.client.PlayTrackFromURIList(m.ctx, uris, offset); err != nil {
				return errorMsg(err.Error())
			}
			return playStartedMsg(playlistName)
		}

		// 通常のプレイリストはコンテキストで再生
		if m.currentPlaylistURI == "" {
			return errorMsg("No playlist context")
		}
		if err := m.client.PlayTrackInContext(m.ctx, m.currentPlaylistURI, offset); err != nil {
			return errorMsg(err.Error())
		}
		return playStartedMsg(playlistName)
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

func (m Model) fetchUser() tea.Cmd {
	return func() tea.Msg {
		user, err := m.client.CurrentUser(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return userMsg(user)
	}
}

func (m Model) fetchQueue() tea.Cmd {
	return func() tea.Msg {
		queue, err := m.client.GetQueue(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return queueMsg(queue)
	}
}

func (m Model) fetchDevices() tea.Cmd {
	return func() tea.Msg {
		devices, err := m.client.PlayerDevices(m.ctx)
		if err != nil {
			return errorMsg(err.Error())
		}
		return devicesMsg(devices)
	}
}
