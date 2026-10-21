-- +goose Up
-- T-080：对象存储统一以技术信息 hosts.object_storage 为准，申请信息不再重复存储
-- 存量校验：dev 库 host_applications.object_storage_size 全为空（无数据丢失）
ALTER TABLE host_applications DROP COLUMN IF EXISTS object_storage_size;

-- +goose Down
ALTER TABLE host_applications ADD COLUMN IF NOT EXISTS object_storage_size VARCHAR(32);
