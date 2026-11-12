-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

CREATE TABLE students_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL CHECK (length(trim(first_name)) BETWEEN 2 AND 60),
    last_name TEXT NOT NULL CHECK (length(trim(last_name)) BETWEEN 2 AND 60),
    national_code TEXT CHECK (
        national_code IS NULL
        OR (
            length(national_code) = 10
            AND national_code GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]'
        )
    ),
    father_name TEXT NOT NULL DEFAULT '' CHECK (length(trim(father_name)) <= 60),
    student_number TEXT NOT NULL DEFAULT '' CHECK (length(trim(student_number)) <= 20),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO students_new (id, class_id, first_name, last_name, national_code, father_name, student_number, created_at, updated_at)
SELECT
    id,
    class_id,
    first_name,
    last_name,
    CASE
        WHEN national_code IS NULL OR trim(national_code) = '' THEN NULL
        WHEN national_code = printf('%010d', id) THEN NULL
        ELSE national_code
    END,
    father_name,
    student_number,
    created_at,
    updated_at
FROM students;

DROP TABLE students;
ALTER TABLE students_new RENAME TO students;

CREATE INDEX IF NOT EXISTS idx_students_class_id ON students (class_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_national_code_unique ON students (national_code)
WHERE national_code IS NOT NULL AND length(trim(national_code)) > 0;

PRAGMA foreign_keys = ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

CREATE TABLE students_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL CHECK (length(trim(first_name)) BETWEEN 2 AND 60),
    last_name TEXT NOT NULL CHECK (length(trim(last_name)) BETWEEN 2 AND 60),
    national_code TEXT NOT NULL CHECK (length(national_code) = 10 AND national_code GLOB '[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]'),
    father_name TEXT NOT NULL DEFAULT '' CHECK (length(trim(father_name)) <= 60),
    student_number TEXT NOT NULL DEFAULT '' CHECK (length(trim(student_number)) <= 20),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (national_code)
);

INSERT INTO students_new (id, class_id, first_name, last_name, national_code, father_name, student_number, created_at, updated_at)
SELECT
    id,
    class_id,
    first_name,
    last_name,
    coalesce(nullif(trim(national_code), ''), printf('%010d', id)),
    father_name,
    student_number,
    created_at,
    updated_at
FROM students;

DROP TABLE students;
ALTER TABLE students_new RENAME TO students;

CREATE INDEX IF NOT EXISTS idx_students_class_id ON students (class_id);

PRAGMA foreign_keys = ON;
-- +goose StatementEnd
