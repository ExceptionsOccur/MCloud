-- +goose Up
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS system_name VARCHAR(128);

-- +goose Down
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS system_name;
