-- +goose Up
-- T-072：hosts.disk 实为对象存储，字段名与显示名称统一改为 object_storage
ALTER TABLE hosts RENAME COLUMN disk TO object_storage;

-- +goose Down
ALTER TABLE hosts RENAME COLUMN object_storage TO disk;
