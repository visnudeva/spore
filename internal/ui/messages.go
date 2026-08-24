package ui

import "github.com/spore-player/spore/internal/radio"

// Internal Bubbletea messages.

type tickMsg struct{}
type visTick struct{}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type stationsLoadedMsg struct {
	stations []radio.Station
	err      error
}

type playStartedMsg struct {
	name string
	url  string
	uuid string // radio station UUID; empty for local files
	radio bool
}

type streamTitleMsg struct {
	title string
}

type favoritesChangedMsg struct{}
