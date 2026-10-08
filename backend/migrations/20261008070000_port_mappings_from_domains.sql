-- +goose Up
CREATE TABLE IF NOT EXISTS port_mappings (
    id SERIAL PRIMARY KEY,
    public_ip VARCHAR(45) NOT NULL,
    host_id BIGINT NOT NULL REFERENCES hosts(id),
    external_ports TEXT NOT NULL DEFAULT '',
    internal_ports TEXT NOT NULL DEFAULT '',
    domain VARCHAR(255),
    isp VARCHAR(128),
    exit_location VARCHAR(128),
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uni_port_mappings_domain
    ON port_mappings (domain)
    WHERE domain IS NOT NULL AND domain <> '';
CREATE INDEX IF NOT EXISTS idx_port_mappings_host_id ON port_mappings (host_id);

-- 迁移存量 domains：单端口 host_port 拆成外网=内网=该端口；无主机关联的行丢弃
INSERT INTO port_mappings (
    public_ip, host_id, external_ports, internal_ports,
    domain, isp, exit_location, remark, created_at, updated_at
)
SELECT
    COALESCE(public_ip, ''),
    host_id,
    CASE WHEN host_port > 0 THEN host_port::text ELSE '' END,
    CASE WHEN host_port > 0 THEN host_port::text ELSE '' END,
    NULLIF(TRIM(domain), ''),
    isp,
    exit_location,
    remark,
    COALESCE(created_at, NOW()),
    COALESCE(updated_at, NOW())
FROM domains
WHERE host_id IS NOT NULL AND host_id > 0;

-- 由映射台账重算 ip_mapped
UPDATE hosts h
SET ip_mapped = EXISTS (
    SELECT 1 FROM port_mappings pm WHERE pm.host_id = h.id
);

DROP TABLE IF EXISTS domains;

-- +goose Down
CREATE TABLE IF NOT EXISTS domains (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(255),
    public_ip VARCHAR(45),
    isp VARCHAR(128),
    exit_location VARCHAR(128),
    host_id BIGINT,
    host_port INTEGER,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO domains (
    domain, public_ip, isp, exit_location, host_id, host_port, remark, created_at, updated_at
)
SELECT
    domain,
    public_ip,
    isp,
    exit_location,
    host_id,
    CASE
        WHEN internal_ports ~ '^[0-9]+$' THEN internal_ports::integer
        ELSE NULL
    END,
    remark,
    created_at,
    updated_at
FROM port_mappings;

DROP TABLE IF EXISTS port_mappings;
