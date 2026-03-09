package tailer

import (
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/nxadm/tail"
)

// DiscoverFiles returns sorted matching files for one directory and pattern.
func DiscoverFiles(dir, pattern string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(matches))
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			continue
		}
		files = append(files, match)
	}

	slices.Sort(files)
	return files, nil
}

// InitialOffset decides where an existing file should start.
func InitialOffset(path string, fromStart bool) (int64, error) {
	if fromStart {
		return 0, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func tailLocation(path string, fromStart bool) (*tail.SeekInfo, error) {
	offset, err := InitialOffset(path, fromStart)
	if err != nil {
		return nil, err
	}

	return &tail.SeekInfo{
		Offset: offset,
		Whence: io.SeekStart,
	}, nil
}
