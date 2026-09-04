package config

import (
	"os"
	"path/filepath"
)

// %APPDATA%\Leneki
func dataDir() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, appDir), nil
}

// %LOCALAPPDATA%\Leneki\Cache
func cacheDir() (string, error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, appDir, "Cache"), nil
}
