// Package player provides the audio engine for MP3 playback with
// a 10-band parametric EQ, volume control, and sample capture for visualization.
package audio

import (
	"sync/atomic"

	"github.com/gopxl/beep/v2"
)

// tap is a streamer wrapper that copies samples into a ring buffer
// for real-time FFT visualization. It sits in the audio pipeline
// before the volume control so the visualizer sees pre-volume amplitude.
//
// The write position is updated atomically, allowing the audio thread
// (sole writer) and the visualizer thread to operate without mutex contention.
// Minor sample tearing at the read boundary is invisible in visualization.
type tap struct {
	s    beep.Streamer
	buf  [][2]float64
	pos  atomic.Int64
	size int
}

// newTap wraps a streamer with a ring buffer of the given size.
func newTap(s beep.Streamer, bufSize int) *tap {
	return &tap{
		s:    s,
		buf:  make([][2]float64, bufSize),
		size: bufSize,
	}
}

// Stream passes audio through while capturing stereo frames in the ring buffer.
func (t *tap) Stream(samples [][2]float64) (int, bool) {
	n, ok := t.s.Stream(samples)
	p := int(t.pos.Load())
	for i := range n {
		t.buf[p] = samples[i]
		p = (p + 1) % t.size
	}
	t.pos.Store(int64(p))
	return n, ok
}

// Err returns the underlying streamer's error.
func (t *tap) Err() error {
	return t.s.Err()
}

// SamplesInto copies a mono mix of the last len(dst) frames into dst without
// allocating or using per-sample modulo in the ring-buffer read loop.
func (t *tap) SamplesInto(dst []float64) int {
	n := min(len(dst), t.size)
	p := int(t.pos.Load())
	start := (p - n + t.size) % t.size
	first := min(n, t.size-start)
	for i, sample := range t.buf[start : start+first] {
		dst[i] = (sample[0] + sample[1]) / 2
	}
	for i, sample := range t.buf[:n-first] {
		dst[first+i] = (sample[0] + sample[1]) / 2
	}
	return n
}

// StereoSamplesInto copies the last len(dst) stereo frames into dst.
func (t *tap) StereoSamplesInto(dst [][2]float64) int {
	n := min(len(dst), t.size)
	p := int(t.pos.Load())
	start := (p - n + t.size) % t.size
	first := min(n, t.size-start)
	copy(dst[:first], t.buf[start:start+first])
	copy(dst[first:n], t.buf[:n-first])
	return n
}
