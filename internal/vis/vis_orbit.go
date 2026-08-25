package vis

import (
	"math"
	"strings"
)

// renderOrbit draws planets on glowing orbital paths. Band energy sets each
// body's radius and brightness; the whole system slowly precesses.
func (v *Visualizer) renderOrbit(bands []float64) string {
	coords := v.pulseCoords()
	height := v.Rows
	width := PanelWidth
	maxR := coords.maxR
	n := len(bands)
	if n == 0 {
		return strings.Repeat("\n", max(0, height-1))
	}

	var avg float64
	for _, e := range bands {
		avg += e
	}
	avg /= float64(n)

	t := float64(v.frame) * (0.025 + avg*0.04)
	bodies := min(n, 8)

	// Precompute body positions in pulse-space.
	type body struct {
		ang, rad, e float64
	}
	bs := make([]body, bodies)
	for i := 0; i < bodies; i++ {
		bi := i * n / bodies
		e := bands[bi]
		bs[i] = body{
			ang: t*(0.4+float64(i)*0.15) + float64(i)*0.9,
			rad: maxR * (0.2 + float64(i)/float64(bodies)*0.7) * (0.7 + e*0.45),
			e:   e,
		}
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
					ang := coords.angle[idx]

					var glow float64
					for _, b := range bs {
						// Orbital ring.
						ringD := math.Abs(dist - b.rad)
						if ringD < 0.55+b.e*0.4 {
							g := (1 - ringD/(0.55+b.e*0.4)) * (0.2 + b.e*0.35)
							glow = math.Max(glow, g)
						}
						// Planet body.
						dAng := ang - b.ang
						dAng = math.Mod(dAng+math.Pi*3, twoPiApprox) - math.Pi
						arc := math.Hypot(dAng*b.rad, dist-b.rad)
						planetR := 1.2 + b.e*3
						if arc < planetR {
							g := (1 - arc/planetR) * (0.55 + b.e*0.55)
							glow = math.Max(glow, g)
						}
					}
					// Soft sun core.
					if dist < maxR*(0.08+avg*0.06) {
						glow = math.Max(glow, 0.7+avg*0.3)
					}

					if glow < 0.14 {
						continue
					}
					if glow < 0.28 && scatterHash(53, row*4+dr, col*2+dc, v.frame) > glow*2.5 {
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
