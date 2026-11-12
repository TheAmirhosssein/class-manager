-- +goose Up
-- +goose StatementBegin
ALTER TABLE schools ADD COLUMN owner_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_schools_owner_user_id ON schools (owner_user_id);
UPDATE schools
SET owner_user_id = (SELECT id FROM users ORDER BY id LIMIT 1)
WHERE owner_user_id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_schools_owner_user_id;
-- SQLite cannot drop column without table rebuild; owner_user_id left in place on rollback.
-- +goose StatementEnd
