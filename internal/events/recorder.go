package events

import "sync"

type Recorded struct {
	Name    string
	Payload any
}

// Recorder is the Emitter used by tests.
type Recorder struct {
	mu       sync.Mutex
	recorded []Recorded
}

func NewRecorder() *Recorder { return &Recorder{} }

func (r *Recorder) Emit(name string, payload any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = append(r.recorded, Recorded{Name: name, Payload: payload})
}

func (r *Recorder) All() []Recorded {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Recorded(nil), r.recorded...)
}

func (r *Recorder) Named(name string) []Recorded {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Recorded
	for _, e := range r.recorded {
		if e.Name == name {
			out = append(out, e)
		}
	}
	return out
}

func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = nil
}
