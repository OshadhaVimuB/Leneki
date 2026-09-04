package store

import (
	"database/sql"
	"fmt"
)

type Segment struct {
	ID        int64
	JobID     string
	Index     int
	StartMS   int64
	EndMS     int64
	Text      string
	WordsJSON string
	Edited    bool
}

type SegmentsRepo struct {
	db *sql.DB
}

// Replace writes the authoritative segment list for a job in one transaction,
// so a reader never sees a half written transcript.
func (r *SegmentsRepo) Replace(jobID string, segs []Segment) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM segments WHERE job_id=?`, jobID); err != nil {
		return fmt.Errorf("clearing segments for job %s: %w", jobID, err)
	}
	stmt, err := tx.Prepare(`
		INSERT INTO segments (job_id, idx, start_ms, end_ms, text, words_json, edited)
		VALUES (?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, s := range segs {
		if _, err := stmt.Exec(jobID, i, s.StartMS, s.EndMS, s.Text, s.WordsJSON, s.Edited); err != nil {
			return fmt.Errorf("inserting segment %d for job %s: %w", i, jobID, err)
		}
	}
	return tx.Commit()
}

func (r *SegmentsRepo) ListByJob(jobID string) ([]Segment, error) {
	rows, err := r.db.Query(`
		SELECT id, job_id, idx, start_ms, end_ms, text, words_json, edited
		FROM segments WHERE job_id=? ORDER BY idx`, jobID)
	if err != nil {
		return nil, fmt.Errorf("listing segments for job %s: %w", jobID, err)
	}
	defer rows.Close()

	var out []Segment
	for rows.Next() {
		var s Segment
		var words sql.NullString
		if err := rows.Scan(&s.ID, &s.JobID, &s.Index, &s.StartMS, &s.EndMS,
			&s.Text, &words, &s.Edited); err != nil {
			return nil, err
		}
		s.WordsJSON = words.String
		out = append(out, s)
	}
	return out, rows.Err()
}

// UpdateText is what autosave calls. It marks the segment as edited.
func (r *SegmentsRepo) UpdateText(id int64, text string) error {
	res, err := r.db.Exec(`UPDATE segments SET text=?, edited=1 WHERE id=?`, text, id)
	if err != nil {
		return fmt.Errorf("updating segment %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("segment %d: %w", id, ErrNotFound)
	}
	return nil
}
