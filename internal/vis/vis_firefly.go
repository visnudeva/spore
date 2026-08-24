package vis

import (
	"math"
	"strings"
)

// bandAvg returns the mean of bands[lo:hi], guarded against out-of-range
// arguments. Shared by visualizers that consume "bass / mid / high" subbands
// rather than the full array.
func bandAvg(b []float64, lo, hi int) float64 {
	if lo < 0 {
		lo = 0
	}
	if hi > len(b) {
		hi = len(b)
	}
	if hi <= lo {
		return 0
	}
	s := 0.0
	for _, x := range b[lo:hi] {
		s += x
	}
	return s / float64(hi-lo)
}

// renderFirefly draws a meadow at dusk: a low silhouette of grass at the
// bottom and many fireflies drifting above. Each firefly traces a slow Lissajous
// curve seeded per index so they never collide rigidly. High-frequency energy
// raises the population's brightness and the chance any given firefly is "lit"
// this frame; bass tilts a gentle wind that nudges them sideways.
func (v *Visualizer) renderFirefly(bands []float64) string {
	height := v.Rows
	dotRows := height * 4
	dotCols := PanelWidth * 2
	if dotRows < 4 || dotCols < 8 {
		return strings.Repeat("\n", max(0, height-1))
	}

	bass := bandAvg(bands, 0, len(bands)/3)
	high := bandAvg(bands, 2*len(bands)/3, len(bands))

	// Grass silhouette: bottom 1-2 rows, ragged edge.
	grass := make([]bool, dotRows*dotCols)
	for x := 0; x < dotCols; x++ {
		// Pseudo-noise heights.
		h := 1 + int(2.5+1.5*math.Sin(float64(x)*0.41)+1.0*math.Sin(float64(x)*0.17+2.3))
		for d := 0; d < h; d++ {
			y := dotRows - 1 - d
			if y >= 0 {
				grass[y*dotCols+x] = true
			}
		}
	}

	// Fireflies.
	const numFlies = 26
	wind := bass * 1.5

	dim := make([]bool, dotRows*dotCols)
	bright := make([]bool, dotRows*dotCols)

	for i := 0; i < numFlies; i++ {
		seed := uint64(i)*2246822519 + 11
		// Two slightly incommensurate frequencies for Lissajous-like wandering.
		fx := 0.012 + float64(seed%17)/3500.0
		fy := 0.018 + float64((seed>>4)%19)/2900.0
		phx := float64(seed%1000) / 1000.0 * 2 * math.Pi
		phy := float64((seed>>8)%1000) / 1000.0 * 2 * math.Pi

		t := float64(v.frame)
		baseX := float64(dotCols/2) + math.Cos(t*fx+phx)*float64(dotCols-6)*0.45
		baseY := float64(dotRows-4)*0.5 + math.Sin(t*fy+phy)*float64(dotRows-6)*0.4
		x := int(baseX + wind*math.Sin(t*0.02+phx))
		y := int(baseY)
		if x < 0 || x >= dotCols || y < 0 || y >= dotRows-1 {
			continue
		}
		// Skip if it would land in the grass silhouette.
		if grass[y*dotCols+x] {
			continue
		}

		// Blink: chance of being "on" depends on per-fly phase plus high band.
		blinkPhase := math.Sin(t*0.18+float64(i)*1.31) * 0.5
		on := blinkPhase+0.5+high*0.4 > 0.55
		if !on {
			// Half-brightness halo so the fly is faintly there.
			dim[y*dotCols+x] = true
			continue
		}

		bright[y*dotCols+x] = true
		// Glow halo (one-dot ring).
		for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			gx := x + d[1]
			gy := y + d[0]
			if gx >= 0 && gx < dotCols && gy >= 0 && gy < dotRows && !grass[gy*dotCols+gx] {
				dim[gy*dotCols+gx] = true
			}
		}
	}

	lines := make([]string, height)
	for row := 0; row < height; row++ {
		var sb, run strings.Builder
		tag := -1
		for col := 0; col < PanelWidth; col++ {
			var braille rune = '⠀'
			cellTag := -1
			for dr := 0; dr < 4; dr++ {
				for dc := 0; dc < 2; dc++ {
					y := row*4 + dr
					x := col*2 + dc
					idx := y*dotCols + x
					on := false
					t := 0
					switch {
					case bright[idx]:
						on = true
						t = 2
					case dim[idx]:
						on = true
						t = 1
					case grass[idx]:
						on = true
						t = 0
					}
					if on {
						braille |= brailleBit[dr][dc]
						if t > cellTag {
							cellTag = t
						}
					}
				}
			}
			if cellTag < 0 {
				cellTag = 0
			}
			if cellTag != tag {
				flushStyleRun(&sb, &run, tag)
				tag = cellTag
			}
			run.WriteRune(braille)
		}
		flushStyleRun(&sb, &run, tag)
		lines[row] = sb.String()
	}
	return strings.Join(lines, "\n")
}
