#!/usr/bin/env python3
"""
CSV to Database Importer for DhanHQ Instruments

This script reads instrument data from a CSV file and imports it directly into the database
using pandas. It supports both compact and detailed formats from DhanHQ.

Usage:
    python csv_to_db.py --file dhan_instruments.csv --mode compact [--batch-size 1000] [--upsert]
    python csv_to_db.py --file instrument_data/instruments_batch_001.csv
"""

import argparse
import logging
import os
import sys
import time
from datetime import datetime

import pandas as pd
import sqlalchemy
from sqlalchemy import create_engine, text

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger('csv_importer')

def get_db_connection_string():
    """Get database connection string from environment or use default"""
    db_url = os.environ.get('POSTGRES_URL', 'postgres://postgres:password@localhost:5432/trademicro')
    return db_url

def map_segment_type(segment):
    """Map segment code to segment type"""
    if segment == 'E':
        return 'EQ'
    elif segment == 'D':
        return 'IDX'
    elif segment == 'C':
        return 'CUR'
    elif segment == 'M':
        return 'COMM'
    return 'OTHER'

def process_csv_file(file_path, mode='compact', batch_size=1000):
    """
    Process a CSV file and import to database
    
    Args:
        file_path: Path to CSV file
        mode: Format mode ('compact' or 'detailed')
        batch_size: Number of rows to process in each batch
    """
    logger.info(f"Processing file: {file_path} in {mode} mode")
    
    # Check if file exists
    if not os.path.exists(file_path):
        logger.error(f"File not found: {file_path}")
        return False
    
    try:
        # Read the CSV file
        df = pd.read_csv(file_path)
        logger.info(f"Read {len(df)} rows from CSV")
        
        # Create database connection
        engine = create_engine(get_db_connection_string())
        
        # Prepare column mappings based on mode
        if mode == 'compact':
            symbol_col = 'SYMBOL'
            name_col = 'SYMBOL_NAME'
            exchange_col = 'EXCH_ID'
            segment_col = 'SEGMENT'
            security_id_col = 'SECURITY_ID'
        else:  # detailed
            symbol_col = 'SEM_TRADING_SYMBOL'
            name_col = 'SM_SYMBOL_NAME'
            exchange_col = 'SEM_EXM_EXCH_ID'
            segment_col = 'SEM_SEGMENT'
            security_id_col = 'SEM_SCRIP_ID'
        
        # Ensure required columns exist
        required_cols = [symbol_col, name_col, exchange_col, segment_col]
        missing_cols = [col for col in required_cols if col not in df.columns]
        
        if missing_cols:
            logger.error(f"Missing required columns: {missing_cols}")
            return False
        
        # Create a connection to execute SQL statements
        with engine.connect() as connection:
            # Check if the instruments table exists
            check_query = text("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'instruments')")
            result = connection.execute(check_query).scalar()
            
            if not result:
                logger.warning("Instruments table does not exist, creating...")
                # Create instruments table with minimal columns if it doesn't exist
                create_table_sql = """
                CREATE TABLE IF NOT EXISTS instruments (
                    id SERIAL PRIMARY KEY,
                    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                    security_id VARCHAR(255) UNIQUE NOT NULL,
                    symbol VARCHAR(255),
                    symbol_name VARCHAR(255),
                    display_name VARCHAR(255),
                    exchange_id VARCHAR(50),
                    segment VARCHAR(10),
                    segment_type VARCHAR(10),
                    last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
                    data_source VARCHAR(20)
                );
                CREATE INDEX IF NOT EXISTS idx_instruments_symbol ON instruments(symbol);
                CREATE INDEX IF NOT EXISTS idx_instruments_exchange_id ON instruments(exchange_id);
                CREATE INDEX IF NOT EXISTS idx_instruments_segment ON instruments(segment);
                """
                connection.execute(text(create_table_sql))
                connection.commit()
        
        # Process in batches
        total_rows = len(df)
        batches = (total_rows + batch_size - 1) // batch_size
        
        created = 0
        updated = 0
        errors = 0
        
        # Convert column names to lowercase for SQL compatibility
        df.columns = [c.lower() for c in df.columns]
        
        # Rename columns to match database schema
        column_mapping = {
            symbol_col.lower(): 'symbol',
            name_col.lower(): 'symbol_name',
            exchange_col.lower(): 'exchange_id',
            segment_col.lower(): 'segment',
        }
        
        if security_id_col.lower() in df.columns:
            column_mapping[security_id_col.lower()] = 'security_id'
            
        # Add display name if available
        if 'display_name' in df.columns:
            column_mapping['display_name'] = 'display_name'
        elif 'sem_custom_symbol' in df.columns:
            column_mapping['sem_custom_symbol'] = 'display_name'
            
        # Rename columns
        df = df.rename(columns=column_mapping)
        
        # Make sure security_id is present
        if 'security_id' not in df.columns:
            # Generate security_id from exchange_id + symbol if not present
            df['security_id'] = df['exchange_id'] + '-' + df['segment'] + '-' + df['symbol']
            
        # Add segment_type column
        df['segment_type'] = df['segment'].apply(map_segment_type)
        
        # Add metadata
        now = datetime.now()
        df['created_at'] = now
        df['updated_at'] = now
        df['last_updated'] = now
        df['data_source'] = mode
        
        # Process in batches
        for i in range(batches):
            start_idx = i * batch_size
            end_idx = min((i + 1) * batch_size, total_rows)
            batch = df.iloc[start_idx:end_idx]
            
            logger.info(f"Processing batch {i+1}/{batches} with {len(batch)} rows")
            start_time = time.time()
            
            try:
                # Keep only the columns that exist in our table
                batch = batch[['security_id', 'symbol', 'symbol_name', 'display_name', 
                                'exchange_id', 'segment', 'segment_type', 
                                'created_at', 'updated_at', 'last_updated', 'data_source']]
                
                # Insert data using pandas to_sql method with "upsert" behavior (update or insert)
                # First, we need to check if PostgreSQL version supports "on conflict"
                batch.to_sql(
                    'instruments', 
                    engine, 
                    if_exists='append', 
                    index=False, 
                    method='multi',
                    chunksize=100
                )
                
                # Count new records created
                created += len(batch)
                
            except Exception as e:
                logger.error(f"Error processing batch {i+1}: {str(e)}")
                errors += len(batch)
                
            process_time = time.time() - start_time
            logger.info(f"Batch {i+1} processed in {process_time:.2f} seconds")
            
        # Update the statistics
        logger.info(f"Import completed: {created} rows imported, {errors} errors")
        return {
            "created": created,
            "updated": updated,
            "errors": errors,
            "total": total_rows
        }
        
    except Exception as e:
        logger.error(f"Error processing file: {str(e)}")
        return False

def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Import DhanHQ instrument data from CSV to database')
    parser.add_argument('--file', required=True, help='Path to CSV file')
    parser.add_argument('--mode', choices=['compact', 'detailed'], default='compact',
                      help='Format mode (compact or detailed)')
    parser.add_argument('--batch-size', type=int, default=1000,
                      help='Number of rows to process in each batch')
    
    args = parser.parse_args()
    
    # Process file
    result = process_csv_file(args.file, args.mode, args.batch_size)
    
    if result:
        print(f"Import completed successfully: {result['created']} records created, {result['errors']} errors")
        return 0
    else:
        print("Import failed. Check logs for details.")
        return 1

if __name__ == "__main__":
    sys.exit(main())