package vis

import (
	"math"
	"strings"
	"time"
)

const lavaBlobCount = 10

type lavaBlob struct {
	x, y, r float64
	vy      float64
	phase   float64
	band    int
}

// lavaDriver is a lava-lamp: soft blobs rise, morph, and fall with the music.
type lavaDriver struct {
	blobs []lavaBlob
	rng   uint64
}

func newLavaDriver() visModeDriver {
	return &lavaDriver{rng: 0x1A8A001}
}

func (*lavaDriver) AnalysisSpec(*Visualizer) VisAnalysisSpec {
	return spectrumAnalysisSpec(DefaultSpectrumBands)
}

func (d *lavaDriver) rand01() float64 {
	d.rng = d.rng*6364136223846793005 + 1442695040888963407
	return float64(d.rng>>33) / float64(1<<31)
}

func (d *lavaDriver) ensure(dotCols, dotRows, bands int) {
	if len(d.blobs) == lavaBlobCount {
		return
	}
	d.blobs = make([]lavaBlob, lavaBlobCount)
	for i := range d.blobs {
		d.blobs[i] = lavaBlob{
			x:     d.rand01() * float64(dotCols),
			y:     d.rand01() * float64(dotRows),
			r:     2 + d.rand01()*5,
			vy:    -(0.15 + d.rand01()*0.35),
			phase: d.rand01() * twoPiApprox,
			band:  i % max(1, bands),
		}
	}
}

func (d *lavaDriver) Tick(v *Visualizer, ctx VisTickContext) {
	defaultDriverTick(v, ctx, d.AnalysisSpec(v))
	bands := v.SmoothedBands()
	n := len(bands)
	dotCols := PanelWidth * 2
	dotRows := v.Rows * 4
	d.ensure(dotCols, dotRows, max(1, n))

	var avg float64
	for _, e := range bands {
		avg += e
	}
	if n > 0 {
		avg /= float64(n)
	}

	for i := range d.blobs {
		b := &d.blobs[i]
		energy := 0.3
		if n > 0 {
			energy = bands[b.band%n]
		}
		b.phase += 0.04 + energy*0.08
		b.r = 2.5 + energy*7 + math.Sin(b.phase)*1.5
		b.x += math.Sin(b.phase*0.7) * (0.2 + energy*0.5)
		b.y += b.vy * (0.6 + avg + energy)
		// Bounce top/bottom like lava lamp.
		if b.y < b.r {
			b.y = b.r
			b.vy = math.Abs(b.vy)
		}
		if b.y > float64(dotRows)-b.r {
			b.y = float64(dotRows) - b.r
			b.vy = -math.Abs(b.vy)
		}
		if b.x < 0 {
			b.x = 0
		}
		if b.x > float64(dotCols) {
			b.x = float64(dotCols)
		}
	}
}

func (d *lavaDriver) TickInterval(_ *Visualizer, ctx VisTickContext) time.Duration {
	return defaultDriverTickInterval(ctx)
}

func (d *lavaDriver) OnEnter(*Visualizer) {
	d.blobs = nil
	d.rng = 0x1A8A001
}

func (*lavaDriver) OnLeave(*Visualizer) {}

func (d *lavaDriver) Render(v *Visualizer) string {
	height := v.Rows
	width := PanelWidth
	dotRows := height * 4
	dotCols := width * 2
	bands := v.SmoothedBands()
	d.ensure(dotCols, dotRows, max(1, len(bands)))

	lines := make([]string, height)
	for row := range height {
		var sb, run strings.Builder
		tag := -1
		for col := range width {
			var braille rune = '\u2800'
			var maxNorm float64
			for dr := range 4 {
				for dc := range 2 {
					px := float64(col*2 + dc)
					py := float64(row*4 + dr)
					var field float64
					for _, b := range d.blobs {
						dx := px - b.x
						dy := (py - b.y) * 0.85
						// Metaball falloff.
						d2 := dx*dx + dy*dy
						r2 := b.r * b.r
						if d2 < r2*4 {
							field += r2 / (d2 + 0.5)
						}
					}
					// Threshold metaballs into soft blobs.
					if field < 1.1 {
						continue
					}
					glow := math.Min(1, (field-1.1)/2.5)
					if glow < 0.2 && scatterHash(50, row*4+dr, col*2+dc, v.frame) > glow*3 {
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
