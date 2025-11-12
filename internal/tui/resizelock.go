package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResizeLockModel struct {
	app *ApplicationTUI
}

func (fl *ResizeLockModel) Init() tea.Cmd                           { return nil }
func (fl *ResizeLockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return nil, nil }
func (fl *ResizeLockModel) View() string {
	goodSty := lipgloss.NewStyle().Foreground(green)
	badSty := lipgloss.NewStyle().Foreground(red)
	h := goodSty.Render(strconv.Itoa(fl.app.height))
	w := goodSty.Render(strconv.Itoa(fl.app.width))
	if fl.app.height <= paneHeight {
		h = badSty.Render(strconv.Itoa(fl.app.height))
	}
	if fl.app.width <= paneWidth {
		w = badSty.Render(strconv.Itoa(fl.app.width))
	}
	return lipgloss.JoinVertical(lipgloss.Center,
		"Terminal size too small!",
		fmt.Sprintf("Current: Width = %s, Height = %s", w, h),
		"",
		fmt.Sprintf("Requires: Width = %d, Height = %d", paneWidth, paneHeight),
	)
}
