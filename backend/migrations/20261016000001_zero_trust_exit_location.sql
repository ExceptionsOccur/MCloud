-- +goose Up
-- T-061：零信任台账去「公网IP」改独立「接入地区」：
-- 删除 public_ip 列（公网IP入口不再需要），新增 exit_location 独立列；
-- 存量数据由原 public_ip 关联的 public_ips.exit_location 回填（幂等：public_ip 列不存在时跳过回填）。
-- DO 块须包裹 StatementBegin/End，否则 goose 切分器按体内分号截断（T-057）。
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS exit_location VARCHAR(128);

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'zero_trusts' AND column_name = 'public_ip'
    ) THEN
        UPDATE zero_trusts zt
        SET exit_location = pi.exit_location
        FROM public_ips pi
        WHERE zt.public_ip = pi.ip
          AND (zt.exit_location IS NULL OR zt.exit_location = '')
          AND pi.exit_location <> '';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE zero_trusts DROP COLUMN IF EXISTS public_ip;

-- +goose Down
-- 还原 public_ip 列（回填值无法还原，置 NULL），删除 exit_location
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS public_ip VARCHAR(45);

ALTER TABLE zero_trusts DROP COLUMN IF EXISTS exit_location;
