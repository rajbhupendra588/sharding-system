-- Orders table
-- Sharded by user_id (ensures co-location with users)
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    shipping_address TEXT NOT NULL,
    payment_method VARCHAR(100),
    tracking_number VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- Note: Foreign key constraints may not work across shards in a sharded system
-- Since user_id is the shard key, user and orders are co-located on same shard
-- Consider application-level validation instead
-- ALTER TABLE orders ADD CONSTRAINT fk_orders_user_id 
--   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

