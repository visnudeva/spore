package vis

import (
	"math"
	"strings"
)

// renderRipple draws expanding concentric water ripples. Bass drops launch
// stronger wavefronts; treble adds fine secondary ripples — MilkDrop pond vibes.
func (v *Visualizer) renderRipple(bands []float64) string {
	coords := v.pulseCoords()
	height := v.Rows
	width := PanelWidth
	maxR := coords.maxR

	var avg, bass, high float64
	n := len(bands)
	if n > 0 {
		for i, e := range bands {
			avg += e
			switch {
			case i < n/3:
				bass += e
			case i >= 2*n/3:
				high += e
			}
		}
		avg /= float64(n)
		bass /= float64(max(1, n/3))
		high /= float64(max(1, n-2*n/3))
	}

	t := float64(v.frame) * (0.08 + avg*0.12 + bass*0.1)
	wavelength := 4.5 - avg*1.5
	if wavelength < 2.5 {
		wavelength = 2.5
	}

	lines := make([]string, height)
	for row := range height {
		var sb, run strings.Builder
		tag := -1

		for col := range width {
			var braille rune = '\u2800'
			var maxNorm float64

			for dr := range 4 {
				for dc := range 2 {
					idx := pulseDotIndex(row, col, dr, dc, width)
					dist := coords.dist[idx]

					// Primary expanding wavefronts.
					wave := math.Sin(dist/wavelength*twoPiApprox - t)
					// Secondary high-frequency shimmer.
					shimmer := math.Sin(dist*1.7 - t*1.6 + high*3) * 0.35 * high
					val := wave*0.65 + shimmer + bass*0.25

					// Crests only.
					if val < 0.35 {
						continue
					}
					fade := 1.0 - dist/(maxR+1)
					if fade < 0.05 {
						continue
					}
					strength := (val - 0.35) * fade
					if strength < 0.12 && scatterHash(0, row*4+dr, col*2+dc, v.frame) > strength*5 {
						continue
					}

					braille |= brailleBit[dr][dc]
					norm := 0.2 + strength*0.8
					if norm > maxNorm {
						maxNorm = norm
					}
				}
			}

			newTag := specTag(maxNorm)
			if newTag != tag {
				flushStyleRun(&sb, &run, tag)
				tag = newTag
			}
			run.WriteRune(braille)
		}
		flushStyleRun(&sb, &run, tag)
		lines[row] = sb.String()
	}
	return strings.Join(lines, "\n")
}

const twoPiApprox = 2 * math.Pi
