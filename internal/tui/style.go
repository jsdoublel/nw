package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (

	// ----- Colors
	// colors from : https://github.com/slugbyte/lackluster.nvim
	lack   = lipgloss.Color("#708090")
	luster = lipgloss.Color("#deeeed")
	orange = lipgloss.Color("#ffaa88")
	yellow = lipgloss.Color("#abab77")
	green  = lipgloss.Color("#789978")
	blue   = lipgloss.Color("#7788aa")
	red    = lipgloss.Color("#d70000")

	black = lipgloss.Color("#000000")
	gray1 = lipgloss.Color("#080808")
	gray2 = lipgloss.Color("#191919")
	gray3 = lipgloss.Color("#2a2a2a")
	gray4 = lipgloss.Color("#444444")
	gray5 = lipgloss.Color("#555555")
	gray6 = lipgloss.Color("#7a7a7a")
	gray7 = lipgloss.Color("#aaaaaa")
	gray8 = lipgloss.Color("#cccccc")
	gray9 = lipgloss.Color("#dddddd")

	unfocused = gray6
	focused   = gray9

	mainStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).Foreground(focused)

	// ----- Add List Screen
	addListTitleColor = lack
	searchInputStyle  = lipgloss.NewStyle().Inherit(mainStyle)
	searchListStyle   = lipgloss.NewStyle().Inherit(mainStyle)
	cursorStyle       = lipgloss.NewStyle()
	lsStyle           = lipgloss.NewStyle().Inherit(mainStyle)

	// ----- Misc. Prompts
	yesNoStyle    = lipgloss.NewStyle().Inherit(mainStyle)
	yesNoSelected = lipgloss.NewStyle().
			Foreground(black).
			Background(focused).
			Padding(0, 2)
	yesNoUnselected = lipgloss.NewStyle().
			Foreground(gray9).
			Padding(0, 2)
)

// Returns styled list.DefaultDelegate
func listStyleDelegate() list.DefaultDelegate {
	listStyleDele := list.NewDefaultDelegate()
	listStyleDele.Styles.NormalTitle = listStyleDele.Styles.NormalTitle.
		Foreground(gray8)
	listStyleDele.Styles.NormalDesc = listStyleDele.Styles.NormalDesc.
		Foreground(gray7)
	listStyleDele.Styles.DimmedTitle = listStyleDele.Styles.DimmedTitle.
		Foreground(gray6)
	listStyleDele.Styles.DimmedDesc = listStyleDele.Styles.DimmedDesc.
		Foreground(gray5)
	listStyleDele.Styles.SelectedTitle = listStyleDele.Styles.SelectedTitle.
		Foreground(luster).
		BorderForeground(orange).
		Bold(true)
	listStyleDele.Styles.SelectedDesc = listStyleDele.Styles.SelectedDesc.
		Foreground(orange).
		BorderForeground(orange)
	listStyleDele.Styles.FilterMatch = listStyleDele.Styles.FilterMatch.
		Foreground(blue).
		Underline(true)
	return listStyleDele
}
