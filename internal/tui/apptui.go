package tui

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"

	"github.com/jsdoublel/nw/internal/app"
)

type GoBackMsg struct{}

func GoBack() tea.Msg { return GoBackMsg{} }

type UpdateScreenMsg struct{}

func UpdateScreen() tea.Msg { return UpdateScreenMsg{} }

type ScreenStack []tea.Model

func (ss *ScreenStack) push(m tea.Model) { *ss = append(*ss, m) }
func (ss *ScreenStack) pop()             { *ss = (*ss)[:len(*ss)-1] }
func (ss ScreenStack) cur() tea.Model    { return ss[len(ss)-1] }

// Main model struct that drives NW TUI
type ApplicationTUI struct {
	*app.Application
	screens    ScreenStack
	status     StatusBarModel
	resizeLock tea.Model
	help       help.Model
	width      int
	height     int
}

func RunApplicationTUI(username string) error {
	logf, err := tea.LogToFile(filepath.Join(app.NWDataPath, "nw.log"), "")
	if err != nil {
		return fmt.Errorf("could not set up logging, %w", err)
	}
	defer func() { _ = logf.Close() }()
	log.Printf("nw version %s starting...", app.Version)
	if app.ConfigErr != nil {
		log.Printf("error loading config, %s", app.ConfigErr)
	}
	if err := app.GetUser(&username, func() string {
		return AskQuestion("What is your Letterboxd username?", "Username")
	}); err != nil {
		return err
	}
	application, err := app.Load(username)
	if err != nil {
		return fmt.Errorf("could not load application data, %w", err)
	}
	defer application.Shutdown()
	if application.ApiKey == "" {
		application.ApiKey = AskQuestion("What is your TMDB api key?", "API Key")
	}
	application.ApiInit()
	a := ApplicationTUI{Application: application}
	a.resizeLock = &ResizeLockModel{&a}
	p := tea.NewProgram(&a, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (a *ApplicationTUI) Init() tea.Cmd {
	a.status = *MakeStatusBar(a)
	a.help = help.New()
	return updateUserDataCmd(a, app.UpdateExpiredCheck)
}

func (a *ApplicationTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && a.TooSmall() { // block all non-quitting keypresses if screen is blocked
		if key.Matches(msg, keys.Quit) {
			return a, tea.Quit
		}
		return a, nil
	}
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = a.width
	case userDataLoadedMsg:
		a.screens.pop()          // remove loading screen
		if len(a.screens) == 0 { // we need different behavior on startup vs. update
			a.screens.push(MakeMainScreen(a))
		} else {
			return a, UpdateScreen
		}
	case userDataFailedMsg:
		if ss, ok := a.screens.cur().(*SplashScreenModel); ok {
			ss.SetError(msg.err)
		} else {
			panic("userDataFailedMsg received without splash screen")
		}
	case statusMessageMsg:
		cmds = append(cmds, a.status.setMessage(msg.message))
	case GoBackMsg:
		if len(a.screens) == 1 {
			return a, tea.Quit
		}
		_, wasLoading := a.screens.cur().(*SplashScreenModel)
		a.screens.pop()
		if _, ok := a.screens.cur().(*MainScreen); ok && !wasLoading {
			return a, updateUserDataCmd(a, app.UpdateNeverCheck)
		}
		return a, UpdateScreen
	case tea.KeyMsg:
		if a.loading() {
			break
		}
		if cmd, handled := a.checkKeyMsgs(msg); handled {
			return a, cmd
		}
	}
	cmds = append(cmds, a.UpdateRouter(msg)...)
	return a, tea.Batch(cmds...)
}

func (a *ApplicationTUI) checkKeyMsgs(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, keys.Update):
		return updateUserDataCmd(a, app.UpdateAlwaysCheck), true
	case key.Matches(msg, keys.StopWatch):
		a.StopDiscordRPC()
		return updateUserDataCmd(a, app.UpdateNeverCheck), true
	case key.Matches(msg, keys.Help):
		a.help.ShowAll = !a.help.ShowAll
		return nil, true
	case key.Matches(msg, keys.About):
		a.Popup(About)
		return nil, true
	case key.Matches(msg, keys.Quit):
		return tea.Quit, true
	default:
		return nil, false
	}
}

func (a *ApplicationTUI) View() string {
	if a.TooSmall() {
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, a.resizeLock.View())
	}
	cur := a.screens.cur()
	main := lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, cur.View())
	if _, ok := cur.(*SplashScreenModel); ok {
		return main
	}
	compStatus := overlay.Composite(a.status.View(), main, overlay.Left, overlay.Top, 0, 0)
	return overlay.Composite(a.help.View(keys), compStatus, overlay.Left, overlay.Bottom, 0, 0)
}

// Handle update routing with overlays
func (a *ApplicationTUI) UpdateRouter(msg tea.Msg) []tea.Cmd {
	var c, bc, sc tea.Cmd
	_, sc = a.status.Update(msg)
	// updating overlay model before non-overlay model avoids issue where a
	// command that launches an overlay, is also sent to an overlay
	if ov, ok := a.screens.cur().(*overlay.Model); ok {
		_, c = ov.Foreground.Update(msg)
		if msg, ok := msg.(UpdateScreenMsg); ok {
			_, bc = ov.Background.Update(msg)
		}
		return []tea.Cmd{c, sc, bc}
	}
	_, c = a.screens.cur().Update(msg)
	return []tea.Cmd{c, sc}
}

func (a *ApplicationTUI) TooSmall() bool {
	return a.width <= paneWidth || a.height <= paneHeight
}

// ---------- User Data Updating

type userDataLoadedMsg struct{}            // success updating user data
type userDataFailedMsg struct{ err error } // failure updating user data

// Updates users data from Letterboxd. There are three possible check conditions:
//  1. `app.UpdateAlwaysCheck`: Always do a full Letterboxd scrape.
//  2. `app.UpdateExpiredCheck`: Do a full Letterboxd scrape if we have not checked in a while.
//  2. `app.UpdateNeverCheck`: Only do a quick check of the RSS feed (if it's not on cooldown).
//
// IMPORTANT: Don't call this if it is already running.
func updateUserDataCmd(a *ApplicationTUI, check app.CheckUpdateCondition) tea.Cmd {
	if a.loading() {
		return nil
	}
	// Don't check if we're only trying to do an RSS check but we're on cool down.
	if check == app.UpdateNeverCheck && !a.CanCheckRSS() {
		return func() tea.Msg { return userDataLoadedMsg{} } // return cmd as we need to make sure we still update screen
	}
	splash, cmd := MakeSplashScreen()
	a.screens.push(splash)
	return tea.Batch(cmd, tea.SetWindowTitle("nw"), func() tea.Msg {
		if a.CanCheckRSS() {
			if err := a.QuickUpdateWatched(); err != nil {
				log.Printf("failed to execute quick update on startup, %s", err)
			}
		}
		if err := a.UpdateUserData(check); err != nil {
			return userDataFailedMsg{err}
		}
		return userDataLoadedMsg{}
	})
}

// Check if app is currently loading
func (app *ApplicationTUI) loading() bool {
	if len(app.screens) != 0 {
		if _, ok := app.screens.cur().(*SplashScreenModel); ok {
			return true
		}
	}
	return false
}
