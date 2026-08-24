package vis

import "math"

// fftInPlace runs a radix-2 Cooley-Tukey FFT on buf in place. len(buf) must be
// a power of 2. w must hold len(buf)/2 precomputed complex roots of unity
// where w[k] = exp(-2πi·k/n). Allocates nothing.
func fftInPlace(buf []complex128, w []complex128) {
	n := len(buf)
	if n < 2 {
		return
	}

	// Bit-reversal permutation: reorder buf so butterflies operate on in-order pairs.
	j := 0
	for i := 1; i < n; i++ {
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}
		j ^= bit
		if i < j {
			buf[i], buf[j] = buf[j], buf[i]
		}
	}

	// Butterfly stages using the shared twiddle table. At stage `size`, stride
	// into the n-sized root-of-unity table is n/size so we reuse w across sizes.
	for size := 2; size <= n; size <<= 1 {
		half := size >> 1
		step := n / size
		for start := 0; start < n; start += size {
			for k := 0; k < half; k++ {
				t := w[k*step] * buf[start+k+half]
				u := buf[start+k]
				buf[start+k] = u + t
				buf[start+k+half] = u - t
			}
		}
	}
}

// buildTwiddles returns the first n/2 complex nth-roots of unity used by
// fftInPlace for an n-point transform.
func buildTwiddles(n int) []complex128 {
	w := make([]complex128, n/2)
	for k := range w {
		angle := -2 * math.Pi * float64(k) / float64(n)
		w[k] = complex(math.Cos(angle), math.Sin(angle))
	}
	return w
}
