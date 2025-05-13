#!/usr/bin/env python3
import pandas as pd
import os
from datetime import datetime
import argparse

def parse_args():
    parser = argparse.ArgumentParser(description='Update symbols from detailed CSV format')
    parser.add_argument('--input', default=os.path.expanduser('~/Downloads/api-scrip-master-detailed.csv'),
                       help='Path to detailed CSV file (default: ~/Downloads/api-scrip-master-detailed.csv)')
    parser.add_argument('--output', default='new_symbols.csv',
                       help='Path to output CSV file (default: new_symbols.csv)')
    return parser.parse_args()

def main():
    args = parse_args()
    print(f"Reading detailed symbols from: {args.input}")
    
    # Read the detailed CSV format
    try:
        df = pd.read_csv(args.input)
        print(f"Successfully read {len(df)} symbols from {args.input}")
    except Exception as e:
        print(f"Error reading CSV: {e}")
        return
    
    # Map the columns from detailed format to our database schema
    # Define mapping from detailed CSV format to database columns
    column_mapping = {
        'EXCH_ID': 'exchange_id',
        'SEGMENT': 'segment',
        'SECURITY_ID': 'security_id',
        'ISIN': 'isin',
        'INSTRUMENT': 'instrument',
        'UNDERLYING_SECURITY_ID': 'underlying_security_id',
        'UNDERLYING_SYMBOL': 'underlying_symbol',
        'SYMBOL_NAME': 'symbol_name',
        'DISPLAY_NAME': 'display_name',
        'INSTRUMENT_TYPE': 'instrument_type',
        'SERIES': 'series',
        'LOT_SIZE': 'lot_size',
        'SM_EXPIRY_DATE': 'expiry_date',
        'STRIKE_PRICE': 'strike_price',
        'OPTION_TYPE': 'option_type',
        'TICK_SIZE': 'tick_size',
        'EXPIRY_FLAG': 'expiry_flag',
        'BRACKET_FLAG': 'bracket_flag',
        'COVER_FLAG': 'cover_flag',
        'ASM_GSM_FLAG': 'asm_gsm_flag', 
        'ASM_GSM_CATEGORY': 'asm_gsm_category',
        'BUY_SELL_INDICATOR': 'buy_sell_indicator',
        'BUY_CO_MIN_MARGIN_PER': 'buy_co_min_margin_per',
        'SELL_CO_MIN_MARGIN_PER': 'sell_co_min_margin_per',
        'BUY_CO_SL_RANGE_MAX_PERC': 'buy_co_sl_range_max_perc',
        'SELL_CO_SL_RANGE_MAX_PERC': 'sell_co_sl_range_max_perc',
        'BUY_CO_SL_RANGE_MIN_PERC': 'buy_co_sl_range_min_perc',
        'SELL_CO_SL_RANGE_MIN_PERC': 'sell_co_sl_range_min_perc',
        'BUY_BO_MIN_MARGIN_PER': 'buy_bo_min_margin_per',
        'SELL_BO_MIN_MARGIN_PER': 'sell_bo_min_margin_per', 
        'BUY_BO_SL_RANGE_MAX_PERC': 'buy_bo_sl_range_max_perc',
        'SELL_BO_SL_RANGE_MAX_PERC': 'sell_bo_sl_range_max_perc',
        'BUY_BO_SL_RANGE_MIN_PERC': 'buy_bo_sl_range_min_perc',
        'SELL_BO_SL_MIN_RANGE': 'sell_bo_min_range',
        'BUY_BO_PROFIT_RANGE_MAX_PERC': 'buy_bo_profit_range_max_perc',
        'SELL_BO_PROFIT_RANGE_MAX_PERC': 'sell_bo_profit_range_max_perc',
        'BUY_BO_PROFIT_RANGE_MIN_PERC': 'buy_bo_profit_range_min_perc',
        'SELL_BO_PROFIT_RANGE_MIN_PERC': 'sell_bo_profit_range_min_perc',
        'MTF_LEVERAGE': 'mtf_leverage'
    }
    
    # Create a new DataFrame with renamed columns
    df_renamed = df.rename(columns=column_mapping)
    
    # Add symbol field as symbol_name if it exists, otherwise use DISPLAY_NAME
    df_renamed['symbol'] = df_renamed['symbol_name'].fillna(df_renamed['display_name'])
    
    # Add name field
    df_renamed['name'] = df_renamed['display_name']
    
    # Add is_active field
    df_renamed['is_active'] = True
    
    # Add timestamp fields
    now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    df_renamed['last_updated'] = now
    df_renamed['created_at'] = now
    df_renamed['updated_at'] = now
    
    # Keep only the columns in our database schema
    db_columns = ['symbol', 'name', 'exchange_id', 'segment', 'security_id', 'isin', 'instrument',
                 'underlying_security_id', 'underlying_symbol', 'symbol_name', 'display_name',
                 'instrument_type', 'series', 'lot_size', 'expiry_date', 'strike_price',
                 'option_type', 'tick_size', 'expiry_flag', 'bracket_flag', 'cover_flag',
                 'asm_gsm_flag', 'asm_gsm_category', 'buy_sell_indicator', 'buy_co_min_margin_per',
                 'sell_co_min_margin_per', 'buy_co_sl_range_max_perc', 'sell_co_sl_range_max_perc',
                 'buy_co_sl_range_min_perc', 'sell_co_sl_range_min_perc', 'buy_bo_min_margin_per',
                 'sell_bo_min_margin_per', 'buy_bo_sl_range_max_perc', 'sell_bo_sl_range_max_perc',
                 'buy_bo_sl_range_min_perc', 'sell_bo_min_range', 'buy_bo_profit_range_max_perc',
                 'sell_bo_profit_range_max_perc', 'buy_bo_profit_range_min_perc', 'sell_bo_profit_range_min_perc',
                 'mtf_leverage', 'is_active', 'last_updated', 'created_at', 'updated_at']
    
    # Filter columns and replace NaN values
    final_df = df_renamed.reindex(columns=db_columns).fillna({
        'isin': 'NA',
        'series': 'NA',
        'bracket_flag': 'N',
        'cover_flag': 'N',
        'asm_gsm_flag': 'N',
        'asm_gsm_category': 'NA',
        'buy_sell_indicator': 'A',
        'buy_co_min_margin_per': 0,
        'sell_co_min_margin_per': 0,
        'buy_co_sl_range_max_perc': 0,
        'sell_co_sl_range_max_perc': 0,
        'buy_co_sl_range_min_perc': 0,
        'sell_co_sl_range_min_perc': 0,
        'buy_bo_min_margin_per': 0,
        'sell_bo_min_margin_per': 0,
        'buy_bo_sl_range_max_perc': 0,
        'sell_bo_sl_range_max_perc': 0,
        'buy_bo_sl_range_min_perc': 0,
        'sell_bo_min_range': 0,
        'buy_bo_profit_range_max_perc': 0,
        'sell_bo_profit_range_max_perc': 0,
        'buy_bo_profit_range_min_perc': 0,
        'sell_bo_profit_range_min_perc': 0,
        'mtf_leverage': 0
    })
    
    # Save to CSV
    final_df.to_csv(args.output, index=False)
    print(f"Successfully saved {len(final_df)} symbols to {args.output}")
    
    # Create a backward compatible format for code that depends on old.symbols.csv
    print("Creating backward compatible format for legacy code...")
    compat_df = pd.DataFrame({
        'SEM_EXM_EXCH_ID': df['EXCH_ID'],
        'SEM_SEGMENT': df['SEGMENT'],
        'SEM_SMST_SECURITY_ID': df['SECURITY_ID'],
        'SEM_INSTRUMENT_NAME': df['INSTRUMENT'],
        'SEM_EXPIRY_CODE': 0,  # Default value
        'SEM_TRADING_SYMBOL': df.apply(lambda x: f"{x['SYMBOL_NAME']}-{x['SM_EXPIRY_DATE'][:10]}-{x['STRIKE_PRICE']}-{x['OPTION_TYPE']}" 
                                      if x['INSTRUMENT'] == 'OPTCUR' or x['INSTRUMENT'] == 'OPTIDX' 
                                      else f"{x['SYMBOL_NAME']}-{x['SM_EXPIRY_DATE'][:10]}-FUT" 
                                      if x['INSTRUMENT'] == 'FUTCUR' or x['INSTRUMENT'] == 'FUTIDX'
                                      else x['SYMBOL_NAME'], axis=1),
        'SEM_LOT_UNITS': df['LOT_SIZE'],
        'SEM_CUSTOM_SYMBOL': df['DISPLAY_NAME'],
        'SEM_EXPIRY_DATE': df['SM_EXPIRY_DATE'],
        'SEM_STRIKE_PRICE': df['STRIKE_PRICE'],
        'SEM_OPTION_TYPE': df['OPTION_TYPE'],
        'SEM_TICK_SIZE': df['TICK_SIZE'],
        'SEM_EXPIRY_FLAG': df['EXPIRY_FLAG'],
        'SEM_EXCH_INSTRUMENT_TYPE': df['INSTRUMENT_TYPE'],
        'SEM_SERIES': df['SERIES'],
        'SM_SYMBOL_NAME': df['SYMBOL_NAME']
    })
    
    compat_df.to_csv('old_symbols_compatible.csv', index=False)
    print(f"Successfully created backward compatible format in old_symbols_compatible.csv")
    
    print("Done!")

if __name__ == "__main__":
    main()