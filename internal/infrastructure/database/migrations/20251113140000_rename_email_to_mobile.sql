-- +goose Up
-- +goose StatementBegin
ALTER TABLE users RENAME COLUMN email TO mobile;
ALTER TABLE admins RENAME COLUMN email TO mobile;

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_admins_email;
CREATE INDEX IF NOT EXISTS idx_users_mobile ON users (mobile);
CREATE INDEX IF NOT EXISTS idx_admins_mobile ON admins (mobile);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_mobile;
DROP INDEX IF EXISTS idx_admins_mobile;
CREATE INDEX IF NOT EXISTS idx_users_email ON users (mobile);
CREATE INDEX IF NOT EXISTS idx_admins_email ON admins (mobile);

ALTER TABLE users RENAME COLUMN mobile TO email;
ALTER TABLE admins RENAME COLUMN mobile TO email;
-- +goose StatementEnd
