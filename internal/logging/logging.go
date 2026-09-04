package logging

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	FileName    = "leneki.log"
	maxLogBytes = 5 << 20
)

// Setup writes logs to a file in dir, and additionally to stderr when debug is
// set. The returned func closes the file.
func Setup(dir string, debug bool) (*slog.Logger, func() error, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	path := filepath.Join(dir, FileName)
	if err := rotateIfLarge(path); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}

	var w io.Writer = f
	level := slog.LevelInfo
	if debug {
		w = io.MultiWriter(f, os.Stderr)
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
	return logger, f.Close, nil
}

// The log is a support tool, not an archive, so one previous copy is enough.
func rotateIfLarge(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Size() < maxLogBytes {
		return nil
	}
	return os.Rename(path, path+".old")
}
