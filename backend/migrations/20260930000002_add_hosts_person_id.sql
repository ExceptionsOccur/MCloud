-- +goose Up
ALTER TABLE hosts ADD COLUMN IF NOT EXISTS person_id BIGINT;
ALTER TABLE hosts DROP CONSTRAINT IF EXISTS fk_hosts_person;
ALTER TABLE hosts ADD CONSTRAINT fk_hosts_person FOREIGN KEY (person_id) REFERENCES persons(id);
CREATE INDEX IF NOT EXISTS idx_hosts_person_id ON hosts(person_id);

-- +goose Down
DROP INDEX IF EXISTS idx_hosts_person_id;
ALTER TABLE hosts DROP CONSTRAINT IF EXISTS fk_hosts_person;
ALTER TABLE hosts DROP COLUMN IF EXISTS person_id;
