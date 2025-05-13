-- Create the symbols table only (no indexes)
CREATE TABLE IF NOT EXISTS symbols (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    exchange_id VARCHAR(30),
    segment VARCHAR(10),
    security_id VARCHAR(30),
    isin VARCHAR(20),
    instrument VARCHAR(30),
    underlying_security_id VARCHAR(30),
    underlying_symbol VARCHAR(30),
    symbol_name VARCHAR(255),
    display_name VARCHAR(255),
    instrument_type VARCHAR(30),
    series VARCHAR(10),
    lot_size DECIMAL(10, 1),
    expiry_date VARCHAR(30),
    strike_price DECIMAL(10, 5),
    option_type VARCHAR(10),
    tick_size DECIMAL(10, 4),
    expiry_flag VARCHAR(5),
    bracket_flag CHAR(1),
    cover_flag CHAR(1),
    asm_gsm_flag CHAR(1),
    asm_gsm_category VARCHAR(10),
    buy_sell_indicator CHAR(1),
    buy_co_min_margin_per DECIMAL(10, 2),
    sell_co_min_margin_per DECIMAL(10, 2),
    buy_co_sl_range_max_perc DECIMAL(10, 2),
    sell_co_sl_range_max_perc DECIMAL(10, 2),
    buy_co_sl_range_min_perc DECIMAL(10, 2),
    sell_co_sl_range_min_perc DECIMAL(10, 2),
    buy_bo_min_margin_per DECIMAL(10, 2),
    sell_bo_min_margin_per DECIMAL(10, 2),
    buy_bo_sl_range_max_perc DECIMAL(10, 2),
    sell_bo_sl_range_max_perc DECIMAL(10, 2),
    buy_bo_sl_range_min_perc DECIMAL(10, 2),
    sell_bo_sl_range_min_perc DECIMAL(10, 2), -- Fixed column name from sell_bo_min_range
    buy_bo_profit_range_max_perc DECIMAL(10, 2),
    sell_bo_profit_range_max_perc DECIMAL(10, 2),
    buy_bo_profit_range_min_perc DECIMAL(10, 2),
    sell_bo_profit_range_min_perc DECIMAL(10, 2),
    mtf_leverage DECIMAL(10, 2),
    custom_symbol VARCHAR(50),
    expiry_code VARCHAR(10),
    trading_symbol VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    last_updated TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add indexes to improve query performance
CREATE INDEX idx_symbols_symbol ON symbols(symbol);
CREATE INDEX idx_symbols_exchange_id ON symbols(exchange_id);
CREATE INDEX idx_symbols_instrument_type ON symbols(instrument_type);
CREATE INDEX idx_symbols_trading_symbol ON symbols(trading_symbol);
CREATE INDEX idx_symbols_is_active ON symbols(is_active);

-- Create a unique constraint for the combination of symbol, exchange_id and instrument_type
CREATE UNIQUE INDEX idx_symbols_unique_identifier ON symbols(symbol, exchange_id, instrument_type) 
WHERE is_active = TRUE;

-- Create index for expiry search
CREATE INDEX idx_symbols_expiry_date ON symbols(expiry_date);

-- Add full text search index for name fields
CREATE INDEX idx_symbols_name_trgm ON symbols USING gin(name gin_trgm_ops);
CREATE INDEX idx_symbols_symbol_name_trgm ON symbols USING gin(symbol_name gin_trgm_ops);
CREATE INDEX idx_symbols_display_name_trgm ON symbols USING gin(display_name gin_trgm_ops);