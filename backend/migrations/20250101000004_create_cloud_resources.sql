-- +goose Up
CREATE TABLE IF NOT EXISTS cloud_resources (
    id SERIAL PRIMARY KEY,
    region VARCHAR(64) NOT NULL UNIQUE,
    physical_cpu INTEGER DEFAULT 0,
    vcpu INTEGER DEFAULT 0,
    memory INTEGER DEFAULT 0,
    storage INTEGER DEFAULT 0,
    bare_metal INTEGER DEFAULT 0,
    gpu_card_count INTEGER DEFAULT 0,
    object_storage INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS cloud_resources;
