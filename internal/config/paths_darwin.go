package config

import (
	"os"
	"path/filepath"
)

// ~/Library/Application Support/Leneki
func dataDir() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, appDir), nil
}

// ~/Library/Caches/Leneki
func cacheDir() (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, appDir), nil
}
