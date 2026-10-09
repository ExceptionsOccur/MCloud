-- +goose Up
-- 空库自愈：把 goose 内联 UNIQUE 产生的 PG 默认约束名（*_key），对齐为 GORM uniqueIndex 期望的 idx_ 唯一索引。
-- 不对齐时：gorm ColumnTypes 认为列 UNIQUE（只认约束）、而模型 uniqueIndex 不设 field.Unique，
-- MigrateColumnUnique 按 uni_<table>_<col> 执行 DROP CONSTRAINT（无存在性守卫）→ 42704 → Migrate() 退出。
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);

ALTER TABLE hosts DROP CONSTRAINT IF EXISTS hosts_private_ip_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_hosts_private_ip ON hosts (private_ip);

ALTER TABLE host_applications DROP CONSTRAINT IF EXISTS host_applications_host_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_host_applications_host_id ON host_applications (host_id);

ALTER TABLE cloud_resources DROP CONSTRAINT IF EXISTS cloud_resources_region_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_cloud_resources_region ON cloud_resources (region);

-- public_ips：20261008050001 建的 uni_public_ips_ip 与 GORM 按模型自建的 idx_public_ips_ip 重复，保留 idx_
DROP INDEX IF EXISTS uni_public_ips_ip;
CREATE UNIQUE INDEX IF NOT EXISTS idx_public_ips_ip ON public_ips (ip);

-- 外键去重：删除 goose 侧命名，GORM AutoMigrate 会按约定名（fk_<表>_<关联>）自建唯一外键
ALTER TABLE host_applications DROP CONSTRAINT IF EXISTS fk_host_applications_host;
ALTER TABLE port_mappings DROP CONSTRAINT IF EXISTS port_mappings_host_id_fkey;

-- +goose Down
ALTER TABLE port_mappings ADD CONSTRAINT port_mappings_host_id_fkey FOREIGN KEY (host_id) REFERENCES hosts(id);
ALTER TABLE host_applications ADD CONSTRAINT fk_host_applications_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_public_ips_ip;
CREATE UNIQUE INDEX IF NOT EXISTS uni_public_ips_ip ON public_ips (ip);

DROP INDEX IF EXISTS idx_cloud_resources_region;
ALTER TABLE cloud_resources ADD CONSTRAINT cloud_resources_region_key UNIQUE (region);

DROP INDEX IF EXISTS idx_host_applications_host_id;
ALTER TABLE host_applications ADD CONSTRAINT host_applications_host_id_key UNIQUE (host_id);

DROP INDEX IF EXISTS idx_hosts_private_ip;
ALTER TABLE hosts ADD CONSTRAINT hosts_private_ip_key UNIQUE (private_ip);

DROP INDEX IF EXISTS idx_users_username;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
