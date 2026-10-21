-- +goose Up
-- T-078：主机记录新增「是否云外主机」标记；存量数据全部为云内（false）
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS is_external BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE hosts DROP COLUMN IF EXISTS is_external;
