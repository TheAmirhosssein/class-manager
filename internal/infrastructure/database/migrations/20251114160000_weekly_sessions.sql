-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    school_id INTEGER NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    session_date TEXT NOT NULL,
    day INTEGER NOT NULL CHECK (day BETWEEN 0 AND 6),
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    subject TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (start_time < end_time),
    UNIQUE (class_id, session_date, start_time)
);

CREATE INDEX IF NOT EXISTS idx_sessions_school_date ON sessions (school_id, session_date DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_class_date ON sessions (class_id, session_date DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
