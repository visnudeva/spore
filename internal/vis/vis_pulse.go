package vis

import "math"

// pulseCoords holds per-dot distance-from-center and angle values for the
// current panel dimensions. Recomputed lazily on resize so polar visualizers
// can skip repeated sqrt/atan2 work and read from flat arrays instead.
type pulseCoords struct {
	width, height int
	maxR          float64
	dist          []float64
	angle         []float64
}

func (v *Visualizer) pulseCoords() *pulseCoords {
	height := v.Rows
	width := PanelWidth
	if c := v.pulseCoordCache; c != nil && c.width == width && c.height == height {
		return c
	}
	dotRows := height * 4
	dotCols := width * 2
	centerX := float64(dotCols) / 2.0
	centerY := float64(dotRows) / 2.0
	xScale := centerY / centerX

	size := height * width * 8
	c := &pulseCoords{
		width:  width,
		height: height,
		maxR:   centerY - 1,
		dist:   make([]float64, size),
		angle:  make([]float64, size),
	}
	for row := range height {
		for col := range width {
			for dr := range 4 {
				for dc := range 2 {
					dx := (float64(col*2+dc) - centerX) * xScale
					dy := float64(row*4+dr) - centerY
					idx := pulseDotIndex(row, col, dr, dc, width)
					c.dist[idx] = math.Sqrt(dx*dx + dy*dy)
					a := math.Atan2(dy, dx)
					if a < 0 {
						a += 2 * math.Pi
					}
					c.angle[idx] = a
				}
			}
		}
	}
	v.pulseCoordCache = c
	return c
}

func pulseDotIndex(row, col, dr, dc, width int) int {
	return ((row*width+col)*4+dr)*2 + dc
}
