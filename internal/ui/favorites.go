package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/spore-player/spore/internal/radio"
	"github.com/spore-player/spore/internal/storage"
)

// FavModel is the favorites screen.
type FavModel struct {
	app    *App
	cursor int
}

func newFavModel(app *App) FavModel {
	return FavModel{app: app}
}

func (m *FavModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return nil
}

func (m *FavModel) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	favs := m.app.favorites
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(favs)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(favs) {
			s := favs[m.cursor]
			return m.app.PlayStation(radio.Station{
				StationUUID: s.UUID,
				Name:        s.Name,
				URL:         s.URL,
			})
		}
	case "d", "delete":
		if m.cursor < len(favs) {
			m.app.favorites = storage.RemoveFavorite(m.app.favorites, favs[m.cursor].UUID)
			go storage.SaveFavorites(m.app.favorites)
			if m.cursor >= len(m.app.favorites) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return nil
}

func (m *FavModel) view() string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Favorites") + "\n")
	sb.WriteString(dimStyle.Render("  [enter: play  d: remove  up/down navigate]\n"))
	favs := m.app.favorites
	if len(favs) == 0 {
		sb.WriteString(dimStyle.Render("\n  No favorites yet.\n  Press f while browsing radio to add stations.\n"))
		return sb.String()
	}
	for i, s := range favs {
		cursor := "  "
		style := itemNormal
		if i == m.cursor {
			cursor = "> "
			style = itemSelected
		}
		meta := fmt.Sprintf("%dkbps %s", s.Bitrate, s.Country)
		sb.WriteString(cursor + favStar.Render("* ") + style.Render(s.Name) + "  " + itemDim.Render(meta) + "\n")
	}
	return sb.String()
}
