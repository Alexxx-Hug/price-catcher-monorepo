-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS shop;

CREATE TABLE shop.products (
    id BIGSERIAL PRIMARY KEY,
    nm_id BIGINT UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    brand VARCHAR(128),
    url TEXT NOT NULL,
    total_quantity INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS shop.products;

-- +goose StatementEnd