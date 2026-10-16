-- +goose Up
-- T-065：零信任接入地区由单值改多值（TEXT，逗号分隔多地区，如 南宁,青年）：
-- 列类型 VARCHAR(128)→TEXT 以容纳多值；存量值规整（去首尾空白，幂等）。
ALTER TABLE zero_trusts ALTER COLUMN exit_location TYPE TEXT;

UPDATE zero_trusts SET exit_location = btrim(exit_location) WHERE exit_location IS NOT NULL;

-- +goose Down
-- 还原单列 VARCHAR(128)（超长多值截断由数据库报错，回滚前需人工收敛）
ALTER TABLE zero_trusts ALTER COLUMN exit_location TYPE VARCHAR(128);
