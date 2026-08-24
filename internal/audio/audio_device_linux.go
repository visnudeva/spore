//go:build linux

package audio

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ListAudioDevices returns available output sinks via pactl.
// Works on PulseAudio and PipeWire (via pipewire-pulse).
func ListAudioDevices() ([]AudioDevice, error) {
	defaultSink := ""
	if out, err := exec.Command("pactl", "get-default-sink").Output(); err == nil {
		defaultSink = strings.TrimSpace(string(out))
	}

	out, err := exec.Command("pactl", "list", "sinks").Output()
	if err != nil {
		return nil, fmt.Errorf("pactl: %w (is PulseAudio/PipeWire running?)", err)
	}

	var devices []AudioDevice
	var cur *AudioDevice

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Sink #") {
			idx, _ := strconv.Atoi(strings.TrimPrefix(line, "Sink #"))
			devices = append(devices, AudioDevice{Index: idx})
			cur = &devices[len(devices)-1]
		} else if cur != nil {
			if key, val, ok := strings.Cut(line, ": "); ok {
				switch key {
				case "Name":
					cur.Name = val
					cur.Active = val == defaultSink
				case "Description":
					cur.Description = val
				}
			}
		}
	}

	return devices, nil
}

// PrepareAudioDevice sets PIPEWIRE_NODE so the PipeWire ALSA plugin
// routes this process's audio to the named device.
// Must be called before player.New(). Returns a cleanup that restores the env.
func PrepareAudioDevice(device string) func() {
	prev, hadPrev := os.LookupEnv("PIPEWIRE_NODE")
	os.Setenv("PIPEWIRE_NODE", device)
	return func() {
		if hadPrev {
			os.Setenv("PIPEWIRE_NODE", prev)
		} else {
			os.Unsetenv("PIPEWIRE_NODE")
		}
	}
}

// SwitchAudioDevice moves cliamp's audio stream to a different sink.
// Falls back to changing the system default if the stream can't be found.
func SwitchAudioDevice(deviceName string) error {
	out, err := exec.Command("pactl", "list", "sink-inputs").Output()
	if err != nil {
		return fmt.Errorf("pactl: %w", err)
	}

	pidStr := strconv.Itoa(os.Getpid())
	sinkInputIdx := -1
	currentIdx := 0
	props := map[string]string{}

	// Parse all sink-inputs, collecting properties per entry.
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Sink Input #") {
			// Check previous entry before moving on.
			if sinkInputIdx < 0 {
				sinkInputIdx = matchCliamp(props, pidStr, currentIdx)
			}
			if sinkInputIdx >= 0 {
				break
			}
			currentIdx, _ = strconv.Atoi(strings.TrimPrefix(line, "Sink Input #"))
			props = map[string]string{}
			continue
		}
		if key, val, ok := strings.Cut(line, "="); ok {
			props[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(val), `"`)
		}
	}
	// Check the last entry.
	if sinkInputIdx < 0 {
		sinkInputIdx = matchCliamp(props, pidStr, currentIdx)
	}

	if sinkInputIdx >= 0 {
		cmd := exec.Command("pactl", "move-sink-input",
			strconv.Itoa(sinkInputIdx), deviceName)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("move-sink-input: %s (%w)", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	// Fallback: change the system default sink.
	if out, err := exec.Command("pactl", "set-default-sink", deviceName).CombinedOutput(); err != nil {
		return fmt.Errorf("set-default-sink: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// matchCliamp checks if a sink-input's properties belong to cliamp.
func matchCliamp(props map[string]string, pidStr string, idx int) int {
	if props["application.process.id"] == pidStr {
		return idx
	}
	if strings.EqualFold(props["application.process.binary"], "cliamp") {
		return idx
	}
	if strings.Contains(strings.ToLower(props["application.name"]), "cliamp") {
		return idx
	}
	return -1
}
