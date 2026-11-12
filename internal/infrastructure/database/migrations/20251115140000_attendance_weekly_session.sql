-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = ON;

ALTER TABLE attendance_sessions ADD COLUMN weekly_session_id INTEGER REFERENCES sessions(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_sessions_weekly_session ON attendance_sessions (weekly_session_id)
WHERE weekly_session_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_attendance_sessions_weekly_session;
-- SQLite cannot DROP COLUMN in older versions; table rebuild omitted for dev rollback.
-- +goose StatementEnd
