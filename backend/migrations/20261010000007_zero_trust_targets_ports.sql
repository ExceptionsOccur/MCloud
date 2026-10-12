-- +goose Up
-- T-056：接入目标端口列改表达式（单端口/范围/列表，如 `22` 或 `1000-10005,22`），
-- integer → text，列名 port → ports；条件守卫保证重复执行 no-op
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'zero_trust_targets'
          AND column_name = 'port'
    ) THEN
        ALTER TABLE zero_trust_targets ALTER COLUMN port TYPE TEXT USING port::text;
        ALTER TABLE zero_trust_targets RENAME COLUMN port TO ports;
    END IF;
END $$;

-- +goose Down
-- 尽力还原：取表达式开头的数字段回 integer（多端口行退化为首端口，需人工核对）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'zero_trust_targets'
          AND column_name = 'ports'
    ) THEN
        ALTER TABLE zero_trust_targets ALTER COLUMN ports DROP NOT NULL;
        ALTER TABLE zero_trust_targets ALTER COLUMN ports TYPE INTEGER
            USING (substring(ports from '^[0-9]+'))::integer;
        ALTER TABLE zero_trust_targets RENAME COLUMN ports TO port;
        ALTER TABLE zero_trust_targets ALTER COLUMN port SET NOT NULL;
    END IF;
END $$;
