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
	FocusQueue
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
	queue     []spotifysdk.FullTrack
	queueList list.Model

	// Devices
	devices      []spotifysdk.PlayerDevice
	activeDevice *spotifysdk.PlayerDevice
	volume       int

	// User
	user *spotifysdk.PrivateUser

	// Autoplay
	autoplayEnabled          bool
	lastRecommendationTime   time.Time
	recommendationInProgress bool
	recentlyQueuedTracks     map[string]bool
	recentlyQueuedList       []string
	isInitialAutoplay        bool // 単曲再生時の初回Autoplayかどうか

	// Polling intervals
	lastQueueFetch  time.Time
	lastDeviceFetch time.Time

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
type recommendationsMsg struct {
	tracks []spotifysdk.SimpleTrack
	err    error
}
type queueTrackResultMsg struct {
	trackID string
	err     error
}

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

	queueDelegate := NewQueueDelegate()
	queueList := list.New([]list.Item{}, queueDelegate, 0, 0)
	queueList.SetShowHelp(false)
	queueList.SetFilteringEnabled(false)
	queueList.SetShowStatusBar(false)
	queueList.SetShowTitle(false)

	return Model{
		ctx:                  ctx,
		client:               client,
		focus:                FocusSidebar,
		playlists:            playlistList,
		trackList:            trackList,
		queueList:            queueList,
		lastUpdate:           time.Now(),
		repeatState:          "off",
		recentlyQueuedTracks: make(map[string]bool),
		recentlyQueuedList:   make([]string, 0),
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

// singleTrackPlayedMsg は単曲再生が完了したことを通知するメッセージ
type singleTrackPlayedMsg struct {
	trackID spotifysdk.ID
}

func (m Model) playTrackAlone(uri spotifysdk.URI) tea.Cmd {
	// URIからIDを抽出 (spotify:track:XXXXX -> XXXXX)
	trackID := spotifysdk.ID(uri[len("spotify:track:"):])
	return func() tea.Msg {
		if err := m.client.PlayTrackAlone(m.ctx, uri); err != nil {
			return errorMsg(err.Error())
		}
		return singleTrackPlayedMsg{trackID: trackID}
	}
}

func (m Model) skipToQueueIndex(index int) tea.Cmd {
	return func() tea.Msg {
		// index 0 = 現在再生中の次の曲なので、index+1回スキップする
		skipCount := index + 1
		if err := m.client.SkipToNth(m.ctx, skipCount); err != nil {
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

// Autoplay 関連メソッド

const (
	autoplayQueueThreshold   = 2                // キュー閾値（補充時）
	autoplayCooldown         = 30 * time.Second // クールダウン（補充時）
	autoplayBatchSize        = 3                // 補充時に追加するトラック数
	autoplayInitialBatchSize = 10               // 単曲再生時に追加するトラック数
	maxRecentlyQueuedSize    = 50               // 重複防止用履歴サイズ
)

// shouldTriggerAutoplay は Autoplay をトリガーすべきか判定する
func (m *Model) shouldTriggerAutoplay() bool {
	if !m.autoplayEnabled {
		return false
	}
	if m.recommendationInProgress {
		return false
	}
	if m.currentTrack == nil || m.currentTrack.Item == nil {
		return false
	}
	if len(m.queue) > autoplayQueueThreshold {
		return false
	}
	if time.Since(m.lastRecommendationTime) < autoplayCooldown {
		return false
	}
	return true
}

// buildSeeds はレコメンデーション用のシードを構築する
func (m *Model) buildSeeds() spotifysdk.Seeds {
	seeds := spotifysdk.Seeds{}

	if m.currentTrack != nil && m.currentTrack.Item != nil {
		// 現在のトラック
		seeds.Tracks = append(seeds.Tracks, m.currentTrack.Item.ID)

		// 現在のアーティスト（最大2）
		for i, artist := range m.currentTrack.Item.Artists {
			if i >= 2 {
				break
			}
			seeds.Artists = append(seeds.Artists, artist.ID)
		}
	}

	// キュー内のトラック（合計5を超えない範囲）
	remaining := 5 - len(seeds.Tracks) - len(seeds.Artists)
	for i, track := range m.queue {
		if i >= remaining || i >= 2 {
			break
		}
		seeds.Tracks = append(seeds.Tracks, track.ID)
	}

	return seeds
}

// isRecentlyQueued はトラックが最近キューに追加されたか確認する
func (m *Model) isRecentlyQueued(trackID string) bool {
	return m.recentlyQueuedTracks[trackID]
}

// markAsQueued はトラックをキュー追加済みとしてマークする
func (m *Model) markAsQueued(trackID string) {
	if m.recentlyQueuedTracks[trackID] {
		return
	}

	// 最大サイズを超えたら古いものを削除
	if len(m.recentlyQueuedList) >= maxRecentlyQueuedSize {
		oldest := m.recentlyQueuedList[0]
		delete(m.recentlyQueuedTracks, oldest)
		m.recentlyQueuedList = m.recentlyQueuedList[1:]
	}

	m.recentlyQueuedTracks[trackID] = true
	m.recentlyQueuedList = append(m.recentlyQueuedList, trackID)
}

// filterRecommendations は重複を除去したレコメンデーションを返す
func (m *Model) filterRecommendations(tracks []spotifysdk.SimpleTrack) []spotifysdk.SimpleTrack {
	filtered := make([]spotifysdk.SimpleTrack, 0)
	for _, track := range tracks {
		if !m.isRecentlyQueued(string(track.ID)) {
			filtered = append(filtered, track)
		}
	}
	return filtered
}

// fetchRecommendations はレコメンデーションを取得するコマンドを返す
func (m Model) fetchRecommendations() tea.Cmd {
	seeds := m.buildSeeds()
	return func() tea.Msg {
		recs, err := m.client.GetRecommendations(m.ctx, seeds, 10)
		if err != nil {
			return recommendationsMsg{err: err}
		}
		return recommendationsMsg{tracks: recs.Tracks}
	}
}

// queueTrack はトラックをキューに追加するコマンドを返す
func (m Model) queueTrack(trackID spotifysdk.ID) tea.Cmd {
	return func() tea.Msg {
		err := m.client.QueueTrack(m.ctx, trackID)
		return queueTrackResultMsg{trackID: string(trackID), err: err}
	}
}

// fetchRecommendationsForTrack は指定トラックに基づいてレコメンデーションを取得する
func (m Model) fetchRecommendationsForTrack(trackID spotifysdk.ID) tea.Cmd {
	seeds := spotifysdk.Seeds{
		Tracks: []spotifysdk.ID{trackID},
	}
	return func() tea.Msg {
		recs, err := m.client.GetRecommendations(m.ctx, seeds, 20)
		if err != nil {
			return recommendationsMsg{err: err}
		}
		return recommendationsMsg{tracks: recs.Tracks}
	}
}
