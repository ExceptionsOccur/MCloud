-- +goose Up
-- T-051 端口映射防重：整组唯一 (host_id, public_ip, external_ports)。
-- 先清历史重复（保留 id 最小的一行），再落唯一索引；重复执行时无重复行可删、索引已存在跳过。
DELETE FROM port_mappings a
USING port_mappings b
WHERE a.host_id = b.host_id
  AND a.public_ip = b.public_ip
  AND a.external_ports = b.external_ports
  AND a.id > b.id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_port_mappings_host_ip_ports
    ON port_mappings (host_id, public_ip, external_ports);

-- +goose Down
DROP INDEX IF EXISTS idx_port_mappings_host_ip_ports;
