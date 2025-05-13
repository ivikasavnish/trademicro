CREATE TABLE IF NOT EXISTS broker_tokens (
    id SERIAL PRIMARY KEY,
    broker VARCHAR(50) NOT NULL,
    token TEXT NOT NULL,
    "user" VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_broker_tokens_user ON broker_tokens("user");
CREATE INDEX idx_broker_tokens_broker ON broker_tokens(broker);