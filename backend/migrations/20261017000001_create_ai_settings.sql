-- +goose Up
-- AI 接入全局配置（T-067）：单条配置，api_key 以 AES-GCM 密文存储
CREATE TABLE IF NOT EXISTS ai_settings (
    id BIGSERIAL PRIMARY KEY,
    api_url VARCHAR(255) NOT NULL DEFAULT '',
    api_key_enc TEXT NOT NULL DEFAULT '',
    model VARCHAR(128) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS ai_settings;
