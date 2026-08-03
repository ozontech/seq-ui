-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    value TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS roles_permissions(
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users_roles(
    user_id BIGINT NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_users_roles_role_id ON users_roles(role_id);

INSERT INTO permissions(value) VALUES ('roles:create'), ('roles:read'), ('roles:update'), ('roles:delete') ON CONFLICT (value) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_roles_role_id;

DROP TABLE IF EXISTS users_roles;

DROP TABLE IF EXISTS roles_permissions;

DROP TABLE IF EXISTS permissions;

DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
