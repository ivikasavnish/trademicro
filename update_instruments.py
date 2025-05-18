#!/usr/bin/env python3
"""
Script to update instrument data from DhanHQ API
"""

import argparse
import sys
import os
import pandas as pd
from dhanhq import dhanhq
import logging
import time

# Configure logging
logging.basicConfig(level=logging.INFO,
                   format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger('instrument_updater')

def fetch_instruments(mode='compact', save_file=True, output_file='dhan_instruments.csv',
                     exchange_segment=None, batch_size=1000):
    """
    Fetch instruments from DhanHQ API
    
    Args:
        mode (str): 'compact' or 'detailed'
        save_file (bool): Whether to save the CSV to disk
        output_file (str): File name to save CSV
        exchange_segment (str): Optional exchange segment to filter by
        batch_size (int): Size of batches for processing
        
    Returns:
        pd.DataFrame: DataFrame with instrument data
    """
    try:
        logger.info(f"Initializing DhanHQ client")
        # Initialize DhanHQ client (tokens not needed for public API calls)
        client = dhanhq("", "")
        
        logger.info(f"Fetching instruments in {mode} mode...")
        # Download CSV file using DhanHQ client
        df = client.fetch_security_list(mode=mode, filename=output_file)
        
        logger.info(f"Raw data contains {len(df)} instruments")
        
        # Filter by exchange segment if specified
        if exchange_segment:
            logger.info(f"Filtering by exchange segment: {exchange_segment}")
            # First, create a combined column for easier filtering
            if 'EXCH_ID' in df.columns and 'SEGMENT' in df.columns:
                df['EXCHANGE_SEGMENT'] = df['EXCH_ID'] + '_' + df['SEGMENT']
            elif 'SEM_EXM_EXCH_ID' in df.columns and 'SEM_SEGMENT' in df.columns:
                df['EXCHANGE_SEGMENT'] = df['SEM_EXM_EXCH_ID'] + '_' + df['SEM_SEGMENT']
            
            # Filter the dataframe
            if 'EXCHANGE_SEGMENT' in df.columns:
                df = df[df['EXCHANGE_SEGMENT'].str.contains(exchange_segment, case=False)]
        
        # If no rows found after filtering
        if df.empty:
            logger.warning(f"No instruments found with the specified criteria")
            return df
        
        logger.info(f"After filtering: {len(df)} instruments")
        
        # Save to file if requested
        if save_file:
            logger.info(f"Saving to file: {output_file}")
            df.to_csv(output_file, index=False)
            
        # Return the processed DataFrame
        return df
        
    except Exception as e:
        logger.error(f"Error fetching instruments: {str(e)}")
        raise

def process_in_batches(df, batch_size=1000, db_func=None):
    """
    Process a DataFrame in batches
    
    Args:
        df (pd.DataFrame): DataFrame to process
        batch_size (int): Number of rows to process in each batch
        db_func (callable): Function to call with each batch
    
    Returns:
        dict: Summary of processing statistics
    """
    total_rows = len(df)
    batches = (total_rows + batch_size - 1) // batch_size  # Ceiling division
    
    results = {
        "total_rows": total_rows,
        "processed": 0,
        "created": 0,
        "updated": 0,
        "errors": 0,
        "batches": batches,
        "batch_stats": []
    }
    
    logger.info(f"Processing {total_rows} instruments in {batches} batches of {batch_size}")
    
    for i in range(batches):
        start_idx = i * batch_size
        end_idx = min((i + 1) * batch_size, total_rows)
        batch = df.iloc[start_idx:end_idx]
        
        logger.info(f"Processing batch {i+1}/{batches} with {len(batch)} instruments")
        
        start_time = time.time()
        
        try:
            # If a database function is provided, call it with the batch
            if db_func:
                batch_result = db_func(batch)
                
                # Update results
                results["processed"] += len(batch)
                if isinstance(batch_result, dict):
                    results["created"] += batch_result.get("created", 0)
                    results["updated"] += batch_result.get("updated", 0)
                    results["errors"] += batch_result.get("errors", 0)
            
            # Track batch statistics
            process_time = time.time() - start_time
            results["batch_stats"].append({
                "batch": i+1,
                "rows": len(batch),
                "time_seconds": process_time
            })
            
            logger.info(f"Batch {i+1} processed in {process_time:.2f} seconds")
            
        except Exception as e:
            logger.error(f"Error processing batch {i+1}: {str(e)}")
            results["errors"] += len(batch)
            results["batch_stats"].append({
                "batch": i+1,
                "rows": len(batch),
                "error": str(e)
            })
    
    return results

def list_segments():
    """Print available exchange segments"""
    segments = [
        {"value": "NSE_EQ", "label": "NSE Equity"},
        {"value": "BSE_EQ", "label": "BSE Equity"},
        {"value": "NSE_FNO", "label": "NSE Futures & Options"},
        {"value": "BSE_FNO", "label": "BSE Futures & Options"},
        {"value": "MCX_COMM", "label": "MCX Commodities"},
        {"value": "NSE_CURRENCY", "label": "NSE Currency"},
        {"value": "BSE_CURRENCY", "label": "BSE Currency"}
    ]
    
    print("Available exchange segments:")
    for segment in segments:
        print(f"  {segment['value']}: {segment['label']}")

def export_csv_batches(df, output_dir="instrument_batches", batch_size=1000):
    """
    Export a DataFrame to multiple CSV files in batches
    
    Args:
        df (pd.DataFrame): DataFrame to export
        output_dir (str): Directory to save CSV files
        batch_size (int): Number of rows in each batch
    """
    total_rows = len(df)
    batches = (total_rows + batch_size - 1) // batch_size  # Ceiling division
    
    # Create output directory if it doesn't exist
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    logger.info(f"Exporting {total_rows} instruments to {batches} CSV files in {output_dir}")
    
    for i in range(batches):
        start_idx = i * batch_size
        end_idx = min((i + 1) * batch_size, total_rows)
        batch = df.iloc[start_idx:end_idx]
        
        batch_file = os.path.join(output_dir, f"instruments_batch_{i+1:03d}.csv")
        batch.to_csv(batch_file, index=False)
        
        logger.info(f"Exported batch {i+1}/{batches} with {len(batch)} instruments to {batch_file}")
    
    return batches

def main():
    parser = argparse.ArgumentParser(description='Update instrument data from DhanHQ API')
    parser.add_argument('--mode', choices=['compact', 'detailed'], default='compact',
                       help='Mode to fetch instrument data (compact or detailed)')
    parser.add_argument('--output', default='dhan_instruments.csv',
                       help='Output file name for CSV data')
    parser.add_argument('--exchange-segment', help='Filter by specific exchange segment')
    parser.add_argument('--list-segments', action='store_true', 
                       help='List available exchange segments and exit')
    parser.add_argument('--no-save', action='store_true',
                       help='Do not save to file (output to stdout)')
    parser.add_argument('--batch-size', type=int, default=1000,
                       help='Batch size for processing instruments')
    parser.add_argument('--export-batches', action='store_true',
                       help='Export data in multiple CSV batch files')
    parser.add_argument('--batch-dir', default='instrument_batches',
                       help='Directory for batch CSV files when using --export-batches')
    
    args = parser.parse_args()
    
    # List segments and exit if requested
    if args.list_segments:
        list_segments()
        return 0
    
    try:
        # Fetch instruments
        df = fetch_instruments(
            mode=args.mode, 
            save_file=not args.no_save,
            output_file=args.output,
            exchange_segment=args.exchange_segment,
            batch_size=args.batch_size
        )
        
        if args.export_batches:
            # Export in batches
            batch_count = export_csv_batches(
                df, 
                output_dir=args.batch_dir,
                batch_size=args.batch_size
            )
            print(f"Exported {len(df)} instruments to {batch_count} CSV files in {args.batch_dir}")
        
        # Output summary info
        print(f"Successfully processed {len(df)} instruments")
        if not args.no_save:
            print(f"Data saved to: {args.output}")
            
        # If no-save, output CSV to stdout
        if args.no_save and not args.export_batches:
            print(df.to_csv(index=False))
            
        return 0
            
    except Exception as e:
        logger.error(f"Error in main: {str(e)}")
        return 1

if __name__ == "__main__":
    sys.exit(main())