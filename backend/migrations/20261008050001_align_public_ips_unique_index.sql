-- +goose Up
-- 将 goose 内联 UNIQUE 产生的默认约束名，统一为 GORM uniqueIndex 期望的 uni_public_ips_ip
ALTER TABLE public_ips DROP CONSTRAINT IF EXISTS public_ips_ip_key;
CREATE UNIQUE INDEX IF NOT EXISTS uni_public_ips_ip ON public_ips (ip);

-- +goose Down
DROP INDEX IF EXISTS uni_public_ips_ip;
ALTER TABLE public_ips ADD CONSTRAINT public_ips_ip_key UNIQUE (ip);
