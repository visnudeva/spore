//go:build linux

package audio

import (
	"os"
	"path/filepath"
	"strconv"
)

// withSharedALSA runs fn with a process-scoped ALSA config that routes the
// "default" PCM through PipeWire or PulseAudio.
//
// oto (via beep) opens ALSA "default" first. On many Linux setups that name
// maps to the raw hardware device (plughw), which spore then holds exclusively.
// Other apps lose output and system volume appears broken because PipeWire
// falls back to a dummy sink. Routing through the sound server keeps mixing
// and volume control working.
//
// The override is applied only for the duration of fn (speaker.Init) by
// pointing XDG_CONFIG_HOME at a temporary directory whose asoundrc is loaded
// last by alsa-lib. Spore's own config is not read during that window.
func withSharedALSA(fn func() error) error {
	conf := sharedALSAConfig()
	if conf == "" {
		return fn()
	}

	dir, err := os.MkdirTemp("", "spore-alsa-*")
	if err != nil {
		return fn()
	}
	defer os.RemoveAll(dir)

	alsaDir := filepath.Join(dir, "alsa")
	if err := os.MkdirAll(alsaDir, 0o700); err != nil {
		return fn()
	}
	if err := os.WriteFile(filepath.Join(alsaDir, "asoundrc"), []byte(conf), 0o600); err != nil {
		return fn()
	}

	prev, had := os.LookupEnv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", dir); err != nil {
		return fn()
	}
	defer func() {
		if had {
			_ = os.Setenv("XDG_CONFIG_HOME", prev)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	return fn()
}

func sharedALSAConfig() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = filepath.Join("/run/user", strconv.Itoa(os.Getuid()))
	}

	// Prefer PipeWire when its socket is present (covers pipewire-pulse too).
	if fileExists(filepath.Join(runtimeDir, "pipewire-0")) {
		return "pcm.!default {\n\ttype pipewire\n}\nctl.!default {\n\ttype pipewire\n}\n"
	}
	// Classic PulseAudio (or pipewire-pulse without the native socket name).
	if fileExists(filepath.Join(runtimeDir, "pulse", "native")) {
		return "pcm.!default {\n\ttype pulse\n}\nctl.!default {\n\ttype pulse\n}\n"
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
