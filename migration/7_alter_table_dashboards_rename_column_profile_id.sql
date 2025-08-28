-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_dashboards_profile_id;
ALTER TABLE IF EXISTS dashboards RENAME COLUMN profile_id TO owner_id;
CREATE INDEX IF NOT EXISTS idx_dashboards_owner_id ON dashboards USING HASH (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_dashboards_owner_id;
ALTER TABLE IF EXISTS dashboards RENAME COLUMN owner_id TO profile_id;
CREATE INDEX IF NOT EXISTS idx_dashboards_profile_id ON dashboards USING HASH (profile_id);
-- +goose StatementEnd
