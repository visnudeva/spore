package audio

// AudioDevice represents an available audio output device (sink/endpoint).
type AudioDevice struct {
	Index       int
	Name        string // internal identifier (sink name, UID, or device ID)
	Description string // human-readable label
	Active      bool   // true when this is the current default
}
