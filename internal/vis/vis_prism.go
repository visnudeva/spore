package vis

import (
	"math"
	"strings"
)

// renderPrism draws angled spectral light shards — diagonal beams that fan
// and refract with the music, like light through cut glass.
func (v *Visualizer) renderPrism(bands []float64) string {
	height := v.Rows
	width := PanelWidth
	dotRows := height * 4
	dotCols := width * 2
	n := len(bands)
	if n == 0 {
		return strings.Repeat("\n", max(0, height-1))
	}

	var avg, high float64
	for i, e := range bands {
		avg += e
		if i >= 2*n/3 {
			high += e
		}
	}
	avg /= float64(n)
	high /= float64(max(1, n-2*n/3))

	t := float64(v.frame) * (0.02 + avg*0.04)
	shardCount := 6 + int(avg*6+high*4)

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

					var glow float64
					for s := 0; s < shardCount; s++ {
						bi := s * n / shardCount
						if bi >= n {
							bi = n - 1
						}
						energy := bands[bi]

						// Each shard is a diagonal band: y ≈ m*x + b
						m := math.Tan(-0.9 + float64(s)*0.22 + math.Sin(t+float64(s))*0.15)
						b := 0.15 + float64(s)/float64(shardCount)*0.7 + math.Sin(t*1.3+float64(s)*0.9)*0.08
						dist := math.Abs(y - (m*x + b))
						// Correct for slope length roughly.
						dist /= math.Sqrt(1 + m*m)

						thick := 0.012 + energy*0.045 + high*0.01
						if dist > thick {
							continue
						}
						g := (1 - dist/thick) * (0.3 + energy*0.95)
						// Soft length fade from a prism origin.
						originX := 0.15 + math.Sin(t*0.5)*0.05
						along := math.Hypot(x-originX, y-0.5)
						g *= 0.4 + 0.6*(1-along*0.7)
						if g > glow {
							glow = g
						}
					}

					if glow < 0.16 {
						continue
					}
					if glow < 0.3 && scatterHash(11, row*4+dr, col*2+dc, v.frame) > glow*2.5 {
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
