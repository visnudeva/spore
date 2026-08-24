package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"charm.land/lipgloss/v2"
)

// overlayOnVis composites UI text over a full-screen visualizer. Empty overlay
// rows (and the right-hand remainder of short rows) show the visualizer.
func overlayOnVis(background, foreground string, width, height int) string {
	bg := padLines(strings.Split(background, "\n"), width, height)
	fg := padLines(strings.Split(foreground, "\n"), width, height)

	var out strings.Builder
	out.Grow(height * (width + 16))
	for i := 0; i < height; i++ {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(overlayLine(bg[i], fg[i], width))
	}
	return out.String()
}

func padLines(lines []string, width, height int) []string {
	out := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(lines) {
			out[i] = lines[i]
		} else {
			out[i] = ""
		}
	}
	return out
}

func overlayLine(bg, fg string, width int) string {
	opaque := lipgloss.Width(strings.TrimRight(ansi.Strip(fg), " "))
	if opaque <= 0 {
		return ansi.Cut(bg, 0, width)
	}
	if opaque >= width {
		return ansi.Cut(fg, 0, width)
	}
	left := ansi.Cut(fg, 0, opaque)
	right := ansi.TruncateLeft(bg, opaque, "")
	right = ansi.Cut(right, 0, width-opaque)
	return left + "\x1b[0m" + right
}
