-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_profiles ADD COLUMN IF NOT EXISTS role_id INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_profiles DROP COLUMN IF EXISTS role_id;
-- +goose StatementEnd
