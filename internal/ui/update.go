package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	spotifysdk "github.com/zmb3/spotify/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// searchMode中は特別処理
		if m.searchMode {
			switch key {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.searchResults = nil
				m.searchIndex = 0
				return m, nil
			case "enter":
				// 検索結果がある場合は再生、ない場合は検索実行
				if len(m.searchResults) > 0 {
					track := m.searchResults[m.searchIndex]
					return m, m.playTrackAlone(track.URI)
				} else if m.searchQuery != "" {
					return m, m.performSearch(m.searchQuery)
				}
				return m, nil
			case "backspace", "ctrl+h":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
				return m, nil
			case "up":
				if len(m.searchResults) > 0 && m.searchIndex > 0 {
					m.searchIndex--
				}
				return m, nil
			case "down":
				if len(m.searchResults) > 0 && m.searchIndex < len(m.searchResults)-1 {
					m.searchIndex++
				}
				return m, nil
			default:
				// 通常の文字を追加
				if len(key) == 1 {
					m.searchQuery += key
				}
			}
			return m, nil
		}

		// グローバルキーを先に処理（listに渡さない）
		var cmd tea.Cmd
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit

		case " ":
			cmd = m.togglePlayPause()

		case "n":
			cmd = m.nextTrack()

		case "p":
			cmd = m.previousTrack()

		case "s":
			m.shuffle = !m.shuffle
			cmd = func() tea.Msg {
				if err := m.client.ToggleShuffle(m.ctx, m.shuffle); err != nil {
					return errorMsg(err.Error())
				}
				return nil
			}

		case "r":
			states := []string{"off", "context", "track"}
			for i, s := range states {
				if s == m.repeatState {
					m.repeatState = states[(i+1)%len(states)]
					break
				}
			}
			cmd = func() tea.Msg {
				if err := m.client.SetRepeat(m.ctx, m.repeatState); err != nil {
					return errorMsg(err.Error())
				}
				return nil
			}

		case "tab", "shift+tab":
			// フォーカス切り替え
			if m.focus == FocusSidebar {
				m.focus = FocusMain
			} else {
				m.focus = FocusSidebar
			}
			// listの更新をスキップするため早期リターン
			return m, nil

		case "/":
			m.searchMode = true
			return m, nil

		case "enter":
			if m.focus == FocusSidebar {
				if item, ok := m.playlists.SelectedItem().(playlistItem); ok {
					m.loadingTracks = true
					m.currentPlaylistName = item.name
					if item.id == "liked" {
						cmd = m.fetchSavedTracks()
					} else {
						cmd = m.fetchPlaylistTracks(spotifysdk.ID(item.id))
					}
				}
			} else if m.focus == FocusMain && len(m.tracks) > 0 {
				// プレイリストのコンテキストで再生
				if item, ok := m.trackList.SelectedItem().(trackItem); ok {
					cmd = m.playTrackInPlaylist(item.index)
				}
			}

		case "up", "k", "down", "j":
			if m.focus == FocusSidebar {
				m.playlists, cmd = m.playlists.Update(msg)
			} else if m.focus == FocusMain {
				m.trackList, cmd = m.trackList.Update(msg)
			}

		default:
			// その他のキーはlistに渡す
			if m.focus == FocusSidebar {
				m.playlists, cmd = m.playlists.Update(msg)
			} else if m.focus == FocusMain {
				m.trackList, cmd = m.trackList.Update(msg)
			}
		}

		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 下部プレイヤー(7行)を引く
		contentHeight := msg.Height - 7
		// タイトル分(2行)を引く
		listHeight := contentHeight - 2 - 2
		if listHeight < 3 {
			listHeight = 3
		}
		// 3:4:3 layout
		leftWidth := msg.Width * 3 / 10
		mainWidth := msg.Width * 4 / 10
		m.playlists.SetSize(leftWidth-4, listHeight)
		m.trackList.SetSize(mainWidth-4, listHeight)

	case tickMsg:
		// シークバーをスムーズに更新
		if m.isPlaying && m.currentTrack != nil {
			elapsed := time.Since(m.lastUpdate)
			m.progress += elapsed
			if m.progress > m.duration {
				m.progress = m.duration
			}
		}
		m.lastUpdate = time.Now()

		cmds = append(cmds, tickCmd(), m.fetchCurrentPlayback(), m.fetchQueue(), m.fetchDevices())

	case playbackMsg:
		if msg != nil && msg.Item != nil {
			m.currentTrack = msg
			newPlayingURI := string(msg.Item.URI)
			// 再生中の曲が変わった場合、trackListのアイテムを更新
			if newPlayingURI != m.playingTrackURI {
				m.playingTrackURI = newPlayingURI
				if len(m.trackList.Items()) > 0 {
					selectedIdx := m.trackList.Index()
					m.trackList.SetItems(m.updateTrackListItems(newPlayingURI))
					m.trackList.Select(selectedIdx)
				}
			}
			m.isPlaying = msg.Playing
			m.progress = time.Duration(msg.Progress) * time.Millisecond
			m.duration = time.Duration(msg.Item.Duration) * time.Millisecond
			m.lastUpdate = time.Now()
			m.shuffle = msg.ShuffleState
			m.repeatState = msg.RepeatState
		}

	case playlistsMsg:
		// Liked Songsを先頭に追加
		items := make([]list.Item, 0, len(msg)+1)
		items = append(items, playlistItem{
			id:   "liked", // 特別なID
			name: "💚 Liked Songs",
		})
		for _, pl := range msg {
			items = append(items, playlistItem{
				id:   string(pl.ID),
				name: pl.Name,
			})
		}
		m.playlists.SetItems(items)

	case tracksMsg:
		m.tracks = msg.tracks
		m.currentPlaylistURI = msg.playlistURI
		m.isLikedSongs = false
		m.loadingTracks = false
		// trackListを更新
		items := make([]list.Item, len(msg.tracks))
		for i, t := range msg.tracks {
			items[i] = trackItem{
				index:  i,
				name:   t.Track.Name,
				artist: t.Track.Artists[0].Name,
				uri:    string(t.Track.URI),
			}
		}
		m.trackList.SetItems(items)
		m.trackList.Select(0)

	case savedTracksMsg:
		// SavedTrackをPlaylistTrack形式に変換
		tracks := make([]spotifysdk.PlaylistTrack, len(msg))
		for i, st := range msg {
			tracks[i] = spotifysdk.PlaylistTrack{
				Track: st.FullTrack,
			}
		}
		m.tracks = tracks
		m.isLikedSongs = true
		m.currentPlaylistURI = "" // URIは使用しない
		m.loadingTracks = false
		// trackListを更新
		items := make([]list.Item, len(tracks))
		for i, t := range tracks {
			items[i] = trackItem{
				index:  i,
				name:   t.Track.Name,
				artist: t.Track.Artists[0].Name,
				uri:    string(t.Track.URI),
			}
		}
		m.trackList.SetItems(items)
		m.trackList.Select(0)

	case searchResultsMsg:
		m.searchResults = msg
		m.searchIndex = 0

	case userMsg:
		m.user = msg

	case playStartedMsg:
		m.playingPlaylistName = string(msg)

	case queueMsg:
		if msg != nil {
			m.queue = msg.Items
		}

	case devicesMsg:
		m.devices = msg
		// アクティブなデバイスを見つける
		m.activeDevice = nil
		for i := range msg {
			if msg[i].Active {
				m.activeDevice = &msg[i]
				m.volume = int(msg[i].Volume)
				break
			}
		}

	case errorMsg:
		m.err = string(msg)
		cmds = append(cmds, clearErrorAfter(3*time.Second))
	}

	return m, tea.Batch(cmds...)
}

func clearErrorAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return errorMsg("")
	})
}

type playlistItem struct {
	id   string
	name string
}

func (i playlistItem) FilterValue() string { return i.name }
func (i playlistItem) Title() string       { return i.name }
func (i playlistItem) Description() string { return "" }

type trackItem struct {
	index     int
	name      string
	artist    string
	uri       string
	isPlaying bool
}

func (i trackItem) FilterValue() string { return i.name }
func (i trackItem) Title() string       { return i.name }
func (i trackItem) Description() string { return i.artist }

func (m Model) updateTrackListItems(playingURI string) []list.Item {
	items := m.trackList.Items()
	newItems := make([]list.Item, len(items))
	for i, item := range items {
		if t, ok := item.(trackItem); ok {
			t.isPlaying = t.uri == playingURI
			newItems[i] = t
		} else {
			newItems[i] = item
		}
	}
	return newItems
}

type bindingMap struct{}

func (b bindingMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (b bindingMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
