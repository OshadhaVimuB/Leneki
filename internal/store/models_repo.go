package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type InstalledModel struct {
	Name        string
	Path        string
	SizeBytes   int64
	SHA256      string
	InstalledAt time.Time
}

type ModelsRepo struct {
	db *sql.DB
}

// Add records a model that is already verified and in place.
func (r *ModelsRepo) Add(m InstalledModel) error {
	_, err := r.db.Exec(`
		INSERT INTO installed_models (name, path, size_bytes, sha256, installed_at)
		VALUES (?,?,?,?,?)
		ON CONFLICT(name) DO UPDATE SET
			path=excluded.path, size_bytes=excluded.size_bytes,
			sha256=excluded.sha256, installed_at=excluded.installed_at`,
		m.Name, m.Path, m.SizeBytes, m.SHA256, m.InstalledAt.UnixMilli())
	if err != nil {
		return fmt.Errorf("recording model %s: %w", m.Name, err)
	}
	return nil
}

func (r *ModelsRepo) Get(name string) (InstalledModel, error) {
	var m InstalledModel
	var installed int64
	err := r.db.QueryRow(`
		SELECT name, path, size_bytes, sha256, installed_at
		FROM installed_models WHERE name=?`, name).
		Scan(&m.Name, &m.Path, &m.SizeBytes, &m.SHA256, &installed)
	if errors.Is(err, sql.ErrNoRows) {
		return InstalledModel{}, fmt.Errorf("model %s: %w", name, ErrNotFound)
	}
	if err != nil {
		return InstalledModel{}, err
	}
	m.InstalledAt = time.UnixMilli(installed).UTC()
	return m, nil
}

func (r *ModelsRepo) List() ([]InstalledModel, error) {
	rows, err := r.db.Query(`
		SELECT name, path, size_bytes, sha256, installed_at
		FROM installed_models ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing models: %w", err)
	}
	defer rows.Close()

	var out []InstalledModel
	for rows.Next() {
		var m InstalledModel
		var installed int64
		if err := rows.Scan(&m.Name, &m.Path, &m.SizeBytes, &m.SHA256, &installed); err != nil {
			return nil, err
		}
		m.InstalledAt = time.UnixMilli(installed).UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ModelsRepo) Remove(name string) error {
	res, err := r.db.Exec(`DELETE FROM installed_models WHERE name=?`, name)
	if err != nil {
		return fmt.Errorf("removing model %s: %w", name, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("model %s: %w", name, ErrNotFound)
	}
	return nil
}
