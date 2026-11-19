-- +goose Up
ALTER TABLE products
ALTER COLUMN price TYPE BIGINT;

-- +goose Down
ALTER TABLE products
ALTER COLUMN price TYPE INTEGER;