-- +goose Up
-- T-083：主机「开放端口」字段不再使用（dev 库 5 条有值数据经确认接受丢失）
ALTER TABLE hosts DROP COLUMN IF EXISTS open_ports;

-- +goose Down
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS open_ports TEXT;
