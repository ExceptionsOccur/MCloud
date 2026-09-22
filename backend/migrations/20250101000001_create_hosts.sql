-- +goose Up
CREATE TABLE IF NOT EXISTS hosts (
    id SERIAL PRIMARY KEY,
    region VARCHAR(64) NOT NULL,
    instance_id VARCHAR(128),
    name VARCHAR(128) NOT NULL,
    private_ip VARCHAR(45) NOT NULL UNIQUE,
    public_ip VARCHAR(45),
    asset_type VARCHAR(32),
    os VARCHAR(64),
    cpu INTEGER,
    cpu_arch VARCHAR(16),
    memory INTEGER,
    disk INTEGER,
    system_disk INTEGER,
    data_disk INTEGER,
    env_type VARCHAR(16),
    is_db_server BOOLEAN DEFAULT FALSE,
    status VARCHAR(32),
    open_ports TEXT,
    tags TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS hosts;
