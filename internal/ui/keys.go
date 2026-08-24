package ui

// Key help entries used across screens.
var (
	globalKeys = []keyHint{
		{"tab", "switch screen"},
		{"q / ctrl+c", "quit"},
	}

	playerKeys = []keyHint{
		{"space", "pause/resume"},
		{"s", "stop"},
		{"+/-", "volume"},
		{"m", "mute"},
		{"v", "cycle visualizer"},
		{"r", "reload station"},
		{"f", "toggle favorite"},
		{"tab", "switch screen"},
		{"q", "quit"},
	}

	radioKeys = []keyHint{
		{"enter", "play selected"},
		{"/", "search"},
		{"t", "browse tags"},
		{"↑↓", "navigate"},
		{"f", "toggle favorite"},
		{"tab", "switch screen"},
		{"q", "quit"},
	}

	localKeys = []keyHint{
		{"enter", "open/play"},
		{"backspace", "go up"},
		{"↑↓", "navigate"},
		{"tab", "switch screen"},
		{"q", "quit"},
	}

	searchKeys = []keyHint{
		{"enter", "submit"},
		{"esc", "cancel"},
		{"↑↓", "navigate results"},
	}
)

type keyHint struct {
	key  string
	desc string
}
