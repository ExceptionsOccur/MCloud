-- +goose Up
-- 配对化：host_ids/ports 两列按位置等长合并为 targets（host_id:port 多组配对），并新增公网IP入口
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS targets TEXT;
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS public_ip VARCHAR(45);

UPDATE zero_trusts zt
SET targets = COALESCE((
    SELECT string_agg(h.val || ':' || COALESCE(p.val, f.val), ',' ORDER BY h.ord)
    FROM unnest(string_to_array(zt.host_ids, ',')) WITH ORDINALITY AS h(val, ord)
    CROSS JOIN LATERAL (SELECT val FROM unnest(string_to_array(zt.ports, ',')) LIMIT 1) AS f
    LEFT JOIN LATERAL (
        SELECT p2.val
        FROM unnest(string_to_array(zt.ports, ',')) WITH ORDINALITY AS p2(val, ord)
        WHERE p2.ord = h.ord
    ) p ON TRUE
), '');

ALTER TABLE zero_trusts DROP COLUMN IF EXISTS host_ids;
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS ports;

ALTER TABLE zero_trusts ALTER COLUMN targets SET NOT NULL;

-- +goose Down
-- 还原两列（按位置拆回 host_ids/ports；主机多于端口时多余主机配第一个端口）
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS host_ids TEXT;
ALTER TABLE zero_trusts ADD COLUMN IF NOT EXISTS ports TEXT;

UPDATE zero_trusts zt
SET host_ids = s.hids, ports = s.pts
FROM (
    SELECT z.id,
           string_agg(split_part(t.val, ':', 1), ',' ORDER BY t.ord) AS hids,
           string_agg(split_part(t.val, ':', 2), ',' ORDER BY t.ord) AS pts
    FROM zero_trusts z,
    LATERAL unnest(string_to_array(z.targets, ',')) WITH ORDINALITY AS t(val, ord)
    GROUP BY z.id
) s
WHERE s.id = zt.id;

ALTER TABLE zero_trusts DROP COLUMN IF EXISTS targets;
ALTER TABLE zero_trusts DROP COLUMN IF EXISTS public_ip;

ALTER TABLE zero_trusts ALTER COLUMN host_ids SET NOT NULL;
ALTER TABLE zero_trusts ALTER COLUMN ports SET NOT NULL;
