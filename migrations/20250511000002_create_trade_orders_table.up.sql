CREATE TABLE IF NOT EXISTS trade_orders (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(30) NOT NULL,
    unit INT NOT NULL,
    diff DECIMAL(10, 4) NOT NULL,
    zag INT NOT NULL,
    type VARCHAR(20) NOT NULL,
    "user" VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_trade_orders_user ON trade_orders("user");
CREATE INDEX idx_trade_orders_symbol ON trade_orders(symbol);
CREATE INDEX idx_trade_orders_status ON trade_orders(status);