// Package storage manages favorites, playback history, and session state on disk.
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func configDir() (string, error) {
	if d, ok := os.LookupEnv("SPORE_CONFIG_DIR"); ok && d != "" {
		return d, nil
	}
	var dir string
	if xdg, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && xdg != "" {
		dir = filepath.Join(xdg, "spore")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config", "spore")
	}
	migrateLegacyConfig(dir)
	return dir, nil
}

// migrateLegacyConfig renames ~/.config/wavr to the spore config dir when the
// latter does not exist yet.
func migrateLegacyConfig(dir string) {
	if _, err := os.Stat(dir); err == nil {
		return
	}
	legacy := filepath.Join(filepath.Dir(dir), "wavr")
	if _, err := os.Stat(legacy); err != nil {
		return
	}
	_ = os.Rename(legacy, dir)
}

func dataDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// FavoriteStation is a saved radio station.
type FavoriteStation struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Tags     string `json:"tags,omitempty"`
	Country  string `json:"country,omitempty"`
	Bitrate  int    `json:"bitrate,omitempty"`
	Codec    string `json:"codec,omitempty"`
	AddedAt  string `json:"added_at"`
}

type favorites struct {
	Stations []FavoriteStation `json:"stations"`
}

func favoritesPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "favorites.json"), nil
}

// LoadFavorites returns the saved favorite stations.
func LoadFavorites() ([]FavoriteStation, error) {
	path, err := favoritesPath()
	if err != nil {
		return nil, err
	}
	var f favorites
	if err := readJSON(path, &f); err != nil {
		return nil, err
	}
	return f.Stations, nil
}

// SaveFavorites writes the favorites list to disk.
func SaveFavorites(stations []FavoriteStation) error {
	path, err := favoritesPath()
	if err != nil {
		return err
	}
	return writeJSON(path, favorites{Stations: stations})
}

// IsFavorite reports whether stationUUID is in the favorites list.
func IsFavorite(stations []FavoriteStation, uuid string) bool {
	for _, s := range stations {
		if s.UUID == uuid {
			return true
		}
	}
	return false
}

// AddFavorite appends a station to the list if not already present.
func AddFavorite(stations []FavoriteStation, s FavoriteStation) []FavoriteStation {
	for _, existing := range stations {
		if existing.UUID == s.UUID {
			return stations
		}
	}
	s.AddedAt = time.Now().Format(time.RFC3339)
	return append(stations, s)
}

// RemoveFavorite removes a station by UUID.
func RemoveFavorite(stations []FavoriteStation, uuid string) []FavoriteStation {
	out := stations[:0]
	for _, s := range stations {
		if s.UUID != uuid {
			out = append(out, s)
		}
	}
	return out
}

// HistoryEntry is a single playback history record.
type HistoryEntry struct {
	UUID      string `json:"uuid,omitempty"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	PlayedAt  string `json:"played_at"`
	IsLocal   bool   `json:"is_local,omitempty"`
}

type history struct {
	Entries []HistoryEntry `json:"entries"`
}

const maxHistory = 200

func historyPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

// LoadHistory returns recent playback history (newest first).
func LoadHistory() ([]HistoryEntry, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}
	var h history
	if err := readJSON(path, &h); err != nil {
		return nil, err
	}
	return h.Entries, nil
}

// AppendHistory prepends an entry and trims to maxHistory entries.
func AppendHistory(entries []HistoryEntry, e HistoryEntry) []HistoryEntry {
	e.PlayedAt = time.Now().Format(time.RFC3339)
	out := make([]HistoryEntry, 0, min(len(entries)+1, maxHistory))
	out = append(out, e)
	for _, h := range entries {
		if len(out) >= maxHistory {
			break
		}
		out = append(out, h)
	}
	return out
}

// SaveHistory writes history to disk.
func SaveHistory(entries []HistoryEntry) error {
	path, err := historyPath()
	if err != nil {
		return err
	}
	return writeJSON(path, history{Entries: entries})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Session holds the last radio station and visualizer choice so spore can
// resume them on the next launch.
type Session struct {
	StationUUID string `json:"station_uuid,omitempty"`
	StationName string `json:"station_name,omitempty"`
	StationURL  string `json:"station_url,omitempty"`
	VisMode     string `json:"vis_mode,omitempty"`
	VisEnabled  bool   `json:"vis_enabled"`
}

func sessionPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// LoadSession returns the saved session, or a zero Session if none exists.
func LoadSession() (Session, error) {
	path, err := sessionPath()
	if err != nil {
		return Session{}, err
	}
	var s Session
	if err := readJSON(path, &s); err != nil {
		return Session{}, err
	}
	return s, nil
}

// SaveSession writes the current session to disk.
func SaveSession(s Session) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	return writeJSON(path, s)
}
