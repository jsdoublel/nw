package tui

import (
	"fmt"
	"io"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jsdoublel/nw/internal/app"
)

type addFilmlistItem struct {
	app.FilmList
	selected bool
}

type addFilmListDelegate struct {
	list.DefaultDelegate
	app *ApplicationTUI
}

type removeListMsg struct {
	ok bool
}

func (d addFilmListDelegate) Height() int   { return 1 }
func (d addFilmListDelegate) Spaceing() int { return 0 }

func (d addFilmListDelegate) Update(msg tea.Msg, ls *list.Model) tea.Cmd {
	item := ls.SelectedItem()
	if item == nil {
		return nil
	}
	fl, ok := item.(*addFilmlistItem)
	if !ok {
		panic(fmt.Sprintf("(Add List) ListSelector item should be addFilmlistItem, instead item is %T", ls.SelectedItem()))
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if !fl.selected {
				if err := d.app.AddList(&fl.FilmList); err != nil {
					log.Print(err.Error())
					return nil
				}
				fl.selected = true
			} else {
				d.app.AskYesNo(fmt.Sprintf("Stop tracking list %s?", fl.Title()), func(b bool) tea.Msg {
					return removeListMsg{ok: b}
				})
			}
		}
	case removeListMsg:
		if !fl.selected {
			panic("should not ask to stop tracking untracked list")
		}
		if msg.ok {
			if err := d.app.RemoveList(&fl.FilmList); err != nil {
				log.Print(err.Error())
				return nil
			}
			fl.selected = false
		}
	}
	return nil
}

func (d addFilmListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	fl, ok := listItem.(*addFilmlistItem)
	if !ok {
		panic(fmt.Sprintf("(Add List) ListSelector item should be addFilmlistItem, instead item is %T", listItem))
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

func MakeAddListPane(a *ApplicationTUI) *SearchModel {
	items := make([]list.Item, 0, len(a.ListHeaders))
	for _, lh := range a.ListHeaders {
		items = append(items, &addFilmlistItem{*lh, a.IsListTracked(lh)})
	}
	return MakeSearchModel(a, items, "Search lists...", addFilmListDelegate{list.NewDefaultDelegate(), a})
}
