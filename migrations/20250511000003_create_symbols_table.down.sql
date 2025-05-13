-- Drop the full text search indexes
DROP INDEX IF EXISTS idx_symbols_name_trgm;
DROP INDEX IF EXISTS idx_symbols_symbol_name_trgm;
DROP INDEX IF EXISTS idx_symbols_display_name_trgm;

-- Drop other indexes
DROP INDEX IF EXISTS idx_symbols_expiry_date;
DROP INDEX IF EXISTS idx_symbols_unique_identifier;
DROP INDEX IF EXISTS idx_symbols_symbol;
DROP INDEX IF EXISTS idx_symbols_exchange_id;
DROP INDEX IF EXISTS idx_symbols_instrument_type;
DROP INDEX IF EXISTS idx_symbols_trading_symbol;
DROP INDEX IF EXISTS idx_symbols_is_active;

-- Drop the symbols table
DROP TABLE IF EXISTS symbols;