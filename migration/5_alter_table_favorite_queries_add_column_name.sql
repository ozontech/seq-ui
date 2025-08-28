-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS favorite_queries ADD COLUMN IF NOT EXISTS name text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS favorite_queries DROP COLUMN IF EXISTS name;
-- +goose StatementEnd
