-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dashboards(
    uuid UUID PRIMARY KEY,
    profile_id BIGSERIAL NOT NULL,
    name VARCHAR(64) NOT NULL,
    meta text NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dashboards_profile_id ON dashboards USING HASH (profile_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_dashboards_profile_id;
DROP TABLE IF EXISTS dashboards;
-- +goose StatementEnd
