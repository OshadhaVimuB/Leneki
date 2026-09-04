package main

import (
	"context"
	"log/slog"

	"github.com/OshadhaVimuB/Leneki/internal/binaries"
	"github.com/OshadhaVimuB/Leneki/internal/config"
	"github.com/OshadhaVimuB/Leneki/internal/events"
	"github.com/OshadhaVimuB/Leneki/internal/store"
)

type App struct {
	ctx      context.Context
	paths    config.Paths
	bins     binaries.Paths
	store    *store.Store
	settings config.Settings
	emitter  events.Emitter
}

func NewApp(paths config.Paths, bins binaries.Paths, db *store.Store, settings config.Settings) *App {
	return &App{paths: paths, bins: bins, store: db, settings: settings}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.emitter = newWailsEmitter(ctx)
	slog.Info("frontend ready")
}

// Ping proves the frontend can reach the Go core. It returns the app version.
func (a *App) Ping() string {
	return Version
}
