package config

import (
	"os"
	"path/filepath"
)

// $XDG_DATA_HOME/Leneki, falling back to ~/.local/share/Leneki. os.UserConfigDir
// is deliberately not used here: it returns ~/.config, which is for settings.
func dataDir() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, appDir), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", appDir), nil
}

// ~/.cache/Leneki
func cacheDir() (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, appDir), nil
}
