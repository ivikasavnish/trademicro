CREATE TABLE IF NOT EXISTS favorite_symbols (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    symbol_id INTEGER NOT NULL,
    notes TEXT,
    category VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE,
    UNIQUE (user_id, symbol_id)
);

-- Create indexes
CREATE INDEX idx_favorite_symbols_user_id ON favorite_symbols(user_id);
CREATE INDEX idx_favorite_symbols_symbol_id ON favorite_symbols(symbol_id);
CREATE INDEX idx_favorite_symbols_category ON favorite_symbols(category);