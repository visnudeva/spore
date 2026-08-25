package vis

import (
	"math"
	"strings"
)

// renderKaleido mirrors spectrum energy across six kaleidoscope sectors.
// Rotation and sector brightness track the music — classic MilkDrop kaleidoscope.
func (v *Visualizer) renderKaleido(bands []float64) string {
	coords := v.pulseCoords()
	height := v.Rows
	width := PanelWidth
	bandCount := len(bands)
	if bandCount == 0 {
		return strings.Repeat("\n", max(0, height-1))
	}
	maxR := coords.maxR

	var avg float64
	for _, e := range bands {
		avg += e
	}
	avg /= float64(bandCount)

	rot := float64(v.frame) * (0.018 + avg*0.04)
	sectors := 6
	secW := twoPiApprox / float64(sectors)

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
					if dist > maxR {
						continue
					}

					ang := coords.angle[idx] + rot
					ang = math.Mod(ang, twoPiApprox)
					if ang < 0 {
						ang += twoPiApprox
					}

					// Fold into one sector, then mirror inside it.
					local := math.Mod(ang, secW)
					if local > secW/2 {
						local = secW - local
					}
					u := local / (secW / 2)          // 0..1 across sector
					r := dist / max(0.001, maxR)     // 0..1 radial

					// Sample spectrum with polar coords.
					bandPos := (u*0.55 + r*0.45) * float64(bandCount-1)
					bi := int(bandPos)
					if bi >= bandCount-1 {
						bi = bandCount - 1
					}
					frac := bandPos - float64(bi)
					energy := bands[bi]
					if bi+1 < bandCount {
						energy = bands[bi]*(1-frac) + bands[bi+1]*frac
					}

					// Organic cell boundaries + energy fill.
					cell := math.Sin(u*9+r*7+rot*2)*0.5 + 0.5
					thresh := 0.55 - energy*0.5
					val := energy*0.7 + cell*0.3
					if val < thresh {
						continue
					}

					braille |= brailleBit[dr][dc]
					norm := r*0.5 + energy*0.5
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
