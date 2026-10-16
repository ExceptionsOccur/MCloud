-- +goose Up
-- T-065：零信任接入目标增加 TCP/UDP 协议维度（每目标行单协议，缺省 tcp，存量回填 tcp）
ALTER TABLE zero_trust_targets ADD COLUMN IF NOT EXISTS protocol VARCHAR(8);

UPDATE zero_trust_targets SET protocol = 'tcp' WHERE protocol IS NULL OR protocol = '';

ALTER TABLE zero_trust_targets ALTER COLUMN protocol SET DEFAULT 'tcp';
ALTER TABLE zero_trust_targets ALTER COLUMN protocol SET NOT NULL;

-- +goose Down
ALTER TABLE zero_trust_targets DROP COLUMN IF EXISTS protocol;
