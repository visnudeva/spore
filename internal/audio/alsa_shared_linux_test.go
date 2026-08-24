//go:build linux

package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWithSharedALSADoesNotOpenHardwarePCM(t *testing.T) {
	if sharedALSAConfig() == "" {
		t.Skip("no PipeWire/Pulse socket; exclusive ALSA is expected")
	}

	p, err := New(Quality{SampleRate: 44100, BufferMs: 100, ResampleQuality: 2, BitDepth: 16})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()

	// Give oto time to finish snd_pcm_open.
	time.Sleep(500 * time.Millisecond)

	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", e.Name()))
		if err != nil {
			continue
		}
		if strings.Contains(target, "/dev/snd/pcm") && !strings.Contains(target, "timer") {
			t.Fatalf("opened hardware PCM %s; expected PipeWire/Pulse mixing", target)
		}
	}
}
