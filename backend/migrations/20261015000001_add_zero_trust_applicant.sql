-- +goose Up
-- T-060：零信任台账新增「申请人」字段，与账户名区分；可空，存量记录保持空
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS applicant VARCHAR(64);

-- +goose Down
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS applicant;
