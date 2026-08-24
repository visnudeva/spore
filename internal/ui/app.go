// Package ui implements the spore terminal user interface using Bubbletea.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spore-player/spore/internal/audio"
	"github.com/spore-player/spore/internal/radio"
	"github.com/spore-player/spore/internal/storage"
	"github.com/spore-player/spore/internal/vis"
)

type screen int

const (
	screenLocal screen = iota
	screenRadio
	screenFavorites
	numScreens
)

var screenNames = [numScreens]string{"Files", "Radio", "Favorites"}

// App is the root Bubbletea model.
type App struct {
	width, height int
	screen        screen

	player     audio.Engine
	sampleRate int

	visualizer *vis.Visualizer
	visEnabled bool

	nowPlaying  string
	streamTitle string
	isPlaying   bool
	isPaused    bool

	radioScreen     RadioModel
	localScreen     LocalModel
	favScreen       FavModel
	statusMsg       string
	statusExpiry    time.Time

	favorites []storage.FavoriteStation
	history   []storage.HistoryEntry

	// Last radio station, restored from session and updated on play.
	lastStationUUID string
	lastStationName string
	lastStationURL  string
	resumeOnStart   bool // auto-play lastStationURL once after Init
}

// New creates a new App.
func New() (*App, error) {
	q := audio.Quality{
		SampleRate:      44100,
		BufferMs:        100,
		ResampleQuality: 2,
		BitDepth:        16,
	}
	p, err := audio.New(q)
	if err != nil {
		return nil, fmt.Errorf("audio init: %w", err)
	}

	favs, _ := storage.LoadFavorites()
	hist, _ := storage.LoadHistory()
	sess, _ := storage.LoadSession()
	if sess.StationURL == "" {
		sess = seedSessionFromHistory(sess, hist, favs)
	}

	v := vis.NewVisualizer(float64(q.SampleRate))
	v.Mode = vis.VisBars
	if sess.VisMode != "" {
		if mode, ok := vis.StringToVisModeExact(sess.VisMode); ok {
			v.SetMode(mode)
		}
	}

	visEnabled := true
	if sess.VisMode != "" {
		visEnabled = sess.VisEnabled
	}

	a := &App{
		player:          p,
		sampleRate:      q.SampleRate,
		visualizer:      v,
		visEnabled:      visEnabled,
		favorites:       favs,
		history:         hist,
		screen:          screenLocal,
		lastStationUUID: sess.StationUUID,
		lastStationName: sess.StationName,
		lastStationURL:  sess.StationURL,
		resumeOnStart:   sess.StationURL != "",
	}
	if a.resumeOnStart {
		if a.lastStationUUID != "" && storage.IsFavorite(favs, a.lastStationUUID) {
			a.screen = screenFavorites
		} else {
			a.screen = screenRadio
		}
	}
	a.radioScreen = newRadioModel(a)
	a.localScreen = newLocalModel(a)
	a.favScreen = newFavModel(a)
	return a, nil
}

func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd(), visTickCmd()}
	if a.resumeOnStart {
		a.resumeOnStart = false
		s := radio.Station{
			StationUUID: a.lastStationUUID,
			Name:        a.lastStationName,
			URL:         a.lastStationURL,
		}
		cmds = append(cmds, a.PlayStation(s))
	}
	return tea.Batch(cmds...)
}

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

func visTickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg { return visTick{} })
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.visualizer.Cols = msg.Width
		a.visualizer.Rows = msg.Height
		vis.PanelWidth = msg.Width
		a.radioScreen.width = msg.Width
		a.radioScreen.height = msg.Height
		a.localScreen.width = msg.Width
		a.localScreen.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			_ = storage.SaveSession(a.sessionSnapshot())
			a.player.Stop()
			a.player.Close()
			return a, tea.Quit
		case "tab":
			a.screen = (a.screen + 1) % numScreens
		case "shift+tab":
			a.screen = (a.screen - 1 + numScreens) % numScreens
		default:
			return a.delegateKey(msg, cmds)
		}

	case tickMsg:
		a.isPlaying = a.player.IsPlaying()
		a.isPaused = a.player.IsPaused()
		if t := a.player.StreamTitle(); t != "" && t != a.streamTitle {
			a.streamTitle = t
		}
		cmds = append(cmds, tickCmd())

	case visTick:
		if a.visEnabled && a.isPlaying {
			ctx := vis.VisTickContext{
				Now:     time.Now(),
				Playing: a.isPlaying,
				Paused:  a.isPaused,
				Analyze: func(spec vis.VisAnalysisSpec) []float64 {
					buf := a.visualizer.EnsureSampleBuf(spec.FFTSize)
					n := a.player.SamplesInto(buf)
					if n == 0 {
						return nil
					}
					return a.visualizer.Analyze(buf[:n], spec)
				},
				StereoSamplesInto: func(dst [][2]float64) int {
					return a.player.StereoSamplesInto(dst)
				},
			}
			a.visualizer.Tick(ctx)
		}
		cmds = append(cmds, visTickCmd())

	case stationsLoadedMsg:
		cmd := a.radioScreen.update(msg)
		cmds = append(cmds, cmd)

	case playStartedMsg:
		a.nowPlaying = msg.name
		a.streamTitle = ""
		a.isPlaying = true
		a.isPaused = false
		entry := storage.HistoryEntry{
			UUID: msg.uuid,
			Name: msg.name,
			URL:  msg.url,
		}
		if !msg.radio {
			entry.IsLocal = true
		} else {
			a.lastStationUUID = msg.uuid
			a.lastStationName = msg.name
			a.lastStationURL = msg.url
			a.persistSession()
		}
		a.history = storage.AppendHistory(a.history, entry)
		go storage.SaveHistory(a.history)

	case errMsg:
		a.statusMsg = "Error: " + msg.err.Error()
		a.statusExpiry = time.Now().Add(5 * time.Second)

	case favoritesChangedMsg:
		a.favScreen = newFavModel(a)
	}

	return a, tea.Batch(cmds...)
}

func (a *App) delegateKey(msg tea.KeyPressMsg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	key := msg.String()
	radioTyping := a.screen == screenRadio && a.radioScreen.typing

	if !radioTyping {
		switch key {
		case " ", "space", "p":
			if a.isPlaying || a.nowPlaying != "" {
				a.player.TogglePause()
				a.isPaused = a.player.IsPaused()
			}
		case "s":
			a.player.Stop()
			a.isPlaying = false
			a.isPaused = false
			a.nowPlaying = ""
			a.streamTitle = ""
		case "+", "=":
			a.player.SetVolume(min64(a.player.Volume()+2, 6))
		case "-":
			a.player.SetVolume(max64(a.player.Volume()-2, -30))
		case "m":
			if a.player.Volume() > -30 {
				a.player.SetVolume(-99)
			} else {
				a.player.SetVolume(0)
			}
		case "v":
			a.visualizer.CycleMode()
			a.persistSession()
		case "V":
			a.visEnabled = !a.visEnabled
			a.persistSession()
		}
	}

	switch a.screen {
	case screenRadio:
		cmds = append(cmds, a.radioScreen.update(msg))
	case screenLocal:
		cmds = append(cmds, a.localScreen.update(msg))
	case screenFavorites:
		cmds = append(cmds, a.favScreen.update(msg))
	}

	return a, tea.Batch(cmds...)
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (a *App) View() tea.View {
	if a.width == 0 {
		v := tea.NewView("Loading spore...")
		v.AltScreen = true
		return v
	}

	w, h := a.width, a.height
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	var ui strings.Builder
	ui.WriteString(a.renderTabBar())
	ui.WriteByte('\n')
	ui.WriteString(a.renderNowPlaying())
	ui.WriteByte('\n')

	contentHeight := h - 5
	if contentHeight < 4 {
		contentHeight = 4
	}
	switch a.screen {
	case screenRadio:
		a.radioScreen.height = contentHeight
		ui.WriteString(a.radioScreen.view())
	case screenFavorites:
		ui.WriteString(a.favScreen.view())
	default:
		a.localScreen.height = contentHeight
		ui.WriteString(a.localScreen.view())
	}
	ui.WriteByte('\n')
	ui.WriteString(a.renderStatusBar())

	screen := ui.String()
	if a.visEnabled && (a.isPlaying || a.isPaused) {
		a.visualizer.Cols = w
		a.visualizer.Rows = h
		vis.PanelWidth = w
		if visStr := a.visualizer.Render(); visStr != "" {
			screen = overlayOnVis(visStr, screen, w, h)
		}
	}

	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}

// ─── Rendering helpers ───────────────────────────────────────────────────────

var (
	tabActiveStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10)).Underline(true)
	tabInactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))
	headerStyle      = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11)).Bold(true)
	dimStyle         = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(8))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(9))
	greenStyle       = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(10))
)

func (a *App) renderTabBar() string {
	var tabs []string
	for i, name := range screenNames {
		label := fmt.Sprintf(" %s ", name)
		if screen(i) == a.screen {
			tabs = append(tabs, tabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(label))
		}
	}
	bar := strings.Join(tabs, dimStyle.Render("│"))
	return bar + dimStyle.Render("  [tab: switch  q: quit  v: cycle vis]")
}

func (a *App) renderNowPlaying() string {
	if a.nowPlaying == "" {
		return dimStyle.Render("  Not playing")
	}

	state := " > "
	if a.isPaused {
		state = " || "
	}

	vol := a.player.Volume()
	volStr := fmt.Sprintf("vol: %+.0fdB", vol)

	title := a.nowPlaying
	if a.streamTitle != "" && a.streamTitle != a.nowPlaying {
		title = a.nowPlaying + "  " + dimStyle.Render("~ "+a.streamTitle)
	}

	line1 := greenStyle.Render(state) + headerStyle.Render(title)
	line2 := dimStyle.Render("  " + volStr + "  vis: " + a.visualizer.ModeName() + "  [space: pause  s: stop  +/-: vol  v: vis]")
	return line1 + "\n" + line2
}

func (a *App) renderStatusBar() string {
	if a.statusMsg != "" && time.Now().Before(a.statusExpiry) {
		return errorStyle.Render("  " + a.statusMsg)
	}
	return dimStyle.Render("  spore – web radio & local player")
}

// PlayStation starts playing a radio station.
func (a *App) PlayStation(s radio.Station) tea.Cmd {
	return func() tea.Msg {
		if err := a.player.Play(s.URL, 0); err != nil {
			return errMsg{err}
		}
		if s.StationUUID != "" {
			go radio.Click(context.Background(), s.StationUUID)
		}
		return playStartedMsg{name: s.Name, url: s.URL, uuid: s.StationUUID, radio: true}
	}
}

// PlayFile starts playing a local file.
func (a *App) PlayFile(path, name string) tea.Cmd {
	return func() tea.Msg {
		if err := a.player.Play(path, 0); err != nil {
			return errMsg{err}
		}
		return playStartedMsg{name: name, url: path, radio: false}
	}
}

// ToggleFavoriteStation adds or removes a station from favorites.
func (a *App) ToggleFavoriteStation(s radio.Station) {
	if storage.IsFavorite(a.favorites, s.StationUUID) {
		a.favorites = storage.RemoveFavorite(a.favorites, s.StationUUID)
	} else {
		a.favorites = storage.AddFavorite(a.favorites, storage.FavoriteStation{
			UUID:    s.StationUUID,
			Name:    s.Name,
			URL:     s.URL,
			Tags:    s.Tags,
			Country: s.Country,
			Bitrate: s.Bitrate,
			Codec:   s.Codec,
		})
	}
	go storage.SaveFavorites(a.favorites)
}

func (a *App) sessionSnapshot() storage.Session {
	return storage.Session{
		StationUUID: a.lastStationUUID,
		StationName: a.lastStationName,
		StationURL:  a.lastStationURL,
		VisMode:     a.visualizer.ModeName(),
		VisEnabled:  a.visEnabled,
	}
}

func (a *App) persistSession() {
	go storage.SaveSession(a.sessionSnapshot())
}

// seedSessionFromHistory fills in the last radio station from playback history
// when no session.json exists yet (first run of this feature).
func seedSessionFromHistory(sess storage.Session, hist []storage.HistoryEntry, favs []storage.FavoriteStation) storage.Session {
	for _, h := range hist {
		if h.IsLocal || h.URL == "" {
			continue
		}
		sess.StationURL = h.URL
		sess.StationName = h.Name
		sess.StationUUID = h.UUID
		for _, f := range favs {
			if f.URL == h.URL || (h.UUID != "" && f.UUID == h.UUID) {
				sess.StationUUID = f.UUID
				if sess.StationName == "" {
					sess.StationName = f.Name
				}
				break
			}
		}
		break
	}
	return sess
}
