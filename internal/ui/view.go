package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#1DB954") // Spotify Green
	secondaryColor = lipgloss.Color("#FFFFFF")
	accentColor    = lipgloss.Color("#B3B3B3")
	bgColor        = lipgloss.Color("#121212")
	highlightColor = lipgloss.Color("#282828")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	sidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(0)

	mainPanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(0)

	playerBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	focusedStyle = lipgloss.NewStyle().
			BorderForeground(primaryColor)

	trackStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	selectedTrackStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Background(highlightColor)

	playingTrackStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1DB954")).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// 3:4:3 layout
	leftWidth := (m.width * 3) / 10
	mainWidth := (m.width * 4) / 10
	rightWidth := m.width - leftWidth - mainWidth

	// Bottom bar height: 7 total (5 content + 2 border)
	bottomBarHeight := 7

	// Content area is remaining height
	contentHeight := m.height - bottomBarHeight
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Render top row content
	sidebarContent := m.renderSidebar(leftWidth-2, contentHeight-2)
	mainContent := m.renderMainPanel(mainWidth-2, contentHeight-2)
	queueContent := m.renderQueue(rightWidth-2, contentHeight-2)

	// Apply borders and styling
	sidebarStyleFinal := sidebarStyle.
		Width(leftWidth - 2).
		Height(contentHeight)

	mainPanelStyleFinal := mainPanelStyle.
		Width(mainWidth - 2).
		Height(contentHeight)

	rightPanelStyleFinal := mainPanelStyle.Copy().
		Width(rightWidth - 2).
		Height(contentHeight)

	if m.focus == FocusSidebar {
		sidebarStyleFinal = sidebarStyleFinal.Copy().BorderForeground(primaryColor)
	} else {
		mainPanelStyleFinal = mainPanelStyleFinal.Copy().BorderForeground(primaryColor)
	}

	// Top row: sidebar + main + queue
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarStyleFinal.Render(sidebarContent),
		mainPanelStyleFinal.Render(mainContent),
		rightPanelStyleFinal.Render(queueContent),
	)

	// Bottom bar: user info (left) + player bar (center) + device info (right)
	userInfoContent := m.renderUserInfo(leftWidth - 4)
	playerBarContent := m.renderPlayerBar(mainWidth - 4)
	deviceInfoContent := m.renderDeviceInfo(rightWidth - 4)

	userInfoFinal := playerBarStyle.
		Width(leftWidth - 2).
		Render(userInfoContent)

	playerBarFinal := playerBarStyle.
		Width(mainWidth - 2).
		Render(playerBarContent)

	deviceInfoFinal := playerBarStyle.
		Width(rightWidth - 2).
		Render(deviceInfoContent)

	bottomRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		userInfoFinal,
		playerBarFinal,
		deviceInfoFinal,
	)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
}

func (m Model) renderSidebar(width, height int) string {
	title := titleStyle.Render(" 🎵 My Library")

	var content string
	if len(m.playlists.Items()) == 0 {
		content = lipgloss.Place(
			width, height-4,
			lipgloss.Center, lipgloss.Center,
			"Loading...",
		)
	} else {
		content = m.playlists.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content)
}

func (m Model) renderMainPanel(width, height int) string {
	if m.searchMode {
		return m.renderSearchView(width, height)
	}

	if m.loadingTracks {
		return lipgloss.Place(
			width, height,
			lipgloss.Center, lipgloss.Center,
			"Loading tracks...",
		)
	}

	if len(m.tracks) == 0 {
		return lipgloss.Place(
			width, height,
			lipgloss.Center, lipgloss.Center,
			"Select a playlist from the sidebar",
		)
	}

	title := titleStyle.Render(" 📀 Tracks")
	content := m.trackList.View()

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content)
}

func (m Model) renderSearchView(width, height int) string {
	var lines []string
	title := titleStyle.Render(" 🔍 Search")
	query := fmt.Sprintf(" Query: %s_", m.searchQuery)
	lines = append(lines, title, "", query, "")

	if len(m.searchResults) == 0 {
		if m.searchQuery == "" {
			hint := lipgloss.NewStyle().
				Foreground(accentColor).
				Render(" Type to search, press Enter to execute, Esc to exit")
			lines = append(lines, hint)
		} else {
			lines = append(lines, " No results found")
		}
		return strings.Join(lines, "\n")
	}

	// 検索結果を表示
	lines = append(lines, lipgloss.NewStyle().
		Foreground(accentColor).
		Render(fmt.Sprintf(" Found %d results:", len(m.searchResults))))
	lines = append(lines, "")

	// スクロール可能な結果リスト
	visibleLines := height - 6
	if visibleLines < 1 {
		visibleLines = 1
	}

	start := m.searchIndex
	if start > len(m.searchResults)-visibleLines {
		start = len(m.searchResults) - visibleLines
	}
	if start < 0 {
		start = 0
	}

	for i := start; i < len(m.searchResults) && i < start+visibleLines; i++ {
		track := m.searchResults[i]
		line := fmt.Sprintf(" %2d. %s - %s",
			i+1,
			track.Name,
			track.Artists[0].Name,
		)

		if i == m.searchIndex {
			line = selectedTrackStyle.Render(" ▶" + line)
		} else {
			line = trackStyle.Render("  " + line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderUserInfo(width int) string {
	var lines []string

	title := titleStyle.Render("👤 User")
	lines = append(lines, title)

	if m.user != nil {
		// 名前とIDの表示（幅に収まらない場合はIDを省略）
		nameWithID := fmt.Sprintf(" Name:      %s (%s)", m.user.DisplayName, m.user.ID)
		nameOnly := fmt.Sprintf(" Name:      %s", m.user.DisplayName)
		if len(nameWithID) <= width {
			lines = append(lines, nameWithID)
		} else {
			lines = append(lines, nameOnly)
		}
		if m.user.Product != "" {
			lines = append(lines, fmt.Sprintf(" Plan:      %s", m.user.Product))
		}
		lines = append(lines, fmt.Sprintf(" Followers: %d", m.user.Followers.Count))
	} else {
		lines = append(lines, " Loading...")
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPlayerBar(width int) string {
	var lines []string

	// Context info (playlist/album name)
	contextInfo := ""
	if m.currentTrack != nil && m.currentTrack.PlaybackContext.Type != "" {
		switch m.currentTrack.PlaybackContext.Type {
		case "playlist":
			// 再生開始時のプレイリスト名を使用
			if m.playingPlaylistName != "" {
				contextInfo = m.playingPlaylistName
			} else {
				contextInfo = "Playlist"
			}
		case "album":
			if m.currentTrack.Item != nil {
				contextInfo = m.currentTrack.Item.Album.Name
			}
		case "artist":
			contextInfo = "Artist"
		case "collection":
			contextInfo = "Liked Songs"
		default:
			contextInfo = string(m.currentTrack.PlaybackContext.Type)
		}
	} else if m.playingPlaylistName != "" {
		// コンテキストがない場合でも再生中のプレイリスト名を表示
		contextInfo = m.playingPlaylistName
	}

	// Track info
	trackInfo := "No track playing"
	if m.currentTrack != nil && m.currentTrack.Item != nil {
		if contextInfo != "" {
			trackInfo = fmt.Sprintf("♫ %s - %s | %s",
				m.currentTrack.Item.Name,
				m.currentTrack.Item.Artists[0].Name,
				contextInfo,
			)
		} else {
			trackInfo = fmt.Sprintf("♫ %s - %s",
				m.currentTrack.Item.Name,
				m.currentTrack.Item.Artists[0].Name,
			)
		}
	}
	lines = append(lines, trackInfo)

	// Progress bar
	progressBar := m.renderProgressBar(width)
	lines = append(lines, progressBar)

	// Controls
	playPauseIcon := "⏸"
	if !m.isPlaying {
		playPauseIcon = "▶"
	}

	shuffleIcon := "🔀"
	if !m.shuffle {
		shuffleIcon = "➡"
	}

	repeatIcon := "🔁"
	switch m.repeatState {
	case "track":
		repeatIcon = "🔂"
	case "off":
		repeatIcon = "➡"
	}

	controls := fmt.Sprintf("%s %s %s", shuffleIcon, playPauseIcon, repeatIcon)
	lines = append(lines, controls)

	// Keybindings
	keybindings := "[Space] Play/Pause | [n] Next | [p] Prev | [Tab] Switch | [/] Search | [q] Quit"
	if m.err != "" {
		keybindings = errorStyle.Render("Error: " + m.err)
	}
	// 幅に収まらない場合はカットして...を追加
	if len(keybindings) > width {
		keybindings = keybindings[:width-3] + "..."
	}
	lines = append(lines, keybindings)

	return strings.Join(lines, "\n")
}

func (m Model) renderProgressBar(width int) string {
	if m.currentTrack == nil || m.duration == 0 {
		return "[" + strings.Repeat("░", width-20) + "] 0:00 / 0:00"
	}

	barWidth := width - 20
	progress := float64(m.progress) / float64(m.duration)
	filled := int(progress * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	currentTime := formatDuration(m.progress)
	totalTime := formatDuration(m.duration)

	return fmt.Sprintf("[%s] %s / %s", bar, currentTime, totalTime)
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func (m Model) renderQueue(width, height int) string {
	var lines []string
	title := titleStyle.Render(" 📋 Queue")
	lines = append(lines, title, "")

	if len(m.queue) == 0 {
		lines = append(lines, " No tracks in queue")
		return strings.Join(lines, "\n")
	}

	// 表示可能な行数（2行/曲）
	visibleTracks := (height - 3) / 2
	if visibleTracks < 1 {
		visibleTracks = 1
	}

	for i := 0; i < len(m.queue) && i < visibleTracks; i++ {
		track := m.queue[i]
		trackName := track.Name
		artistName := track.Artists[0].Name

		// 幅制限
		if len(trackName) > width-3 {
			trackName = trackName[:width-6] + "..."
		}
		if len(artistName) > width-3 {
			artistName = artistName[:width-6] + "..."
		}

		lines = append(lines, fmt.Sprintf(" %s", trackName))
		artistLine := lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf(" %s", artistName))
		lines = append(lines, artistLine)
	}

	if len(m.queue) > visibleTracks {
		lines = append(lines, fmt.Sprintf(" ... +%d more", len(m.queue)-visibleTracks))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderDeviceInfo(width int) string {
	var lines []string

	title := titleStyle.Render("🔊 Device")
	lines = append(lines, title)

	if m.activeDevice != nil {
		deviceName := m.activeDevice.Name
		if len(deviceName) > width-2 {
			deviceName = deviceName[:width-5] + "..."
		}
		lines = append(lines, fmt.Sprintf(" %s", deviceName))
		lines = append(lines, fmt.Sprintf(" Type: %s", m.activeDevice.Type))

		// Volume bar
		volBarWidth := width - 12
		if volBarWidth < 5 {
			volBarWidth = 5
		}
		volFilled := (m.volume * volBarWidth) / 100
		volBar := strings.Repeat("█", volFilled) + strings.Repeat("░", volBarWidth-volFilled)
		lines = append(lines, fmt.Sprintf(" Vol: [%s]", volBar))
	} else if len(m.devices) > 0 {
		lines = append(lines, " No active device")
	} else {
		lines = append(lines, " Loading...")
	}

	return strings.Join(lines, "\n")
}
