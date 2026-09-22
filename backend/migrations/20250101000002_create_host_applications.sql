-- +goose Up
CREATE TABLE IF NOT EXISTS host_applications (
    id SERIAL PRIMARY KEY,
    host_id INTEGER NOT NULL UNIQUE,
    apply_unit VARCHAR(128),
    applicant VARCHAR(64),
    applicant_contact VARCHAR(64),
    project VARCHAR(128),
    apply_reason TEXT,
    apply_config TEXT,
    apply_time VARCHAR(32),
    object_storage_size VARCHAR(32),
    remark TEXT,
    CONSTRAINT fk_host_applications_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS host_applications;
