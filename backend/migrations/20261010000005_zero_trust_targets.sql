-- +goose Up
-- zero_trusts.targets 拆子表（T-052）：`host_id:port` 逗号多值 text → 1NF 子表，
-- 补 FK（RESTRICT，任务口径）与索引；回填后删除旧列。
-- 外键命名与 GORM 约定名（fk_<表>_<关联>，RelationshipFKName）一致，AutoMigrate 不会重复建。
CREATE TABLE IF NOT EXISTS zero_trust_targets (
    id BIGSERIAL PRIMARY KEY,
    zero_trust_id BIGINT NOT NULL,
    host_id BIGINT NOT NULL,
    port INTEGER NOT NULL,
    CONSTRAINT fk_zero_trust_targets_zero_trust
        FOREIGN KEY (zero_trust_id) REFERENCES zero_trusts(id) ON DELETE RESTRICT,
    CONSTRAINT fk_zero_trust_targets_host
        FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_zero_trust_targets_zero_trust_id ON zero_trust_targets (zero_trust_id);
CREATE INDEX IF NOT EXISTS idx_zero_trust_targets_host_id ON zero_trust_targets (host_id);

-- 回填：解析口径与原应用层存储串一致（段内恰一个冒号、纯数字段、端口 1-5 位）；
-- 分层保证先过字符校验再做数值转换；主机已不存在的悬空配对跳过（FK 不允许悬空）
INSERT INTO zero_trust_targets (zero_trust_id, host_id, port)
SELECT g.zid, g.hid::bigint, g.port::int
FROM (
    SELECT s.zid, s.hid, s.port
    FROM (
        SELECT z.id AS zid,
               btrim(split_part(t.val, ':', 1)) AS hid,
               btrim(split_part(t.val, ':', 2)) AS port
        FROM zero_trusts z
        CROSS JOIN LATERAL unnest(string_to_array(z.targets, ',')) AS t(val)
        WHERE position(':' in t.val) > 0
          AND split_part(t.val, ':', 3) = ''
    ) s
    WHERE s.hid ~ '^[0-9]{1,18}$'
      AND s.port ~ '^[0-9]{1,5}$'
) g
WHERE g.port::int BETWEEN 1 AND 65535
  AND EXISTS (SELECT 1 FROM hosts hs WHERE hs.id = g.hid::bigint);

ALTER TABLE zero_trusts DROP COLUMN IF EXISTS targets;

-- +goose Down
-- 还原 targets 列（由子表聚合回 `host_id:port` 逗号串），再删子表
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS targets TEXT NOT NULL DEFAULT '';

UPDATE zero_trusts z
SET targets = COALESCE((
    SELECT string_agg(t.host_id::text || ':' || t.port::text, ',' ORDER BY t.id)
    FROM zero_trust_targets t
    WHERE t.zero_trust_id = z.id
), '');

ALTER TABLE zero_trusts ALTER COLUMN targets DROP DEFAULT;

DROP TABLE IF EXISTS zero_trust_targets;
