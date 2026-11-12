-- +goose Up
-- +goose StatementBegin
UPDATE schools
SET name = name || ' — ' || id
WHERE owner_user_id IS NOT NULL
  AND id NOT IN (
    SELECT MIN(id) FROM schools
    WHERE owner_user_id IS NOT NULL
    GROUP BY owner_user_id, name
  );
CREATE UNIQUE INDEX IF NOT EXISTS idx_schools_owner_name_unique ON schools (owner_user_id, name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_schools_owner_name_unique;
-- +goose StatementEnd
