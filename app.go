package main

import "context"

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Ping proves the frontend can reach the Go core. It returns the app version.
func (a *App) Ping() string {
	return Version
}
