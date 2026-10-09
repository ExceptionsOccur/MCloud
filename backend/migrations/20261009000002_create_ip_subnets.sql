-- +goose Up
-- ip_subnets 此前无迁移 SQL（仅靠启动时 AutoMigrate 建表），补齐以对齐「模型变更必须随附迁移」口径。
-- 索引名 idx_ip_subnets_c_id_r 为 GORM 对字段 CIDR 的实际命名（IndexName 走 toDBName(CIDR)=c_id_r），与存量库一致。
CREATE TABLE IF NOT EXISTS ip_subnets (
    id SERIAL PRIMARY KEY,
    cidr VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ip_subnets_c_id_r ON ip_subnets (cidr);

-- +goose Down
DROP TABLE IF EXISTS ip_subnets;
