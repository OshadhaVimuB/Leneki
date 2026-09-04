package events

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/OshadhaVimuB/Leneki/internal/apperr"
)

func TestRecorderCapturesInOrder(t *testing.T) {
	r := NewRecorder()
	r.Emit(JobState, JobStatePayload{ID: "a", State: "queued"})
	r.Emit(JobProgress, JobProgressPayload{ID: "a", Percent: 50})
	r.Emit(JobState, JobStatePayload{ID: "a", State: "done"})

	all := r.All()
	if len(all) != 3 {
		t.Fatalf("recorded %d events, want 3", len(all))
	}
	states := r.Named(JobState)
	if len(states) != 2 {
		t.Fatalf("Named returned %d, want 2", len(states))
	}
	if states[1].Payload.(JobStatePayload).State != "done" {
		t.Error("events came back out of order")
	}
}

func TestRecorderIsSafeUnderConcurrency(t *testing.T) {
	r := NewRecorder()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Emit(JobProgress, nil)
		}()
	}
	wg.Wait()
	if len(r.All()) != 50 {
		t.Fatalf("recorded %d events, want 50", len(r.All()))
	}
}

func TestJobStateNeverSerializesTheCause(t *testing.T) {
	p := JobStatePayload{
		ID:    "a",
		State: "failed",
		Error: apperr.New(apperr.CodeTranscribeFailed, errNoisy{}),
	}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "exit status 137") {
		t.Fatalf("internal cause leaked to the frontend: %s", out)
	}
	if !strings.Contains(string(out), apperr.Message(apperr.CodeTranscribeFailed)) {
		t.Fatalf("user message missing: %s", out)
	}
}

type errNoisy struct{}

func (errNoisy) Error() string { return "exit status 137" }
