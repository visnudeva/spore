package vis

import (
	"math"
	"strings"
)

// renderPlasma draws a classic multi-sine plasma field. Audio energy speeds the
// phase drift and lowers the ignition threshold so loud passages flood the
// screen with swirling color bands — a MilkDrop staple.
func (v *Visualizer) renderPlasma(bands []float64) string {
	height := v.Rows
	width := PanelWidth
	dotRows := height * 4
	dotCols := width * 2

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

	t := float64(v.frame) * (0.035 + avg*0.09 + bass*0.04)
	// Louder → more of the field lights up.
	thresh := 0.55 - avg*0.45 - high*0.1
	if thresh < 0.08 {
		thresh = 0.08
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
					x := float64(col*2+dc) / float64(max(1, dotCols-1))
					y := float64(row*4+dr) / float64(max(1, dotRows-1))

					// Classic plasma: sum of angled sines + a radial term.
					v1 := math.Sin(x*6.0 + t)
					v2 := math.Sin(y*7.0 + t*1.3)
					v3 := math.Sin((x+y)*4.5 + t*0.7)
					v4 := math.Sin(math.Hypot(x-0.5, y-0.5)*10.0 - t*1.1)
					v5 := math.Sin(x*3.0 - y*5.0 + t*0.9 + bass*2)
					raw := (v1 + v2 + v3 + v4 + v5) / 5.0 // [-1,1]
					val := raw*0.5 + 0.5                   // [0,1]

					if val < thresh {
						continue
					}
					// Soft dither at the threshold edge.
					edge := (val - thresh) / max(0.001, 1-thresh)
					if edge < 0.2 && scatterHash(0, row*4+dr, col*2+dc, v.frame) > edge*4 {
						continue
					}

					braille |= brailleBit[dr][dc]
					if val > maxNorm {
						maxNorm = val
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
