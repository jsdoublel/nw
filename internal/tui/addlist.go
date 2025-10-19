package tui

import (
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

type addFilmlistItem struct {
	app.FilmList
	selected bool
}

type listHighlightorDelegate struct {
	list.DefaultDelegate
}

// func (d listHighlightorDelegate) Height() int                             { return 1 }
// func (d listHighlightorDelegate) Spaceing() int                           { return 0 }
func (d listHighlightorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d listHighlightorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	fl, ok := listItem.(*addFilmlistItem)
	if !ok {
		panic("ListSelector item should be *addFilmlistItem")
	}
	dd := d.DefaultDelegate
	if fl.selected {
		dd.Styles.NormalTitle = dd.Styles.NormalTitle.
			Foreground(lipgloss.Color("205")).Bold(true)
		dd.Styles.NormalDesc = dd.Styles.NormalDesc.
			Foreground(lipgloss.Color("205"))
	}
	dd.Render(w, m, index, listItem)
}

func MakeAddListPane(a *app.Application) *ListSelector {
	items := make([]list.Item, 0, len(a.ListHeaders))
	for _, lh := range a.ListHeaders {
		items = append(items, &addFilmlistItem{*lh, a.IsListTracked(lh)})
	}
	return MakeListSelector(a, items, listHighlightorDelegate{list.NewDefaultDelegate()}, func(it list.Item) error {
		fl, ok := it.(*addFilmlistItem)
		if !ok {
			panic("(Add List) ListSelector item should be addFilmlistItem")
		}
		if !fl.selected {
			if err := a.AddList(&fl.FilmList); err != nil {
				return err
			}
			fl.selected = true
		}
		return nil
	})
}
