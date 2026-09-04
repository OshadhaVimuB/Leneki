package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "leneki.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenAppliesMigrationsToAnEmptyFile(t *testing.T) {
	s := openTemp(t)

	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("user_version = %d, want 1", version)
	}

	want := []string{"jobs", "segments", "installed_models", "benchmark", "settings"}
	for _, table := range want {
		var name string
		err := s.db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %s was not created: %v", table, err)
		}
	}
}

func TestOpenIsSafeToRepeat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "leneki.db")

	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Jobs().Create(sampleJob("keep-me")); err != nil {
		t.Fatal(err)
	}
	first.Close()

	second, err := Open(path)
	if err != nil {
		t.Fatalf("reopening an existing database failed: %v", err)
	}
	defer second.Close()

	if _, err := second.Jobs().Get("keep-me"); err != nil {
		t.Errorf("data did not survive reopening: %v", err)
	}
}

func TestOpenRefusesANewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "leneki.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()

	raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec("PRAGMA user_version = 999"); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	_, err = Open(path)
	if err == nil {
		t.Fatal("opening a newer database should fail rather than corrupt it")
	}
	msg := err.Error()
	for _, want := range []string{"newer version", "999"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message should mention %q, got: %s", want, msg)
		}
	}
}

func TestForeignKeysAreEnforced(t *testing.T) {
	s := openTemp(t)
	err := s.Segments().Replace("no-such-job", []Segment{{StartMS: 0, EndMS: 1, Text: "hi"}})
	if err == nil {
		t.Fatal("inserting a segment for a missing job should violate the foreign key")
	}
}

func TestMigrationsAreNumberedAndOrdered(t *testing.T) {
	all, err := loadMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no migrations found")
	}
	for i, m := range all {
		if i > 0 && m.version <= all[i-1].version {
			t.Errorf("migration %s is out of order or duplicated", m.name)
		}
	}
}
