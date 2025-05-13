#!/usr/bin/env python
import os
import csv
import psycopg2
from datetime import datetime

# Database connection parameters - consider using environment variables
DB_URL = os.getenv("POSTGRES_URL", "postgresql://postgres:password@localhost:5432/trademicro")
CSV_FILE_PATH = os.path.expanduser("~/Downloads/api-scrip-master-detailed.csv")

def seed_symbols():
    print(f"Connecting to database: {DB_URL}")
    conn = psycopg2.connect(DB_URL)
    cursor = conn.cursor()
    
    # Truncate the symbols table to start fresh
    cursor.execute("TRUNCATE TABLE symbols RESTART IDENTITY;")
    
    # Read the CSV file and insert data
    with open(CSV_FILE_PATH, 'r') as csv_file:
        csv_reader = csv.DictReader(csv_file)
        inserted_count = 0
        
        for row in csv_reader:
            try:
                # Map CSV columns to database columns
                cursor.execute("""
                INSERT INTO symbols (
                    symbol, name, exchange_id, segment, security_id, isin, instrument,
                    underlying_security_id, underlying_symbol, symbol_name, display_name,
                    instrument_type, series, lot_size, expiry_date, strike_price,
                    option_type, tick_size, expiry_flag, bracket_flag, cover_flag,
                    asm_gsm_flag, asm_gsm_category, buy_sell_indicator,
                    buy_co_min_margin_per, sell_co_min_margin_per,
                    buy_co_sl_range_max_perc, sell_co_sl_range_max_perc,
                    buy_co_sl_range_min_perc, sell_co_sl_range_min_perc,
                    buy_bo_min_margin_per, sell_bo_min_margin_per,
                    buy_bo_sl_range_max_perc, sell_bo_sl_range_max_perc,
                    buy_bo_sl_range_min_perc, sell_bo_min_range,
                    buy_bo_profit_range_max_perc, sell_bo_profit_range_max_perc,
                    buy_bo_profit_range_min_perc, sell_bo_profit_range_min_perc,
                    mtf_leverage, trading_symbol, custom_symbol, is_active, last_updated
                ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, 
                          %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, 
                          %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s);
                """, (
                    row.get("SYMBOL_NAME", ""), 
                    row.get("DISPLAY_NAME", ""),
                    row.get("EXCH_ID", ""),
                    row.get("SEGMENT", ""),
                    row.get("SECURITY_ID", ""),
                    row.get("ISIN", ""),
                    row.get("INSTRUMENT", ""),
                    row.get("UNDERLYING_SECURITY_ID", ""),
                    row.get("UNDERLYING_SYMBOL", ""),
                    row.get("SYMBOL_NAME", ""),
                    row.get("DISPLAY_NAME", ""),
                    row.get("INSTRUMENT_TYPE", ""),
                    row.get("SERIES", ""),
                    float(row.get("LOT_SIZE", 0)) if row.get("LOT_SIZE") else 0,
                    row.get("SM_EXPIRY_DATE", ""),
                    float(row.get("STRIKE_PRICE", -0.01)) if row.get("STRIKE_PRICE") else -0.01,
                    row.get("OPTION_TYPE", ""),
                    float(row.get("TICK_SIZE", 0)) if row.get("TICK_SIZE") else 0,
                    row.get("EXPIRY_FLAG", ""),
                    row.get("BRACKET_FLAG", ""),
                    row.get("COVER_FLAG", ""),
                    row.get("ASM_GSM_FLAG", ""),
                    row.get("ASM_GSM_CATEGORY", ""),
                    row.get("BUY_SELL_INDICATOR", ""),
                    float(row.get("BUY_CO_MIN_MARGIN_PER", 0)) if row.get("BUY_CO_MIN_MARGIN_PER") else 0,
                    float(row.get("SELL_CO_MIN_MARGIN_PER", 0)) if row.get("SELL_CO_MIN_MARGIN_PER") else 0,
                    float(row.get("BUY_CO_SL_RANGE_MAX_PERC", 0)) if row.get("BUY_CO_SL_RANGE_MAX_PERC") else 0,
                    float(row.get("SELL_CO_SL_RANGE_MAX_PERC", 0)) if row.get("SELL_CO_SL_RANGE_MAX_PERC") else 0,
                    float(row.get("BUY_CO_SL_RANGE_MIN_PERC", 0)) if row.get("BUY_CO_SL_RANGE_MIN_PERC") else 0,
                    float(row.get("SELL_CO_SL_RANGE_MIN_PERC", 0)) if row.get("SELL_CO_SL_RANGE_MIN_PERC") else 0,
                    float(row.get("BUY_BO_MIN_MARGIN_PER", 0)) if row.get("BUY_BO_MIN_MARGIN_PER") else 0,
                    float(row.get("SELL_BO_MIN_MARGIN_PER", 0)) if row.get("SELL_BO_MIN_MARGIN_PER") else 0,
                    float(row.get("BUY_BO_SL_RANGE_MAX_PERC", 0)) if row.get("BUY_BO_SL_RANGE_MAX_PERC") else 0,
                    float(row.get("SELL_BO_SL_RANGE_MAX_PERC", 0)) if row.get("SELL_BO_SL_RANGE_MAX_PERC") else 0,
                    float(row.get("BUY_BO_SL_RANGE_MIN_PERC", 0)) if row.get("BUY_BO_SL_RANGE_MIN_PERC") else 0,
                    float(row.get("SELL_BO_MIN_RANGE", 0)) if row.get("SELL_BO_MIN_RANGE") else 0,
                    float(row.get("BUY_BO_PROFIT_RANGE_MAX_PERC", 0)) if row.get("BUY_BO_PROFIT_RANGE_MAX_PERC") else 0,
                    float(row.get("SELL_BO_PROFIT_RANGE_MAX_PERC", 0)) if row.get("SELL_BO_PROFIT_RANGE_MAX_PERC") else 0,
                    float(row.get("BUY_BO_PROFIT_RANGE_MIN_PERC", 0)) if row.get("BUY_BO_PROFIT_RANGE_MIN_PERC") else 0,
                    float(row.get("SELL_BO_PROFIT_RANGE_MIN_PERC", 0)) if row.get("SELL_BO_PROFIT_RANGE_MIN_PERC") else 0,
                    float(row.get("MTF_LEVERAGE", 0)) if row.get("MTF_LEVERAGE") else 0,
                    row.get("SYMBOL_NAME", "") + "-" + row.get("INSTRUMENT_TYPE", ""),  # Custom trading symbol
                    row.get("SYMBOL_NAME", "") + " " + row.get("INSTRUMENT_TYPE", ""),  # Custom symbol
                    True,  # is_active
                    datetime.now()  # last_updated
                ))
                inserted_count += 1
                if inserted_count % 100 == 0:
                    print(f"Inserted {inserted_count} symbols...")
                    
            except Exception as e:
                print(f"Error inserting row: {row}")
                print(f"Error details: {str(e)}")
                conn.rollback()
                continue
    
    # Commit the transaction
    conn.commit()
    cursor.close()
    conn.close()
    
    print(f"Successfully seeded {inserted_count} symbols.")

if __name__ == "__main__":
    # Make sure the scripts directory exists
    os.makedirs(os.path.dirname(os.path.abspath(__file__)), exist_ok=True)
    
    seed_symbols()