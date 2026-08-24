package ui

import (
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/spore-player/spore/internal/local"
)

// LocalModel is the local filesystem browser screen.
type LocalModel struct {
	app     *App
	width   int
	height  int
	dir     string
	root    string
	entries []local.Entry
	cursor  int
	err     error
}

func newLocalModel(app *App) LocalModel {
	home := local.HomeDir()
	entries, err := local.List(home, home)
	return LocalModel{
		app:     app,
		dir:     home,
		root:    home,
		entries: entries,
		err:     err,
	}
}

func (m *LocalModel) update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return nil
}

func (m *LocalModel) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(m.entries) {
			e := m.entries[m.cursor]
			if e.IsDir {
				entries, err := local.List(e.Path, m.root)
				if err == nil {
					m.dir = e.Path
					m.entries = entries
					m.cursor = 0
					m.err = nil
				} else {
					m.err = err
				}
			} else {
				name := filepath.Base(e.Path)
				return m.app.PlayFile(e.Path, name)
			}
		}
	case "backspace":
		parent := filepath.Dir(m.dir)
		if parent != m.dir {
			entries, err := local.List(parent, m.root)
			if err == nil {
				m.dir = parent
				m.entries = entries
				m.cursor = 0
				m.err = nil
			}
		}
	}
	return nil
}

func (m *LocalModel) view() string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Files: "+m.dir) + "\n")
	sb.WriteString(dimStyle.Render("  [up/down navigate  enter open/play  backspace up]\n"))
	if m.err != nil {
		sb.WriteString(errorStyle.Render("  Error: " + m.err.Error() + "\n"))
		return sb.String()
	}
	visible := m.height - 3
	if visible < 1 {
		visible = 10
	}
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	for i := start; i < len(m.entries) && i < start+visible; i++ {
		e := m.entries[i]
		cursor := "  "
		style := itemNormal
		if i == m.cursor {
			cursor = "> "
			style = itemSelected
		}
		icon := "  "
		if e.IsDir {
			icon = "[D] "
		} else {
			icon = "    "
		}
		sb.WriteString(cursor + icon + style.Render(e.Name) + "\n")
	}
	return sb.String()
}
