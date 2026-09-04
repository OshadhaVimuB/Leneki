package store

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/OshadhaVimuB/Leneki/internal/config"
)

type SettingsRepo struct {
	db *sql.DB
}

const (
	keySelectedModel = "selected_model"
	keyModelDir      = "model_dir"
	keyTempDir       = "temp_dir"
	keyThreads       = "threads"
	keyLastExportDir = "last_export_dir"
)

// Load returns stored settings, falling back to defaults for anything unset,
// so a partially written settings table can never leave a field empty.
func (r *SettingsRepo) Load(defaults config.Settings) (config.Settings, error) {
	rows, err := r.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return defaults, fmt.Errorf("loading settings: %w", err)
	}
	defer rows.Close()

	s := defaults
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return defaults, err
		}
		switch k {
		case keySelectedModel:
			s.SelectedModel = v
		case keyModelDir:
			s.ModelDir = v
		case keyTempDir:
			s.TempDir = v
		case keyLastExportDir:
			s.LastExportDir = v
		case keyThreads:
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				s.Threads = n
			}
		}
	}
	return s, rows.Err()
}

func (r *SettingsRepo) Save(s config.Settings) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO settings (key, value) VALUES (?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	pairs := [][2]string{
		{keySelectedModel, s.SelectedModel},
		{keyModelDir, s.ModelDir},
		{keyTempDir, s.TempDir},
		{keyLastExportDir, s.LastExportDir},
		{keyThreads, strconv.Itoa(s.Threads)},
	}
	for _, p := range pairs {
		if _, err := stmt.Exec(p[0], p[1]); err != nil {
			return fmt.Errorf("saving setting %s: %w", p[0], err)
		}
	}
	return tx.Commit()
}
