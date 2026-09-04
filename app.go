package main

import (
	"context"
	"log/slog"

	"github.com/OshadhaVimuB/Leneki/internal/config"
	"github.com/OshadhaVimuB/Leneki/internal/events"
)

type App struct {
	ctx     context.Context
	paths   config.Paths
	emitter events.Emitter
}

func NewApp(paths config.Paths) *App {
	return &App{paths: paths}
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
