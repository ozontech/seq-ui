-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS favorite_queries ADD COLUMN IF NOT EXISTS relative_from bigint NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS favorite_queries DROP COLUMN IF EXISTS relative_from;
-- +goose StatementEnd
