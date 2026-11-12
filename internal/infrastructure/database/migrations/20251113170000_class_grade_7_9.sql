-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

CREATE TABLE classes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    school_id INTEGER NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    grade TEXT NOT NULL CHECK (grade IN ('7', '8', '9', '10', '11', '12')),
    room TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO classes_new (id, school_id, title, grade, room, created_at, updated_at)
SELECT id, school_id, title, grade, room, created_at, updated_at FROM classes;

DROP TABLE classes;
ALTER TABLE classes_new RENAME TO classes;

CREATE INDEX IF NOT EXISTS idx_classes_school_id ON classes (school_id);

PRAGMA foreign_keys = ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

CREATE TABLE classes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    school_id INTEGER NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    grade TEXT NOT NULL CHECK (grade IN ('10', '11', '12')),
    room TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO classes_new (id, school_id, title, grade, room, created_at, updated_at)
SELECT id, school_id, title, grade, room, created_at, updated_at FROM classes
WHERE grade IN ('10', '11', '12');

DROP TABLE classes;
ALTER TABLE classes_new RENAME TO classes;

CREATE INDEX IF NOT EXISTS idx_classes_school_id ON classes (school_id);

PRAGMA foreign_keys = ON;
-- +goose StatementEnd
