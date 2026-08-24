//go:build !linux

package audio

// withSharedALSA is a no-op outside Linux; oto does not use ALSA there.
func withSharedALSA(fn func() error) error {
	return fn()
}
