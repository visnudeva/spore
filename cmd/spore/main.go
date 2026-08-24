// spore – a terminal web radio and local music player with FFT visualizer.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/spore-player/spore/internal/ui"
)

func main() {
	app, err := ui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "spore: init error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "spore: %v\n", err)
		os.Exit(1)
	}
}
