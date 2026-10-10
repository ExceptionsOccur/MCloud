-- +goose Up
-- 审计日志（T-047）：只增不删；仅普通索引、禁用 UNIQUE 约束、不设 users 外键（操作人以快照落库）
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    operator_id BIGINT NOT NULL,
    operator_name VARCHAR(64) NOT NULL DEFAULT '',
    action VARCHAR(32) NOT NULL,
    resource_type VARCHAR(32) NOT NULL,
    resource_id BIGINT,
    detail JSONB,
    request_id VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_operator_id ON audit_logs (operator_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs (resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs (resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs (request_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
