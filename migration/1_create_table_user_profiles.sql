-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_profiles(
    id BIGSERIAL PRIMARY KEY,
    user_name TEXT NOT NULL,
    timezone VARCHAR(64) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS user_name ON user_profiles(user_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_profiles;
-- +goose StatementEnd
