-- +goose Up
-- T-063：零信任台账新增「账号姓名」字段（account_name_zh），作为账户名（account_name）的补充，选填
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS account_name_zh VARCHAR(64);

-- +goose Down
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS account_name_zh;
