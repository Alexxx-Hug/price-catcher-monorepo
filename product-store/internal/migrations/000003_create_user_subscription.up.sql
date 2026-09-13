CREATE TABLE IF NOT EXISTS shop.subscriptions (
    id BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT NOT NULL,
    product_size_id BIGINT NOT NULL REFERENCES shop.product_sizes(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE (telegram_user_id, product_size_id)
);
