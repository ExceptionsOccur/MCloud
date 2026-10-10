-- +goose Up
-- T-048 迁移 SQL ↔ 实际库漂移对齐（范围已确认：不做软删除/CHECK/普通索引补齐/people_pkey 改名）
-- 1) 主键 integer → bigint（含序列类型），2) created_at 缺省补 DEFAULT now()，
-- 3) 唯一索引 idx_ 前缀对齐，4) 列类型按实库对齐，5) host_applications 补 created_at/updated_at，
-- 6) 旧迁移 ON DELETE CASCADE 与实库 NO ACTION 的口径对齐。全部语句幂等，可重跑。

-- 1) 主键 integer → bigint（同类型重复执行无害）
ALTER TABLE port_mappings ALTER COLUMN id TYPE bigint USING id::bigint;
ALTER SEQUENCE port_mappings_id_seq AS bigint;
ALTER TABLE public_ips ALTER COLUMN id TYPE bigint USING id::bigint;
ALTER SEQUENCE public_ips_id_seq AS bigint;

-- 2) created_at 缺省补 DEFAULT now()（重复 SET 幂等）
ALTER TABLE cloud_resources ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE hosts           ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE ip_subnets      ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE persons         ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE users           ALTER COLUMN created_at SET DEFAULT now();
ALTER TABLE zero_trusts     ALTER COLUMN created_at SET DEFAULT now();

-- 3) 唯一索引 idx_ 前缀对齐（重跑时 uni_ 已不存在，IF EXISTS 跳过）
ALTER INDEX IF EXISTS uni_port_mappings_domain RENAME TO idx_port_mappings_domain;
ALTER TABLE host_applications DROP CONSTRAINT IF EXISTS host_applications_host_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_host_applications_host_id ON host_applications (host_id);

-- 4) 列类型按实库对齐（实库 bigint；旧迁移写 INTEGER，同类型重复执行无害）
ALTER TABLE users ALTER COLUMN failed_attempts TYPE bigint;
ALTER TABLE host_applications ALTER COLUMN host_id TYPE bigint USING host_id::bigint;

-- 5) host_applications 补 created_at / updated_at
ALTER TABLE host_applications ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT now();
ALTER TABLE host_applications ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now();

-- 6) FK 口径对齐实库 NO ACTION（旧迁移写了 ON DELETE CASCADE，Up 块不可变；
--    先 DROP 后 ADD，重跑时同名约束先删再建，幂等）
ALTER TABLE host_applications DROP CONSTRAINT IF EXISTS fk_host_applications_host;
ALTER TABLE host_applications DROP CONSTRAINT IF EXISTS fk_hosts_application;
ALTER TABLE host_applications ADD CONSTRAINT fk_hosts_application
    FOREIGN KEY (host_id) REFERENCES hosts(id);

-- +goose Down
-- 对齐迁移为幂等收敛操作，Down 不还原（还原会重新引入漂移）。
SELECT 1;
