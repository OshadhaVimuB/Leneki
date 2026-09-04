package store

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	db *sql.DB
}

// Open connects to the database at path, applies any pending migrations, and
// returns a store. The directory must already exist.
func Open(path string) (*Store, error) {
	dsn := "file:" + url.PathEscape(filepath.ToSlash(path)) +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=foreign_keys(on)" +
		"&_pragma=busy_timeout(5000)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	// One connection, so writes serialise here rather than as SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Jobs() *JobsRepo         { return &JobsRepo{db: s.db} }
func (s *Store) Segments() *SegmentsRepo { return &SegmentsRepo{db: s.db} }
func (s *Store) Models() *ModelsRepo     { return &ModelsRepo{db: s.db} }
func (s *Store) Settings() *SettingsRepo { return &SettingsRepo{db: s.db} }

type migration struct {
	version int
	name    string
	sql     string
}

func migrate(db *sql.DB) error {
	all, err := loadMigrations()
	if err != nil {
		return err
	}

	var current int
	if err := db.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("reading schema version: %w", err)
	}

	latest := all[len(all)-1].version
	if current > latest {
		return fmt.Errorf("this database was written by a newer version of Leneki "+
			"(schema %d, this build understands %d). Update Leneki to open it", current, latest)
	}

	for _, m := range all {
		if m.version <= current {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %s: %w", m.name, err)
		}
		// PRAGMA does not accept a bound parameter.
		if _, err := tx.Exec("PRAGMA user_version = " + strconv.Itoa(m.version)); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", m.name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", m.name, err)
		}
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var out []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		version, err := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0])
		if err != nil {
			return nil, fmt.Errorf("migration %s does not start with a number", e.Name())
		}
		body, err := migrationFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, migration{version: version, name: e.Name(), sql: string(body)})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no migrations were embedded")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}
