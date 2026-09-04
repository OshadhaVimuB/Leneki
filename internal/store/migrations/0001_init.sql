CREATE TABLE jobs (
    id              TEXT PRIMARY KEY,
    source_path     TEXT NOT NULL,
    display_name    TEXT NOT NULL,
    state           TEXT NOT NULL,
    queue_position  INTEGER NOT NULL,
    paused          INTEGER NOT NULL DEFAULT 0,
    model_name      TEXT NOT NULL,
    language        TEXT,
    detected_lang   TEXT,
    translate       INTEGER NOT NULL DEFAULT 0,
    audio_track     INTEGER NOT NULL DEFAULT 0,
    duration_ms     INTEGER,
    playback_path   TEXT,
    error_code      TEXT,
    error_message   TEXT,
    created_at      INTEGER NOT NULL,
    started_at      INTEGER,
    finished_at     INTEGER
);

CREATE INDEX idx_jobs_queue ON jobs (state, queue_position);

CREATE TABLE segments (
    id          INTEGER PRIMARY KEY,
    job_id      TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    idx         INTEGER NOT NULL,
    start_ms    INTEGER NOT NULL,
    end_ms      INTEGER NOT NULL,
    text        TEXT NOT NULL,
    words_json  TEXT,
    edited      INTEGER NOT NULL DEFAULT 0,
    UNIQUE (job_id, idx)
);

CREATE INDEX idx_segments_job_start ON segments (job_id, start_ms);

CREATE TABLE installed_models (
    name         TEXT PRIMARY KEY,
    path         TEXT NOT NULL,
    size_bytes   INTEGER NOT NULL,
    sha256       TEXT NOT NULL,
    installed_at INTEGER NOT NULL
);

CREATE TABLE benchmark (
    fingerprint TEXT PRIMARY KEY,
    rtf_base    REAL NOT NULL,
    cpu         TEXT NOT NULL,
    gpu         TEXT,
    ram_bytes   INTEGER NOT NULL,
    ran_at      INTEGER NOT NULL
);

CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
