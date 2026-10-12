-- +goose Up
-- T-055 端口映射协议：新增并列协议列（与端口等长逗号并行），存量回填 tcp；
-- 唯一索引纳入外网协议列（同端口不同协议不再互相排斥）。
-- 幂等：列已存在则跳过、回填仅补空值、索引 IF NOT EXISTS（重复执行为 no-op）。
ALTER TABLE port_mappings ADD COLUMN IF NOT EXISTS external_protocols text NOT NULL DEFAULT '';
ALTER TABLE port_mappings ADD COLUMN IF NOT EXISTS internal_protocols text NOT NULL DEFAULT '';

UPDATE port_mappings
SET external_protocols = (SELECT string_agg('tcp', ',') FROM unnest(string_to_array(external_ports, ','))),
    internal_protocols = (SELECT string_agg('tcp', ',') FROM unnest(string_to_array(internal_ports, ',')))
WHERE external_protocols = '' OR internal_protocols = '';

DROP INDEX IF EXISTS idx_port_mappings_host_ip_ports;
CREATE UNIQUE INDEX IF NOT EXISTS idx_port_mappings_host_ip_ports
    ON port_mappings (host_id, public_ip, external_ports, external_protocols);

-- +goose Down
DROP INDEX IF EXISTS idx_port_mappings_host_ip_ports;
CREATE UNIQUE INDEX IF NOT EXISTS idx_port_mappings_host_ip_ports
    ON port_mappings (host_id, public_ip, external_ports);
ALTER TABLE port_mappings DROP COLUMN IF EXISTS external_protocols;
ALTER TABLE port_mappings DROP COLUMN IF EXISTS internal_protocols;
