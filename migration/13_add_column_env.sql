-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS async_searches ADD COLUMN IF NOT EXISTS env text NOT NULL DEFAULT '';
ALTER TABLE IF EXISTS async_searches DROP CONSTRAINT IF EXISTS async_searches_pkey;
ALTER TABLE IF EXISTS async_searches ADD PRIMARY KEY (env, search_id);

ALTER TABLE IF EXISTS users_roles ADD COLUMN IF NOT EXISTS env text NOT NULL DEFAULT '';
ALTER TABLE IF EXISTS users_roles DROP CONSTRAINT IF EXISTS users_roles_pkey;
ALTER TABLE IF EXISTS users_roles ADD PRIMARY KEY (user_id, env, role_id);

DROP INDEX IF EXISTS idx_users_roles_role_id;
CREATE INDEX IF NOT EXISTS idx_users_roles_role_id_env ON users_roles(role_id, env);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_roles_role_id_env;
CREATE INDEX IF NOT EXISTS idx_users_roles_role_id ON users_roles(role_id);

ALTER TABLE IF EXISTS users_roles DROP CONSTRAINT IF EXISTS users_roles_pkey;
ALTER TABLE IF EXISTS users_roles ADD PRIMARY KEY (user_id, role_id);
ALTER TABLE IF EXISTS users_roles DROP COLUMN IF EXISTS env;

ALTER TABLE IF EXISTS async_searches DROP CONSTRAINT IF EXISTS async_searches_pkey;
ALTER TABLE IF EXISTS async_searches ADD PRIMARY KEY (search_id);
ALTER TABLE IF EXISTS async_searches DROP COLUMN IF EXISTS env;
-- +goose StatementEnd
