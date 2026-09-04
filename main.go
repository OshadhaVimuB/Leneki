package main

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/OshadhaVimuB/Leneki/internal/binaries"
	"github.com/OshadhaVimuB/Leneki/internal/config"
	"github.com/OshadhaVimuB/Leneki/internal/logging"
	"github.com/OshadhaVimuB/Leneki/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Leneki failed to start:", err)
		os.Exit(1)
	}
}

func run() error {
	paths, err := config.Resolve()
	if err != nil {
		return fmt.Errorf("resolving application directories: %w", err)
	}
	if err := paths.Ensure(); err != nil {
		return fmt.Errorf("creating application directories: %w", err)
	}

	logger, closeLog, err := logging.Setup(paths.Data, strings.Contains(Version, "dev"))
	if err != nil {
		return fmt.Errorf("opening the log file: %w", err)
	}
	defer closeLog()
	slog.SetDefault(logger)

	slog.Info("starting", "version", Version)
	slog.Info("paths resolved", "data", paths.Data, "cache", paths.Cache)

	bins, err := binaries.Ensure(paths.Bin)
	if err != nil {
		return fmt.Errorf("preparing the bundled tools: %w", err)
	}
	slog.Info("bundled tools ready", "dir", paths.Bin)

	db, err := store.Open(paths.DB)
	if err != nil {
		return fmt.Errorf("opening the database: %w", err)
	}
	defer db.Close()

	settings, err := db.Settings().Load(config.Defaults(paths))
	if err != nil {
		return fmt.Errorf("loading settings: %w", err)
	}
	slog.Info("database ready", "path", paths.DB, "threads", settings.Threads)

	app := NewApp(paths, bins, db, settings)
	if err := wails.Run(&options.App{
		Title:  "Leneki",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	}); err != nil {
		slog.Error("wails exited with an error", "error", err)
		return err
	}
	return nil
}
