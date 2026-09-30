-- +goose Up
CREATE TABLE IF NOT EXISTS persons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    contact VARCHAR(64),
    unit VARCHAR(128),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS persons;
