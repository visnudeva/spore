// Package local provides a local filesystem music browser.
package local

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SupportedExts lists audio file extensions spore can play.
var SupportedExts = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".ogg":  true,
	".m4a":  true,
	".aac":  true,
	".opus": true,
	".webm": true,
	".m4b":  true,
}

// Entry represents a file or directory in the browser.
type Entry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

// List returns the contents of dir, showing only audio files and directories.
// It returns a parent-dir entry ("..") first when dir is not the root.
func List(dir string, root string) ([]Entry, error) {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs, files []Entry

	if dir != root {
		parent := filepath.Dir(dir)
		dirs = append(dirs, Entry{Name: "..", Path: parent, IsDir: true})
	}

	for _, info := range infos {
		if strings.HasPrefix(info.Name(), ".") {
			continue // skip hidden
		}
		fi, err := info.Info()
		if err != nil {
			continue
		}
		if info.IsDir() {
			dirs = append(dirs, Entry{
				Name:  info.Name() + "/",
				Path:  filepath.Join(dir, info.Name()),
				IsDir: true,
			})
		} else if IsAudio(info.Name()) {
			files = append(files, Entry{
				Name: info.Name(),
				Path: filepath.Join(dir, info.Name()),
				Size: fi.Size(),
			})
		}
	}

	sort.Slice(dirs[1:], func(i, j int) bool {
		return strings.ToLower(dirs[1:][i].Name) < strings.ToLower(dirs[1:][j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return append(dirs, files...), nil
}

// IsAudio reports whether a filename has a supported audio extension.
func IsAudio(name string) bool {
	return SupportedExts[strings.ToLower(filepath.Ext(name))]
}

// HomeDir returns the user's home directory.
func HomeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "/"
}

// Walk recursively collects all audio files under root.
func Walk(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && IsAudio(d.Name()) {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
