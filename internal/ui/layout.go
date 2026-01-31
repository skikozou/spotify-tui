package ui

// Layout constants
const (
	// Bottom bar: 4 content lines + 2 border lines = 6 total
	bottomContentLines = 4
	borderSize         = 2
	bottomBarHeight    = bottomContentLines + borderSize

	// Top panel title and spacing
	topPanelTitleLines = 2 // title + empty line
)

// Layout holds calculated dimensions for UI rendering
type Layout struct {
	// Panel widths (including border)
	LeftWidth  int
	MainWidth  int
	RightWidth int

	// Panel content widths (excluding border and padding)
	LeftContentWidth  int
	MainContentWidth  int
	RightContentWidth int

	// Heights
	TopPanelHeight      int // total height including border
	TopContentHeight    int // content area inside border
	ListHeight          int // height for list components
	BottomBarHeight     int // total bottom bar height
	BottomContentHeight int // content area for bottom bar
}

// CalculateLayout computes all UI dimensions based on terminal size
func CalculateLayout(width, height int) Layout {
	// 3:4:3 horizontal split
	leftWidth := (width * 3) / 10
	mainWidth := (width * 4) / 10
	rightWidth := width - leftWidth - mainWidth

	// Vertical split
	topPanelHeight := height - bottomBarHeight
	if topPanelHeight < 5 {
		topPanelHeight = 5
	}

	// Content heights (subtract border)
	topContentHeight := topPanelHeight - borderSize

	// List height (subtract title lines from content)
	listHeight := topContentHeight - topPanelTitleLines
	if listHeight < 3 {
		listHeight = 3
	}

	return Layout{
		LeftWidth:  leftWidth,
		MainWidth:  mainWidth,
		RightWidth: rightWidth,

		// Content width = panel width - border(2) - padding(2)
		LeftContentWidth:  leftWidth - 4,
		MainContentWidth:  mainWidth - 4,
		RightContentWidth: rightWidth - 4,

		TopPanelHeight:      topPanelHeight,
		TopContentHeight:    topContentHeight,
		ListHeight:          listHeight,
		BottomBarHeight:     bottomBarHeight,
		BottomContentHeight: bottomContentLines,
	}
}
