package store

import (
	"errors"
	"testing"
	"time"

	"github.com/OshadhaVimuB/Leneki/internal/config"
)

func sampleJob(id string) Job {
	return Job{
		ID:            id,
		SourcePath:    "C:/recordings/interview.mp4",
		DisplayName:   "interview.mp4",
		State:         "queued",
		QueuePosition: 1,
		ModelName:     "base",
		Language:      "",
		AudioTrack:    0,
		CreatedAt:     time.UnixMilli(1700000000000).UTC(),
	}
}

func TestJobRoundTrip(t *testing.T) {
	s := openTemp(t)
	want := sampleJob("job-1")
	if err := s.Jobs().Create(want); err != nil {
		t.Fatal(err)
	}

	got, err := s.Jobs().Get("job-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.SourcePath != want.SourcePath || got.State != want.State || got.ModelName != want.ModelName {
		t.Errorf("job came back changed:\n got %+v\nwant %+v", got, want)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, want.CreatedAt)
	}
	if got.StartedAt != nil || got.FinishedAt != nil {
		t.Error("unset timestamps should come back nil, not zero times")
	}
}

func TestJobUpdatePersistsTimestampsAndErrors(t *testing.T) {
	s := openTemp(t)
	j := sampleJob("job-2")
	if err := s.Jobs().Create(j); err != nil {
		t.Fatal(err)
	}

	started := time.UnixMilli(1700000005000).UTC()
	j.State = "failed"
	j.StartedAt = &started
	j.ErrorCode = "TRANSCRIBE_FAILED"
	j.ErrorMessage = "Transcription failed."
	if err := s.Jobs().Update(j); err != nil {
		t.Fatal(err)
	}

	got, err := s.Jobs().Get("job-2")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "failed" || got.ErrorCode != "TRANSCRIBE_FAILED" {
		t.Errorf("failure was not persisted: %+v", got)
	}
	if got.StartedAt == nil || !got.StartedAt.Equal(started) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, started)
	}
}

func TestJobListIsInQueueOrder(t *testing.T) {
	s := openTemp(t)
	for i, id := range []string{"c", "a", "b"} {
		j := sampleJob(id)
		j.QueuePosition = []int{3, 1, 2}[i]
		if err := s.Jobs().Create(j); err != nil {
			t.Fatal(err)
		}
	}
	jobs, err := s.Jobs().List()
	if err != nil {
		t.Fatal(err)
	}
	got := []string{jobs[0].ID, jobs[1].ID, jobs[2].ID}
	want := []string{"a", "b", "c"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("queue order = %v, want %v", got, want)
		}
	}
}

func TestMissingJobIsNotFound(t *testing.T) {
	s := openTemp(t)
	if _, err := s.Jobs().Get("absent"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get = %v, want ErrNotFound", err)
	}
	if err := s.Jobs().Update(sampleJob("absent")); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update = %v, want ErrNotFound", err)
	}
	if err := s.Jobs().Delete("absent"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete = %v, want ErrNotFound", err)
	}
}

func TestSegmentsRoundTripAndReplaceIsWholesale(t *testing.T) {
	s := openTemp(t)
	if err := s.Jobs().Create(sampleJob("job-3")); err != nil {
		t.Fatal(err)
	}

	first := []Segment{
		{StartMS: 0, EndMS: 1500, Text: "hello there", WordsJSON: `[{"w":"hello"}]`},
		{StartMS: 1500, EndMS: 3000, Text: "second line"},
	}
	if err := s.Segments().Replace("job-3", first); err != nil {
		t.Fatal(err)
	}
	got, err := s.Segments().ListByJob("job-3")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d segments, want 2", len(got))
	}
	if got[0].Index != 0 || got[1].Index != 1 {
		t.Errorf("indexes should be assigned in order, got %d and %d", got[0].Index, got[1].Index)
	}
	if got[0].WordsJSON != `[{"w":"hello"}]` {
		t.Errorf("word timings were lost: %q", got[0].WordsJSON)
	}

	if err := s.Segments().Replace("job-3", first[:1]); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Segments().ListByJob("job-3")
	if len(got) != 1 {
		t.Errorf("Replace left %d segments behind, want 1", len(got))
	}
}

func TestUpdateTextMarksSegmentEdited(t *testing.T) {
	s := openTemp(t)
	if err := s.Jobs().Create(sampleJob("job-4")); err != nil {
		t.Fatal(err)
	}
	if err := s.Segments().Replace("job-4", []Segment{{StartMS: 0, EndMS: 1, Text: "Jhon"}}); err != nil {
		t.Fatal(err)
	}
	segs, _ := s.Segments().ListByJob("job-4")
	if segs[0].Edited {
		t.Error("a fresh segment should not be marked edited")
	}

	if err := s.Segments().UpdateText(segs[0].ID, "John"); err != nil {
		t.Fatal(err)
	}
	segs, _ = s.Segments().ListByJob("job-4")
	if segs[0].Text != "John" || !segs[0].Edited {
		t.Errorf("after edit: text=%q edited=%v", segs[0].Text, segs[0].Edited)
	}
}

func TestDeletingAJobRemovesItsSegments(t *testing.T) {
	s := openTemp(t)
	if err := s.Jobs().Create(sampleJob("job-5")); err != nil {
		t.Fatal(err)
	}
	if err := s.Segments().Replace("job-5", []Segment{{StartMS: 0, EndMS: 1, Text: "x"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Jobs().Delete("job-5"); err != nil {
		t.Fatal(err)
	}
	segs, err := s.Segments().ListByJob("job-5")
	if err != nil {
		t.Fatal(err)
	}
	if len(segs) != 0 {
		t.Errorf("%d orphaned segments remain after deleting the job", len(segs))
	}
}

func TestModelRoundTripAndReinstallOverwrites(t *testing.T) {
	s := openTemp(t)
	m := InstalledModel{
		Name: "base", Path: "/models/ggml-base.bin", SizeBytes: 147951465,
		SHA256: "abc123", InstalledAt: time.UnixMilli(1700000000000).UTC(),
	}
	if err := s.Models().Add(m); err != nil {
		t.Fatal(err)
	}
	got, err := s.Models().Get("base")
	if err != nil {
		t.Fatal(err)
	}
	if got.SHA256 != "abc123" || got.SizeBytes != m.SizeBytes {
		t.Errorf("model came back changed: %+v", got)
	}

	m.SHA256 = "def456"
	if err := s.Models().Add(m); err != nil {
		t.Fatalf("reinstalling should overwrite, not fail: %v", err)
	}
	got, _ = s.Models().Get("base")
	if got.SHA256 != "def456" {
		t.Errorf("checksum was not updated on reinstall: %q", got.SHA256)
	}

	if err := s.Models().Remove("base"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Models().Get("base"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Remove = %v, want ErrNotFound", err)
	}
}

func TestSettingsFallBackToDefaultsThenPersist(t *testing.T) {
	s := openTemp(t)
	defaults := config.Settings{ModelDir: "/models", TempDir: "/temp", Threads: 8}

	got, err := s.Settings().Load(defaults)
	if err != nil {
		t.Fatal(err)
	}
	if got != defaults {
		t.Errorf("an empty settings table should yield defaults, got %+v", got)
	}

	changed := defaults
	changed.SelectedModel = "small"
	changed.Threads = 4
	changed.LastExportDir = "/exports"
	if err := s.Settings().Save(changed); err != nil {
		t.Fatal(err)
	}

	got, err = s.Settings().Load(defaults)
	if err != nil {
		t.Fatal(err)
	}
	if got != changed {
		t.Errorf("settings did not round trip:\n got %+v\nwant %+v", got, changed)
	}
}
