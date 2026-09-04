package config

import (
	"os"
	"path/filepath"
)

const appDir = "Leneki"

// Data survives across runs and is expensive to lose. Cache can be rebuilt.
type Paths struct {
	Data     string
	Cache    string
	DB       string
	Models   string
	Playback string
	Bin      string
	Temp     string
}

func Resolve() (Paths, error) {
	data, err := dataDir()
	if err != nil {
		return Paths{}, err
	}
	cache, err := cacheDir()
	if err != nil {
		return Paths{}, err
	}
	return New(data, cache), nil
}

// New derives the layout from the two roots, so tests can point it anywhere.
func New(data, cache string) Paths {
	return Paths{
		Data:     data,
		Cache:    cache,
		DB:       filepath.Join(data, "leneki.db"),
		Models:   filepath.Join(data, "models"),
		Playback: filepath.Join(data, "playback"),
		Bin:      filepath.Join(cache, "bin"),
		Temp:     filepath.Join(cache, "temp"),
	}
}

func (p Paths) Ensure() error {
	for _, d := range []string{p.Data, p.Cache, p.Models, p.Playback, p.Bin, p.Temp} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
