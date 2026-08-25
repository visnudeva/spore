package vis

import (
	"math"
	"strings"
)

// renderIris draws a living eye: dilating pupil, fibrous iris, and specular
// highlight. Bass opens the pupil; highs sparkle the iris fibers.
func (v *Visualizer) renderIris(bands []float64) string {
	coords := v.pulseCoords()
	height := v.Rows
	width := PanelWidth
	maxR := coords.maxR
	n := len(bands)
	if n == 0 {
		return strings.Repeat("\n", max(0, height-1))
	}

	var avg, bass, high float64
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

	pupilR := maxR * (0.12 + bass*0.22 + avg*0.06)
	irisR := maxR * (0.55 + avg*0.2)
	t := float64(v.frame) * (0.015 + avg*0.02)
	twoPi := twoPiApprox

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
					ang := coords.angle[idx]

					var glow float64

					if dist < pupilR {
						// Dark pupil — only a tiny catchlight.
						cxOff := math.Cos(ang - 0.6)
						if dist < pupilR*0.45 && cxOff > 0.7 {
							glow = 0.9
						} else {
							continue
						}
					} else if dist < irisR {
						// Iris fibers radiate from pupil.
						bi := int(ang / twoPi * float64(n))
						if bi < 0 {
							bi += n
						}
						if bi >= n {
							bi = n - 1
						}
						energy := bands[bi]
						fiber := math.Abs(math.Sin(ang*18 + t*3 + energy*4))
						radial := (dist - pupilR) / max(0.001, irisR-pupilR)
						glow = fiber * (0.35 + energy*0.7) * (1 - radial*0.4)
						glow += high * 0.15 * fiber
						// Limbus (dark ring at iris edge).
						if radial > 0.85 {
							glow = math.Max(glow, 0.55)
						}
					} else if dist < irisR+1.5 {
						// Soft sclera rim.
						glow = 0.2 * (1 - (dist-irisR)/1.5)
					} else {
						continue
					}

					if glow < 0.14 {
						continue
					}
					if glow < 0.28 && scatterHash(31, row*4+dr, col*2+dc, v.frame) > glow*2.5 {
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
