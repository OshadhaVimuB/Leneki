package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/OshadhaVimuB/Leneki/internal/events"
)

// The only place the Wails runtime is bound to events.Emitter. Keeping it here
// rather than in internal/events is what stops services depending on a webview.
type wailsEmitter struct {
	ctx context.Context
}

func newWailsEmitter(ctx context.Context) events.Emitter {
	return &wailsEmitter{ctx: ctx}
}

func (e *wailsEmitter) Emit(name string, payload any) {
	runtime.EventsEmit(e.ctx, name, payload)
}
