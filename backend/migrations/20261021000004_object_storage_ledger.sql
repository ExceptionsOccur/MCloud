-- +goose Up
-- T-088：对象存储从主机字段升级为独立台账（外挂存储，可被多台主机共享）
-- hosts.object_storage（GB 数值）改名并改为布尔标记：是否使用对象存储（存量 >0 → true）
-- 注：has_object_storage 为派生拼写，必须显式列名；同时用 DO 块保证重跑幂等
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hosts' AND column_name = 'object_storage'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hosts' AND column_name = 'has_object_storage'
    ) THEN
        ALTER TABLE hosts RENAME COLUMN object_storage TO has_object_storage;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hosts' AND column_name = 'has_object_storage'
          AND data_type IN ('integer', 'bigint', 'smallint', 'numeric')
    ) THEN
        ALTER TABLE hosts ALTER COLUMN has_object_storage DROP DEFAULT;
        ALTER TABLE hosts ALTER COLUMN has_object_storage TYPE BOOLEAN USING (COALESCE(has_object_storage, 0) > 0);
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE hosts ALTER COLUMN has_object_storage SET DEFAULT FALSE;
ALTER TABLE hosts ALTER COLUMN has_object_storage SET NOT NULL;

CREATE TABLE IF NOT EXISTS object_storages (
    id BIGSERIAL PRIMARY KEY,
    bucket VARCHAR(128) NOT NULL,
    purpose TEXT NOT NULL DEFAULT '',
    size_gb BIGINT NOT NULL DEFAULT 0,
    remark TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_object_storages_bucket ON object_storages (bucket);

CREATE TABLE IF NOT EXISTS object_storage_hosts (
    id BIGSERIAL PRIMARY KEY,
    object_storage_id BIGINT NOT NULL REFERENCES object_storages(id) ON DELETE CASCADE,
    host_id BIGINT NOT NULL REFERENCES hosts(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_object_storage_hosts_pair ON object_storage_hosts (object_storage_id, host_id);
CREATE INDEX IF NOT EXISTS idx_object_storage_hosts_host_id ON object_storage_hosts (host_id);

-- +goose Down
DROP TABLE IF EXISTS object_storage_hosts;
DROP TABLE IF EXISTS object_storages;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hosts' AND column_name = 'has_object_storage'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hosts' AND column_name = 'object_storage'
    ) THEN
        ALTER TABLE hosts RENAME COLUMN has_object_storage TO object_storage;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hosts' AND column_name = 'object_storage' AND data_type = 'boolean'
    ) THEN
        ALTER TABLE hosts ALTER COLUMN object_storage DROP DEFAULT;
        ALTER TABLE hosts ALTER COLUMN object_storage DROP NOT NULL;
        ALTER TABLE hosts ALTER COLUMN object_storage TYPE INTEGER USING (CASE WHEN object_storage THEN 1 ELSE 0 END);
    END IF;
END $$;
-- +goose StatementEnd
