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
			case "up", "k":
				if len(m.searchResults) > 0 && m.searchIndex > 0 {
					m.searchIndex--
				}
				return m, nil
			case "down", "j":
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
					cmd = m.fetchPlaylistTracks(item.id)
				}
			} else if m.focus == FocusMain && len(m.tracks) > 0 {
				// プレイリストのコンテキストで再生
				cmd = m.playTrackInPlaylist(m.trackIndex)
			}

		case "up", "k":
			if m.focus == FocusSidebar {
				m.playlists, cmd = m.playlists.Update(msg)
			} else if m.focus == FocusMain && m.trackIndex > 0 {
				m.trackIndex--
			}

		case "down", "j":
			if m.focus == FocusSidebar {
				m.playlists, cmd = m.playlists.Update(msg)
			} else if m.focus == FocusMain && m.trackIndex < len(m.tracks)-1 {
				m.trackIndex++
			}
		default:
			// その他のキーはlistに渡す
			if m.focus == FocusSidebar {
				m.playlists, cmd = m.playlists.Update(msg)
			}
		}

		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// タイトル分(2行)と下部プレイヤー(7行)を引く
		listHeight := msg.Height - 9 - 2
		if listHeight < 3 {
			listHeight = 3
		}
		m.playlists.SetSize(msg.Width*3/10-4, listHeight)

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

		cmds = append(cmds, tickCmd(), m.fetchCurrentPlayback())

	case playbackMsg:
		if msg != nil && msg.Item != nil {
			m.currentTrack = msg
			m.isPlaying = msg.Playing
			m.progress = time.Duration(msg.Progress) * time.Millisecond
			m.duration = time.Duration(msg.Item.Duration) * time.Millisecond
			m.lastUpdate = time.Now()
		}

	case playlistsMsg:
		items := make([]list.Item, len(msg))
		for i, pl := range msg {
			items[i] = playlistItem{
				id:   pl.ID,
				name: pl.Name,
			}
		}
		m.playlists.SetItems(items)

	case tracksMsg:
		m.tracks = msg.tracks
		m.currentPlaylistURI = msg.playlistURI
		m.trackIndex = 0

	case searchResultsMsg:
		m.searchResults = msg
		m.searchIndex = 0

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
	id   spotifysdk.ID
	name string
}

func (i playlistItem) FilterValue() string { return i.name }
func (i playlistItem) Title() string       { return i.name }
func (i playlistItem) Description() string { return "" }

type bindingMap struct{}

func (b bindingMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (b bindingMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
