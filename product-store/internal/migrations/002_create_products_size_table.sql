-- +goose Up
-- +goose StatementBegin

CREATE TABLE shop.product_sizes (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES shop.products(id) ON DELETE CASCADE,
    option_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    orig_name VARCHAR(64),
    price_minor INT,
    quantity INT NOT NULL DEFAULT 0,
    in_stock BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    UNIQUE(product_id, option_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS shop.product_sizes;

-- +goose StatementEnd