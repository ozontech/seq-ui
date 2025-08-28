-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS log_columns(
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGSERIAL NOT NULL,
    columns text NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_log_columns_profile_id ON log_columns(profile_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_log_columns_profile_id;
DROP TABLE IF EXISTS log_columns;
-- +goose StatementEnd
