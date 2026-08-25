package vis

import (
	"math"
	"strings"
)

// renderBloom draws soft expanding bloom / aura rings from the center.
// Loud onsets thicken the glow; quieter passages leave thin luminous shells.
func (v *Visualizer) renderBloom(bands []float64) string {
	coords := v.pulseCoords()
	height := v.Rows
	width := PanelWidth
	maxR := coords.maxR

	var avg, bass float64
	n := len(bands)
	if n > 0 {
		for i, e := range bands {
			avg += e
			if i < n/3 {
				bass += e
			}
		}
		avg /= float64(n)
		bass /= float64(max(1, n/3))
	}

	t := float64(v.frame) * (0.05 + avg*0.08)
	ringCount := 3 + int(avg*3)

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
					if dist > maxR+2 {
						continue
					}

					var glow float64
					for i := 0; i < ringCount; i++ {
						phase := math.Mod(t*0.35+float64(i)/float64(ringCount), 1.0)
						ringR := maxR * (0.15 + phase*0.85)
						thick := 1.2 + avg*2.5 + bass*1.5
						d := math.Abs(dist - ringR)
						if d < thick {
							g := (1 - d/thick) * (1 - phase*0.55)
							if g > glow {
								glow = g
							}
						}
					}

					// Soft core bloom.
					core := math.Max(0, 1-dist/(maxR*(0.2+avg*0.35)))
					glow = math.Max(glow, core*avg*0.9)

					if glow < 0.18 {
						continue
					}
					if glow < 0.35 && scatterHash(0, row*4+dr, col*2+dc, v.frame) > glow*2 {
						continue
					}

					braille |= brailleBit[dr][dc]
					if glow > maxNorm {
						maxNorm = glow
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
