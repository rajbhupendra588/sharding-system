-- Order items table
-- Sharded by user_id (via order.user_id - co-located with orders)
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);

-- Note: Foreign key constraints may not work across shards in a sharded system
-- Consider application-level validation instead
-- ALTER TABLE order_items ADD CONSTRAINT fk_order_items_order_id 
--   FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;
