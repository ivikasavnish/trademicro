-- Create instruments table for DhanHQ instruments

CREATE TABLE IF NOT EXISTS instruments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Basic symbol identification
    security_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(255),
    symbol_name VARCHAR(255),
    display_name VARCHAR(255),
    
    -- Exchange and market information
    exchange_id VARCHAR(50),
    segment VARCHAR(10),
    segment_type VARCHAR(10),
    isin VARCHAR(50),
    instrument VARCHAR(50),
    instrument_type VARCHAR(50),
    series VARCHAR(20),
    
    -- Contract specifications
    lot_size BIGINT,
    tick_size DECIMAL(18, 4),
    
    -- Derivative specific information
    expiry_date TIMESTAMP,
    expiry_code VARCHAR(10),
    expiry_flag VARCHAR(10),
    strike_price DECIMAL(18, 4),
    option_type VARCHAR(10),
    underlying_security_id VARCHAR(255),
    underlying_symbol VARCHAR(255),
    
    -- Trading restrictions and parameters
    bracket_flag VARCHAR(10),
    cover_flag VARCHAR(10),
    asm_gsm_flag VARCHAR(10),
    asm_gsm_category VARCHAR(50),
    buy_sell_indicator VARCHAR(10),
    
    -- Order margins and limits
    buy_co_min_margin_per DECIMAL(18, 4),
    sell_co_min_margin_per DECIMAL(18, 4),
    buy_co_sl_range_max_perc DECIMAL(18, 4),
    sell_co_sl_range_max_perc DECIMAL(18, 4),
    buy_co_sl_range_min_perc DECIMAL(18, 4),
    sell_co_sl_range_min_perc DECIMAL(18, 4),
    buy_bo_min_margin_per DECIMAL(18, 4),
    sell_bo_min_margin_per DECIMAL(18, 4),
    buy_bo_sl_range_max_perc DECIMAL(18, 4),
    sell_bo_sl_range_max_perc DECIMAL(18, 4),
    buy_bo_sl_range_min_perc DECIMAL(18, 4),
    sell_bo_sl_min_range DECIMAL(18, 4),
    buy_bo_profit_range_max_perc DECIMAL(18, 4),
    sell_bo_profit_range_max_perc DECIMAL(18, 4),
    buy_bo_profit_range_min_perc DECIMAL(18, 4),
    sell_bo_profit_range_min_perc DECIMAL(18, 4),
    mtf_leverage DECIMAL(18, 4),
    
    -- Additional metadata
    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    data_source VARCHAR(20)
);

-- Create indexes for common query patterns
CREATE UNIQUE INDEX IF NOT EXISTS idx_instruments_security_id ON instruments(security_id);
CREATE INDEX IF NOT EXISTS idx_instruments_symbol ON instruments(symbol);
CREATE INDEX IF NOT EXISTS idx_instruments_exchange_id ON instruments(exchange_id);
CREATE INDEX IF NOT EXISTS idx_instruments_segment ON instruments(segment);
CREATE INDEX IF NOT EXISTS idx_instruments_segment_type ON instruments(segment_type);
CREATE INDEX IF NOT EXISTS idx_instruments_isin ON instruments(isin);
CREATE INDEX IF NOT EXISTS idx_instruments_underlying_symbol ON instruments(underlying_symbol);
CREATE INDEX IF NOT EXISTS idx_instruments_expiry_date ON instruments(expiry_date);
CREATE INDEX IF NOT EXISTS idx_instruments_option_type ON instruments(option_type);
CREATE INDEX IF NOT EXISTS idx_instruments_strike_price ON instruments(strike_price);