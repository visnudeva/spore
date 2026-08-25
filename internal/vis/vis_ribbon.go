package vis

import (
	"math"
	"strings"
)

// renderRibbon draws a twisting 3D-ish spectrum ribbon that snakes across the
// panel. Depth cueing dims the underside; energy widens and brightens the band.
func (v *Visualizer) renderRibbon(bands []float64) string {
	height := v.Rows
	width := PanelWidth
	dotRows := height * 4
	dotCols := width * 2
	n := len(bands)
	if n == 0 {
		return strings.Repeat("\n", max(0, height-1))
	}

	var avg float64
	for _, e := range bands {
		avg += e
	}
	avg /= float64(n)

	t := float64(v.frame) * (0.04 + avg*0.06)
	grid := make([]int8, dotRows*dotCols)

	steps := dotCols * 2
	for s := 0; s < steps; s++ {
		u := float64(s) / float64(steps-1) // 0..1 along ribbon
		bi := int(u * float64(n-1))
		frac := u*float64(n-1) - float64(bi)
		energy := bands[bi]
		if bi+1 < n {
			energy = bands[bi]*(1-frac) + bands[bi+1]*frac
		}

		// Path through space with twist.
		x := u * float64(dotCols-1)
		y := float64(dotRows)/2 +
			math.Sin(u*twoPiApprox*1.5+t)*float64(dotRows)*0.28 +
			math.Sin(u*twoPiApprox*3.2+t*1.4)*float64(dotRows)*0.1

		// Twist angle for ribbon width orientation.
		twist := u*twoPiApprox*2 + t*1.2
		halfW := 1.2 + energy*5 + avg*2
		depth := math.Cos(twist) // -1..1 facing

		for w := -halfW; w <= halfW; w += 0.5 {
			px := x + math.Cos(twist)*w*0.15
			py := y + math.Sin(twist)*w
			xi, yi := int(px), int(py)
			if xi < 0 || xi >= dotCols || yi < 0 || yi >= dotRows {
				continue
			}
			edge := 1 - math.Abs(w)/halfW
			tier := int8(1)
			if depth > 0 && edge > 0.3 {
				tier = 2
			}
			if energy > 0.5 && edge > 0.5 && depth > 0.2 {
				tier = 3
			}
			if depth < -0.3 {
				tier = 1
			}
			i := yi*dotCols + xi
			if tier > grid[i] {
				grid[i] = tier
			}
		}
	}

	lines := make([]string, height)
	for row := range height {
		var sb, run strings.Builder
		tag := -1
		for col := range width {
			var braille rune = '\u2800'
			var maxTier int8
			for dr := range 4 {
				for dc := range 2 {
					t := grid[(row*4+dr)*dotCols+col*2+dc]
					if t > 0 {
						braille |= brailleBit[dr][dc]
						if t > maxTier {
							maxTier = t
						}
					}
				}
			}
			newTag := int(maxTier) - 1
			if maxTier == 0 {
				newTag = -1
			}
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
