package events

// Emitter is how services report progress without importing the Wails runtime,
// which is what lets a whole job run inside go test with no window on screen.
type Emitter interface {
	Emit(name string, payload any)
}
