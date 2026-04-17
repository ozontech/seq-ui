-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    permissions BIGINT NOT NULL DEFAULT 0
);

ALTER TABLE IF EXISTS user_profiles ADD COLUMN IF NOT EXISTS role_id INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_profiles DROP COLUMN IF EXISTS role_id;

DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
