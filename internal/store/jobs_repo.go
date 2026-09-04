package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("not found")

type Job struct {
	ID            string
	SourcePath    string
	DisplayName   string
	State         string
	QueuePosition int
	Paused        bool
	ModelName     string
	Language      string
	DetectedLang  string
	Translate     bool
	AudioTrack    int
	DurationMS    int64
	PlaybackPath  string
	ErrorCode     string
	ErrorMessage  string
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
}

type JobsRepo struct {
	db *sql.DB
}

const jobColumns = `id, source_path, display_name, state, queue_position, paused,
	model_name, language, detected_lang, translate, audio_track, duration_ms,
	playback_path, error_code, error_message, created_at, started_at, finished_at`

func (r *JobsRepo) Create(j Job) error {
	_, err := r.db.Exec(`
		INSERT INTO jobs (`+jobColumns+`)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		j.ID, j.SourcePath, j.DisplayName, j.State, j.QueuePosition, j.Paused,
		j.ModelName, j.Language, j.DetectedLang, j.Translate, j.AudioTrack, j.DurationMS,
		j.PlaybackPath, j.ErrorCode, j.ErrorMessage, j.CreatedAt.UnixMilli(),
		nullMillis(j.StartedAt), nullMillis(j.FinishedAt))
	if err != nil {
		return fmt.Errorf("creating job %s: %w", j.ID, err)
	}
	return nil
}

func (r *JobsRepo) Update(j Job) error {
	res, err := r.db.Exec(`
		UPDATE jobs SET
			source_path=?, display_name=?, state=?, queue_position=?, paused=?,
			model_name=?, language=?, detected_lang=?, translate=?, audio_track=?,
			duration_ms=?, playback_path=?, error_code=?, error_message=?,
			created_at=?, started_at=?, finished_at=?
		WHERE id=?`,
		j.SourcePath, j.DisplayName, j.State, j.QueuePosition, j.Paused,
		j.ModelName, j.Language, j.DetectedLang, j.Translate, j.AudioTrack,
		j.DurationMS, j.PlaybackPath, j.ErrorCode, j.ErrorMessage,
		j.CreatedAt.UnixMilli(), nullMillis(j.StartedAt), nullMillis(j.FinishedAt), j.ID)
	if err != nil {
		return fmt.Errorf("updating job %s: %w", j.ID, err)
	}
	return oneRow(res, j.ID)
}

func (r *JobsRepo) Get(id string) (Job, error) {
	row := r.db.QueryRow(`SELECT `+jobColumns+` FROM jobs WHERE id=?`, id)
	j, err := scanJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, fmt.Errorf("job %s: %w", id, ErrNotFound)
	}
	return j, err
}

// List returns jobs in queue order, newest last.
func (r *JobsRepo) List() ([]Job, error) {
	rows, err := r.db.Query(`SELECT ` + jobColumns + ` FROM jobs ORDER BY queue_position, created_at`)
	if err != nil {
		return nil, fmt.Errorf("listing jobs: %w", err)
	}
	defer rows.Close()

	var out []Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *JobsRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM jobs WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("deleting job %s: %w", id, err)
	}
	return oneRow(res, id)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanJob(s scanner) (Job, error) {
	var j Job
	var created int64
	var started, finished sql.NullInt64
	err := s.Scan(&j.ID, &j.SourcePath, &j.DisplayName, &j.State, &j.QueuePosition,
		&j.Paused, &j.ModelName, &j.Language, &j.DetectedLang, &j.Translate,
		&j.AudioTrack, &j.DurationMS, &j.PlaybackPath, &j.ErrorCode, &j.ErrorMessage,
		&created, &started, &finished)
	if err != nil {
		return Job{}, err
	}
	j.CreatedAt = time.UnixMilli(created).UTC()
	j.StartedAt = millisToTime(started)
	j.FinishedAt = millisToTime(finished)
	return j, nil
}

func nullMillis(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UnixMilli()
}

func millisToTime(n sql.NullInt64) *time.Time {
	if !n.Valid {
		return nil
	}
	t := time.UnixMilli(n.Int64).UTC()
	return &t
}

func oneRow(res sql.Result, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("job %s: %w", id, ErrNotFound)
	}
	return nil
}
