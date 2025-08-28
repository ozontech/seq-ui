-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_profiles ADD COLUMN IF NOT EXISTS onboarding_version text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_profiles DROP COLUMN IF EXISTS onboarding_version;
-- +goose StatementEnd
