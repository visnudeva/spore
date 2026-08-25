package vis

import (
	"math"
	"strings"
	"time"
)

const starfieldCount = 96

type starfieldStar struct {
	x, y, z float64
}

// starfieldDriver renders a classic flying-star warp field. Stars streak toward
// the viewer; bass and overall energy crank the warp speed like a MilkDrop
// hyperspace preset.
type starfieldDriver struct {
	stars []starfieldStar
	rng   uint64
}

func newStarfieldDriver() visModeDriver {
	return &starfieldDriver{rng: 0x57A4F1E1D5EED ^ 0xC0FFEE}
}

func (*starfieldDriver) AnalysisSpec(*Visualizer) VisAnalysisSpec {
	return spectrumAnalysisSpec(DefaultSpectrumBands)
}

func (d *starfieldDriver) rand01() float64 {
	d.rng = d.rng*6364136223846793005 + 1442695040888963407
	return float64(d.rng>>33) / float64(1<<31)
}

func (d *starfieldDriver) spawn() starfieldStar {
	// Uniform-ish disk in [-1,1], depth far away.
	ang := d.rand01() * 2 * math.Pi
	rad := math.Sqrt(d.rand01())
	return starfieldStar{
		x: math.Cos(ang) * rad,
		y: math.Sin(ang) * rad * 0.55, // squash for terminal cell aspect
		z: 0.55 + d.rand01()*0.45,
	}
}

func (d *starfieldDriver) ensure() {
	if len(d.stars) == starfieldCount {
		return
	}
	d.stars = make([]starfieldStar, starfieldCount)
	for i := range d.stars {
		d.stars[i] = d.spawn()
		d.stars[i].z = 0.08 + d.rand01()*0.92
	}
}

func (d *starfieldDriver) Tick(v *Visualizer, ctx VisTickContext) {
	defaultDriverTick(v, ctx, d.AnalysisSpec(v))
	d.ensure()

	bands := v.SmoothedBands()
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

	speed := 0.012 + avg*0.055 + bass*0.04
	for i := range d.stars {
		d.stars[i].z -= speed
		if d.stars[i].z < 0.04 {
			d.stars[i] = d.spawn()
		}
	}
}

func (d *starfieldDriver) TickInterval(_ *Visualizer, ctx VisTickContext) time.Duration {
	return defaultDriverTickInterval(ctx)
}

func (d *starfieldDriver) OnEnter(*Visualizer) {
	d.stars = nil
	d.ensure()
}

func (*starfieldDriver) OnLeave(*Visualizer) {}

func (d *starfieldDriver) Render(v *Visualizer) string {
	d.ensure()
	height := v.Rows
	width := PanelWidth
	dotRows := height * 4
	dotCols := width * 2
	cx := float64(dotCols) / 2
	cy := float64(dotRows) / 2
	// Focal length in dots — fill most of the panel at z≈0.2.
	focal := math.Min(cx, cy) * 0.95

	bands := v.SmoothedBands()
	var avg float64
	for _, e := range bands {
		avg += e
	}
	if len(bands) > 0 {
		avg /= float64(len(bands))
	}

	grid := make([]int8, dotRows*dotCols) // 0 empty, 1..3 brightness tier

	for _, s := range d.stars {
		z := s.z
		if z < 0.04 {
			continue
		}
		sx := int(cx + s.x/z*focal)
		sy := int(cy + s.y/z*focal)
		if sx < 0 || sx >= dotCols || sy < 0 || sy >= dotRows {
			continue
		}

		// Near stars are brighter / thicker; loud passages bloom trails.
		near := 1.0 - z
		tier := int8(1)
		if near > 0.45+avg*0.2 {
			tier = 2
		}
		if near > 0.7+avg*0.15 {
			tier = 3
		}
		grid[sy*dotCols+sx] = tier

		// Motion streak toward center when warping hard.
		if avg > 0.25 && z < 0.55 {
			px := int(cx + s.x/(z+0.08)*focal)
			py := int(cy + s.y/(z+0.08)*focal)
			plotLine(grid, dotCols, dotRows, sx, sy, px, py, max8(1, tier-1))
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
			newTag := int(maxTier) - 1 // 0,1,2 → green/yellow/red
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

func max8(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}

// plotLine stamps a short Bresenham streak into the tier grid.
func plotLine(grid []int8, cols, rows, x0, y0, x1, y1 int, tier int8) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if x0 >= 0 && x0 < cols && y0 >= 0 && y0 < rows {
			i := y0*cols + x0
			if grid[i] < tier {
				grid[i] = tier
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
