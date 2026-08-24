package appdir

import (
	"os"
	"path/filepath"
	"runtime"
)

// Dir returns the spore configuration directory.
func Dir() (string, error) {
	if dir, ok := os.LookupEnv("SPORE_CONFIG_DIR"); ok && dir != "" {
		return dir, nil
	}
	var dir string
	switch {
	case os.Getenv("XDG_CONFIG_HOME") != "":
		dir = filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "spore")
	case os.Getenv("HOME") != "":
		dir = filepath.Join(os.Getenv("HOME"), ".config", "spore")
	case runtime.GOOS == "windows" && os.Getenv("APPDATA") != "":
		dir = filepath.Join(os.Getenv("APPDATA"), "spore")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config", "spore")
	}
	migrateLegacyConfig(dir)
	return dir, nil
}

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

// PluginDir returns the spore plugin directory.
func PluginDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "plugins"), nil
}

// DataDir returns the spore data directory.
func DataDir() (string, error) {
	if home, ok := os.LookupEnv("HOME"); ok && home != "" {
		return filepath.Join(home, ".local", "share", "spore"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "spore"), nil
}
