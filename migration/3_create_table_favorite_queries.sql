-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS favorite_queries(
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGSERIAL NOT NULL,
    query text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_favorite_queries_profile_id ON favorite_queries(profile_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_favorite_queries_profile_id;
DROP TABLE IF EXISTS favorite_queries;
-- +goose StatementEnd
