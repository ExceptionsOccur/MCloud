-- +goose Up
-- T-065：零信任申请时间只记录到日期（timestamptz → VARCHAR(10) 严格 YYYY-MM-DD，对齐 host_applications.apply_time）：
-- 存量按服务器会话时区取日期字符串；幂等（已是 varchar 时跳过类型转换）。
-- DO 块须包裹 StatementBegin/End，否则 goose 切分器按体内分号截断（T-057）。
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'zero_trusts' AND column_name = 'apply_time'
          AND data_type <> 'character varying'
    ) THEN
        ALTER TABLE zero_trusts
            ALTER COLUMN apply_time TYPE VARCHAR(10) USING to_char(apply_time, 'YYYY-MM-DD');
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE zero_trusts
    ALTER COLUMN apply_time TYPE TIMESTAMPTZ USING apply_time::timestamptz;
