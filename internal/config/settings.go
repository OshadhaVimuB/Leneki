package config

import "runtime"

type Settings struct {
	SelectedModel string `json:"selectedModel"`
	ModelDir      string `json:"modelDir"`
	TempDir       string `json:"tempDir"`
	Threads       int    `json:"threads"`
	LastExportDir string `json:"lastExportDir"`
}

func Defaults(p Paths) Settings {
	return Settings{
		ModelDir: p.Models,
		TempDir:  p.Temp,
		Threads:  runtime.NumCPU(),
	}
}
