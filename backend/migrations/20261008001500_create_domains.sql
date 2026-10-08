-- +goose Up
CREATE TABLE IF NOT EXISTS domains (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(255) NOT NULL UNIQUE,
    public_ip VARCHAR(45),
    provider VARCHAR(128),
    expires_at TIMESTAMPTZ,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS domains;
