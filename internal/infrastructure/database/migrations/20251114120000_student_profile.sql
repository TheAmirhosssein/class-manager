-- +goose Up
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
    CASE
        WHEN instr(trim(name), ' ') > 0 THEN trim(substr(name, 1, instr(trim(name), ' ') - 1))
        ELSE trim(name)
    END,
    CASE
        WHEN instr(trim(name), ' ') > 0 THEN trim(substr(name, instr(trim(name), ' ') + 1))
        ELSE 'نامشخص'
    END,
    printf('%010d', id),
    '',
    '',
    created_at,
    updated_at
FROM students;

DROP TABLE students;
ALTER TABLE students_new RENAME TO students;

CREATE INDEX IF NOT EXISTS idx_students_class_id ON students (class_id);

PRAGMA foreign_keys = ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys = OFF;

CREATE TABLE students_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(trim(name)) >= 2),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO students_new (id, class_id, name, created_at, updated_at)
SELECT id, class_id, trim(first_name || ' ' || last_name), created_at, updated_at
FROM students;

DROP TABLE students;
ALTER TABLE students_new RENAME TO students;

CREATE INDEX IF NOT EXISTS idx_students_class_id ON students (class_id);

PRAGMA foreign_keys = ON;
-- +goose StatementEnd
