-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    permissions BIGINT NOT NULL DEFAULT 0
);

ALTER TABLE IF EXISTS user_profiles ADD COLUMN IF NOT EXISTS role_id INTEGER;

CREATE INDEX IF NOT EXISTS idx_user_profiles_role_id ON user_profiles USING HASH (role_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_profiles_role_id;

ALTER TABLE IF EXISTS user_profiles DROP COLUMN IF EXISTS role_id;

DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
