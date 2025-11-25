-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS async_searches ADD COLUMN IF NOT EXISTS meta text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS async_searches DROP COLUMN IF EXISTS meta;
-- +goose StatementEnd
