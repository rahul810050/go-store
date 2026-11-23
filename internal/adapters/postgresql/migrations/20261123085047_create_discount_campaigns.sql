-- +goose Up
CREATE TABLE discount_campaigns (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value INTERGER NOT NULL CHECK (discount_value > 0),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_at > start_at),
    CHECK (discount_type != 'percentage' OR discount_value <= 100)
);

CREATE TABLE discount_campaigns_products (
    campaign_id BIGINT NOT NULL REFERENCES discount_campaigns(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    PRIMARY KEY (campaign_id, product_id)
);

CREATE INDEX idx_discount_campaign_products_product
ON discount_campaigns_products(product_id);

-- +goose Down
DROP TABLE IF EXISTS discount_campaigns_products;
DROP TABLE IF EXISTS discount_campaigns;