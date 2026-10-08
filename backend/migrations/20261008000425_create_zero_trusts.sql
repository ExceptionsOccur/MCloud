-- +goose Up
CREATE TABLE IF NOT EXISTS zero_trusts (
    id SERIAL PRIMARY KEY,
    apply_unit VARCHAR(128) NOT NULL,
    account_name VARCHAR(64) NOT NULL,
    contact VARCHAR(64),
    host_id INTEGER NOT NULL REFERENCES hosts(id),
    port INTEGER NOT NULL,
    apply_time TIMESTAMPTZ NOT NULL,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zero_trusts_host_id ON zero_trusts (host_id);

-- +goose Down
DROP TABLE IF EXISTS zero_trusts;
