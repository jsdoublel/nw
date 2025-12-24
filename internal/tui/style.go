package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

const (
	paneWidth  = 64
	paneHeight = 31

	ellipse = '\u2026'
	hSep    = '\u2500'
	cursor  = '\u2588'

	Title    = "NW – Next Watch"
	Subtitle = "A TUI utility for selecting films to watch from Letterboxd (powered by TMDB)."
	License  = `Copyright (C) 2025 James Willson <jsdoublel@gmail.com>

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <https://www.gnu.org/licenses/>.


This software uses TMDB and the TMDB APIs but is not endorsed, certified, or 
otherwise approved by TMDB.`
)

var (

	// ----- Colors
	// colors from : https://github.com/slugbyte/lackluster.nvim
	lack   = paletteColor(app.Config.Appearance.Colors.Primary, "#708090")
	luster = lipgloss.Color("#deeeed")
	orange = paletteColor(app.Config.Appearance.Colors.Secondary, "#ffaa88")
	yellow = lipgloss.Color("#abab77")
	green  = paletteColor(app.Config.Appearance.Colors.Success, "#789978")
	blue   = lipgloss.Color("#7788aa")
	red    = paletteColor(app.Config.Appearance.Colors.Error, "#d70000")

	gray = grayColors(app.Config.Appearance.LightMode)

	unfocusedColor       = gray[4]
	focusedColor         = gray[6]
	focusedButtonColor   = gray[7]
	unfocusedButtonColor = gray[3]
	textColor            = gray[9]
	textDark             = gray[0]

	mainStyle = mainStyler()

	// ----- Add List Screen
	addListTitleColor = lack
	searchInputStyle  = lipgloss.NewStyle().Inherit(mainStyle)
	searchListStyle   = lipgloss.NewStyle().Inherit(mainStyle)
	cursorStyle       = lipgloss.NewStyle()
	lsStyle           = lipgloss.NewStyle().Inherit(mainStyle)

	// ----- NW Queue
	nwStyle             = lipgloss.NewStyle().Inherit(mainStyle)
	nwItemStyle         = lipgloss.NewStyle()
	nwSelectedItemStyle = lipgloss.NewStyle().Background(lack).Foreground(textDark)
	nwUpdatedItemStyle  = lipgloss.NewStyle().Foreground(green)
	nwSeparatorStyle    = lipgloss.NewStyle().Foreground(focusedColor)

	// ----- Model Join (Joint NW Queue / Details)
	// musts set height based on screen when used
	joinTallThin = lipgloss.NewStyle().PaddingBottom(4).AlignVertical(lipgloss.Bottom)

	// ----- Film Details
	filmDetailsStyle    = lipgloss.NewStyle().Inherit(mainStyle)
	filmTextStyle       = lipgloss.NewStyle().Width(paneWidth).Foreground(textColor)
	filmTitleStyle      = lipgloss.NewStyle().Inherit(filmTextStyle).Bold(true)
	flimDirStyle        = lipgloss.NewStyle().Inherit(filmTextStyle).Italic(true)
	filmCastHeaderStyle = lipgloss.NewStyle().Inherit(filmTextStyle).Underline(true)
	filmActionSelected  = lipgloss.NewStyle().
				Foreground(textDark).
				Background(focusedButtonColor).
				Padding(0, 2)
	filmActionUnselected = lipgloss.NewStyle().
				Foreground(textColor).
				Background(unfocusedButtonColor).
				Padding(0, 2)

	// ----- Film Search
	filmSearchItemStyle     = lipgloss.NewStyle()
	filmSearchSelectedStyle = lipgloss.NewStyle().Background(lack).Foreground(textDark)

	// ----- Status Bar
	statusBarWatchingStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(green).
				Foreground(green).
				Padding(0, 1)
	statusBarMessageStyle = lipgloss.NewStyle().Width(paneWidth).
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1)
	statusBarErrStyle = lipgloss.NewStyle().Width(paneWidth).
				Foreground(red).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(red).
				Padding(0, 1)

	// ----- Splash Screen
	splashSpinnerStyles = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(orange),
		lipgloss.NewStyle().Foreground(green),
		lipgloss.NewStyle().Foreground(lack),
	}

	// ----- Misc. Prompts
	yesNoStyle    = lipgloss.NewStyle().Inherit(mainStyle)
	questionStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			Padding(1, 2)
	yesNoSelected = lipgloss.NewStyle().
			Foreground(textDark).
			Background(focusedButtonColor).
			Padding(0, 2)
	yesNoUnselected = lipgloss.NewStyle().
			Foreground(textColor).
			Background(unfocusedButtonColor).
			Padding(0, 2)
	popupStyle     = lipgloss.NewStyle().Inherit(mainStyle)
	popupTextStyle = lipgloss.NewStyle().Padding(1)
	popupOkStyle   = lipgloss.NewStyle().
			Foreground(gray[0]).
			Background(focusedButtonColor).
			Padding(0, 2)
	About = strings.Join([]string{
		lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s (%s)", Title, app.Version)),
		lipgloss.NewStyle().Italic(true).Render(Subtitle),
		popupTextStyle.Render(License),
	}, "\n")

	startupTextStyle  = lipgloss.NewStyle().Foreground(gray[8]).Italic(true)
	startupInputStyle = lipgloss.NewStyle().Foreground(textColor)
)

func mainStyler() lipgloss.Style {
	mainStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(focusedColor).
		Foreground(textColor)
	bStyles := map[string]lipgloss.Border{
		"rounded": lipgloss.RoundedBorder(),
		"normal":  lipgloss.NormalBorder(),
		"square":  lipgloss.NormalBorder(),
		"double":  lipgloss.DoubleBorder(),
	}
	if border, ok := bStyles[strings.ToLower(app.Config.Appearance.Border)]; ok {
		mainStyle = mainStyle.BorderStyle(border)
	}
	return mainStyle
}

// Returns styled list.DefaultDelegate
func listStyleDelegate() list.DefaultDelegate {
	listStyleDele := list.NewDefaultDelegate()
	listStyleDele.Styles.NormalTitle = listStyleDele.Styles.NormalTitle.
		Foreground(gray[8])
	listStyleDele.Styles.NormalDesc = listStyleDele.Styles.NormalDesc.
		Foreground(gray[7])
	listStyleDele.Styles.DimmedTitle = listStyleDele.Styles.DimmedTitle.
		Foreground(gray[6])
	listStyleDele.Styles.DimmedDesc = listStyleDele.Styles.DimmedDesc.
		Foreground(gray[5])
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

func paletteColor(cfg string, fallback string) lipgloss.Color {
	if cfg == "" {
		return lipgloss.Color(fallback)
	}
	return lipgloss.Color(cfg)
}

func grayColors(lightMode bool) []lipgloss.Color {
	grays := []lipgloss.Color{
		lipgloss.Color("#000000"),
		lipgloss.Color("#080808"),
		lipgloss.Color("#191919"),
		lipgloss.Color("#2a2a2a"),
		lipgloss.Color("#444444"),
		lipgloss.Color("#555555"),
		lipgloss.Color("#7a7a7a"),
		lipgloss.Color("#aaaaaa"),
		lipgloss.Color("#cccccc"),
		lipgloss.Color("#dddddd"),
	}
	if lightMode {
		slices.Reverse(grays)
	}
	return grays
}
