-- +goose Up
-- T-049 删除 port_mappings 冗余列 isp/exit_location（传递依赖，值恒从 public_ips 资源池派生；
-- 前端表单/批量文本/xlsx 导入导出均不写入）。读取统一从 public_ips 带出，API 响应字段不变。
ALTER TABLE port_mappings DROP COLUMN IF EXISTS isp;
ALTER TABLE port_mappings DROP COLUMN IF EXISTS exit_location;

-- +goose Down
-- 恢复空列（历史值已随删除丢弃，不回填；资源池为唯一权威来源）。
ALTER TABLE port_mappings ADD COLUMN IF EXISTS isp VARCHAR(128);
ALTER TABLE port_mappings ADD COLUMN IF EXISTS exit_location VARCHAR(128);
