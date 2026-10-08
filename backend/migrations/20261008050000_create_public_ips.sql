-- +goose Up
CREATE TABLE IF NOT EXISTS public_ips (
    id SERIAL PRIMARY KEY,
    ip VARCHAR(45) NOT NULL UNIQUE,
    isp VARCHAR(128),
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS public_ips;
