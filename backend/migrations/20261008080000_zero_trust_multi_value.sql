-- +goose Up
-- 零信任台账：单主机单端口改为多主机多端口（聚合一条申请记录，笛卡尔授权）
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS host_ids TEXT;
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS ports TEXT;

UPDATE zero_trusts SET host_ids = host_id::text WHERE host_ids IS NULL;
UPDATE zero_trusts SET ports = port::text WHERE ports IS NULL;

ALTER TABLE zero_trusts DROP CONSTRAINT IF EXISTS zero_trusts_host_id_fkey;
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS host_id;
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS port;

ALTER TABLE zero_trusts ALTER COLUMN host_ids SET NOT NULL;
ALTER TABLE zero_trusts ALTER COLUMN ports SET NOT NULL;

DROP INDEX IF EXISTS idx_zero_trusts_host_id;

-- +goose Down
-- 还原单值列（仅保留第一条主机/端口，多值部分丢失）
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS host_id INTEGER;
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS port INTEGER;

UPDATE zero_trusts SET host_id = (string_to_array(host_ids, ','))[1]::integer WHERE host_id IS NULL;
UPDATE zero_trusts SET port = (string_to_array(ports, ','))[1]::integer WHERE port IS NULL;

ALTER TABLE zero_trusts ALTER COLUMN host_id SET NOT NULL;
ALTER TABLE zero_trusts ALTER COLUMN port SET NOT NULL;

ALTER TABLE zero_trusts DROP COLUMN IF EXISTS host_ids;
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS ports;

ALTER TABLE zero_trusts ADD CONSTRAINT zero_trusts_host_id_fkey FOREIGN KEY (host_id) REFERENCES hosts(id);
CREATE INDEX IF NOT EXISTS idx_zero_trusts_host_id ON zero_trusts (host_id);
