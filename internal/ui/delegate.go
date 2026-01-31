package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TrackDelegate はトラックリスト用のカスタムデリゲート
type TrackDelegate struct{}

func NewTrackDelegate() TrackDelegate {
	return TrackDelegate{}
}

func (d TrackDelegate) Height() int                             { return 2 }
func (d TrackDelegate) Spacing() int                            { return 0 }
func (d TrackDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d TrackDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	track, ok := item.(trackItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	isPlaying := track.isPlaying

	// タイトル行
	var titleLine string
	if isPlaying {
		titleLine = fmt.Sprintf(" ♫ %s", track.name)
	} else {
		titleLine = fmt.Sprintf("   %s", track.name)
	}

	// アーティスト行（灰色）
	artistLine := fmt.Sprintf("   %s", track.artist)

	width := m.Width()

	// 幅を制限
	if len(titleLine) > width-2 {
		titleLine = titleLine[:width-5] + "..."
	}
	if len(artistLine) > width-2 {
		artistLine = artistLine[:width-5] + "..."
	}

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	artistStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#B3B3B3"))

	if isSelected {
		titleStyle = titleStyle.Background(lipgloss.Color("#282828")).Bold(true).Foreground(lipgloss.Color("#1DB954"))
		artistStyle = artistStyle.Background(lipgloss.Color("#282828"))
	} else if isPlaying {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#1DB954")).Bold(true)
	}

	fmt.Fprintf(w, "%s\n%s", titleStyle.Width(width).Render(titleLine), artistStyle.Width(width).Render(artistLine))
}

// QueueDelegate はキューリスト用のカスタムデリゲート
type QueueDelegate struct{}

func NewQueueDelegate() QueueDelegate {
	return QueueDelegate{}
}

func (d QueueDelegate) Height() int                             { return 2 }
func (d QueueDelegate) Spacing() int                            { return 0 }
func (d QueueDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d QueueDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	q, ok := item.(queueItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	width := m.Width()

	// タイトル行
	titleLine := fmt.Sprintf("   %s", q.name)
	// アーティスト行（灰色）
	artistLine := fmt.Sprintf("   %s", q.artist)

	// 幅を制限
	titleLine = truncateText(titleLine, width)
	artistLine = truncateText(artistLine, width)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	artistStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#B3B3B3"))

	if isSelected {
		titleStyle = titleStyle.Background(lipgloss.Color("#282828")).Bold(true).Foreground(lipgloss.Color("#1DB954"))
		artistStyle = artistStyle.Background(lipgloss.Color("#282828"))
	}

	fmt.Fprintf(w, "%s\n%s", titleStyle.Width(width).Render(titleLine), artistStyle.Width(width).Render(artistLine))
}

// truncateText はテキストを指定幅に収まるようにカットし、末尾に...を追加する
func truncateText(text string, width int) string {
	if len(text) > width {
		if width > 3 {
			return text[:width-3] + "..."
		}
		return text[:width]
	}
	return text
}

// queueItem はキュー用のリストアイテム
type queueItem struct {
	name   string
	artist string
	uri    string
}

func (i queueItem) FilterValue() string { return i.name }
func (i queueItem) Title() string       { return i.name }
func (i queueItem) Description() string { return i.artist }
