package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Message struct {
	text    string
	error   bool
	timeout time.Duration
}

type statusEntry struct {
	id      int
	message Message
}

type StatusBarModel struct {
	messages []statusEntry
	nextId   int
	app      *ApplicationTUI
}

func (sb *StatusBarModel) Init() tea.Cmd { return nil }

func (sb *StatusBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(statusClearMsg); ok {
		for i, entry := range sb.messages {
			if entry.id == msg.id {
				sb.messages = append(sb.messages[:i], sb.messages[i+1:]...)
				break
			}
		}
	}
	return sb, nil
}

func (sb *StatusBarModel) View() string {
	strs := make([]string, 0)
	if sb.app.DiscordRPC.Watching() {
		strs = append(strs, statusBarWatchingStyle.Render(
			fmt.Sprintf("Watching %s, press %s to stop", sb.app.DiscordRPC, keys.StopWatch.Help().Key),
		))
	}
	for _, entry := range sb.messages {
		if entry.message.error {
			strs = append(strs, statusBarErrStyle.Render(entry.message.text))
		} else {
			strs = append(strs, statusBarMessageStyle.Render(entry.message.text))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, strs...)
}

type statusClearMsg struct{ id int }
type statusMessageMsg struct{ message Message }

func statusMessageCmd(message Message) tea.Cmd {
	return func() tea.Msg { return statusMessageMsg{message: message} }
}

func (sb *StatusBarModel) setMessage(message Message) tea.Cmd {
	id := sb.nextId
	sb.messages = append(sb.messages, statusEntry{id: id, message: message})
	sb.nextId++
	return tea.Tick(message.timeout, func(time.Time) tea.Msg {
		return statusClearMsg{id: id}
	})
}

func MakeStatusBar(a *ApplicationTUI) *StatusBarModel {
	return &StatusBarModel{
		messages: make([]statusEntry, 0),
		app:      a,
	}
}
