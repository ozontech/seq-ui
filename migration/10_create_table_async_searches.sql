-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS async_searches(
    search_id UUID PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    created_at timestamptz NOT NULL default now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_async_searches_owner_id ON async_searches USING HASH (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_async_searches_owner_id;
DROP TABLE IF EXISTS async_searches;
-- +goose StatementEnd
