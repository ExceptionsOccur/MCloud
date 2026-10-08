-- +goose Up
ALTER TABLE domains ADD COLUMN IF NOT EXISTS exit_location VARCHAR(128);
ALTER TABLE public_ips ADD COLUMN IF NOT EXISTS exit_location VARCHAR(128);

-- +goose Down
ALTER TABLE domains DROP COLUMN IF EXISTS exit_location;
ALTER TABLE public_ips DROP COLUMN IF EXISTS exit_location;
