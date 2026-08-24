package vis

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// FitRect clips text to a terminal rectangle without splitting ANSI escapes or
// wide characters. It intentionally does not pad rows, so callers can compose
// compact layouts without introducing trailing whitespace.
func FitRect(text string, width, height int) string {
	if width <= 0 || height <= 0 || text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}
