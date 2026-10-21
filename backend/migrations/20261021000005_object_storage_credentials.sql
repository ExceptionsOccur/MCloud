-- +goose Up
-- T-088：对象存储访问凭证——AccessKey 明文、SecretKey 以 AES-GCM 密文列存储（AI_SECRET_KEY 派生密钥）
ALTER TABLE object_storages ADD COLUMN IF NOT EXISTS access_key VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE object_storages ADD COLUMN IF NOT EXISTS secret_key_enc TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE object_storages DROP COLUMN IF EXISTS access_key;
ALTER TABLE object_storages DROP COLUMN IF EXISTS secret_key_enc;
