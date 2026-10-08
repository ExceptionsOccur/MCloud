-- +goose Up
ALTER TABLE domains ADD COLUMN IF NOT EXISTS host_id BIGINT;
ALTER TABLE domains ADD COLUMN IF NOT EXISTS host_port INTEGER;
ALTER TABLE domains RENAME COLUMN provider TO isp;
ALTER TABLE domains DROP COLUMN IF EXISTS expires_at;

CREATE INDEX IF NOT EXISTS idx_domains_host_id ON domains (host_id);

ALTER TABLE domains DROP CONSTRAINT IF EXISTS fk_domains_host;
ALTER TABLE domains ADD CONSTRAINT fk_domains_host FOREIGN KEY (host_id) REFERENCES hosts(id);

-- +goose Down
ALTER TABLE domains DROP CONSTRAINT IF EXISTS fk_domains_host;
DROP INDEX IF EXISTS idx_domains_host_id;
ALTER TABLE domains ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
ALTER TABLE domains RENAME COLUMN isp TO provider;
ALTER TABLE domains DROP COLUMN IF EXISTS host_id;
ALTER TABLE domains DROP COLUMN IF EXISTS host_port;
