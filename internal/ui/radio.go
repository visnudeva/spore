package ui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/spore-player/spore/internal/radio"
	"github.com/spore-player/spore/internal/storage"
)

// RadioModel is a search-first radio screen. Nothing is loaded until the user
// types a query.
type RadioModel struct {
	app      *App
	width    int
	height   int
	stations []radio.Station
	cursor   int
	search   string
	typing   bool
	loading  bool
	err      error
}

func newRadioModel(app *App) RadioModel {
	return RadioModel{app: app, typing: true}
}

func (m *RadioModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case stationsLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.stations = msg.stations
			m.cursor = 0
			m.typing = false
		}
		return nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return nil
}

func (m *RadioModel) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	if m.typing {
		return m.handleSearchInput(msg)
	}

	switch msg.String() {
	case "/":
		m.typing = true
		return nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.stations)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(m.stations) {
			return m.app.PlayStation(m.stations[m.cursor])
		}
	case "f":
		if m.cursor < len(m.stations) {
			m.app.ToggleFavoriteStation(m.stations[m.cursor])
		}
	}
	return nil
}

func (m *RadioModel) handleSearchInput(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.typing = false
	case "enter":
		query := strings.TrimSpace(m.search)
		if query == "" {
			m.typing = false
			return nil
		}
		m.loading = true
		m.err = nil
		m.cursor = 0
		return func() tea.Msg {
			stations, err := radio.Search(context.Background(), radio.SearchParams{
				Name:  query,
				Limit: 40,
			})
			return stationsLoadedMsg{stations: stations, err: err}
		}
	case "backspace":
		if len(m.search) > 0 {
			r := []rune(m.search)
			m.search = string(r[:len(r)-1])
		}
	case " ", "space":
		m.search += " "
	default:
		s := msg.String()
		if len(s) == 1 {
			m.search += s
		}
	}
	return nil
}

func (m *RadioModel) view() string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Radio"))
	sb.WriteString(dimStyle.Render("  [type to search  enter play  / new search  f fav]\n"))

	prompt := m.search
	if m.typing {
		prompt += "_"
	}
	sb.WriteString(greenStyle.Render("  Search: ") + prompt + "\n")

	if m.loading {
		sb.WriteString(dimStyle.Render("  Searching...\n"))
		return sb.String()
	}
	if m.err != nil {
		sb.WriteString(errorStyle.Render("  Error: " + m.err.Error() + "\n"))
		return sb.String()
	}
	if len(m.stations) == 0 {
		sb.WriteString(dimStyle.Render("\n  Type a station name, city, or genre and press enter.\n"))
		return sb.String()
	}

	visible := m.height - 4
	if visible < 1 {
		visible = 10
	}
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	for i := start; i < len(m.stations) && i < start+visible; i++ {
		s := m.stations[i]
		cursor := "  "
		style := itemNormal
		if i == m.cursor && !m.typing {
			cursor = "> "
			style = itemSelected
		}
		starStr := "  "
		if storage.IsFavorite(m.app.favorites, s.StationUUID) {
			starStr = favStar.Render("* ")
		}
		name := truncate(s.Name, 36)
		meta := fmt.Sprintf("%s  %dkbps  %s", s.Codec, s.Bitrate, s.Country)
		sb.WriteString(cursor + starStr + style.Render(name) + "  " + itemDim.Render(meta) + "\n")
	}
	return sb.String()
}

var (
	itemSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(10))
	itemNormal   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(15))
	itemDim      = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(7))
	favStar      = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(11))
)

func truncate(s string, max int) string {
	if len([]rune(s)) <= max {
		return s
	}
	return string([]rune(s)[:max-1]) + "..."
}
