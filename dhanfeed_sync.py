import asyncio
import json
import logging
from dhanhq import marketfeed
import pandas as pd
from datetime import datetime, time, date
import pytz
from motor.motor_asyncio import AsyncIOMotorClient
from redis import Redis
from bson import ObjectId
from pymongo import MongoClient
import threading
import queue
import websockets

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - [%(threadName)s] - %(name)s - %(filename)s:%(lineno)d - %(funcName)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

# Constants
SYMBOLS_PER_BATCH = 100
RECONNECT_BASE_DELAY = 5
RECONNECT_MAX_DELAY = 60
END_TIME = time(15, 30)  # 3:30 PM
QUEUE_SIZE = 1000
BATCH_SIZE = 1000

# Configuration
MONGO_URI = "mongodb://root:rootpassword@localhost:27017/?authMechanism=DEFAULT"
REDIS_URI = "redis://localhost:6379"

# Initialize clients
redis_client = Redis.from_url(REDIS_URI)
mongo_client = MongoClient(MONGO_URI)
db = mongo_client['market_data']
collection = db['sym_quote']

# Load and process data
data = pd.read_csv("old.symbols.csv", low_memory=False)
nse_data = data[data.SEM_EXM_EXCH_ID == "NSE"]
equity_data = nse_data[nse_data.SEM_SEGMENT == "E"]
equity_series = equity_data[equity_data.SEM_SERIES == "EQ"]
index_data = nse_data[nse_data.SEM_SEGMENT == "D"]
nse200 = pd.read_csv("MW-NIFTY-200-03-Oct-2024.csv")
nse200.columns = nse200.columns.str.strip()
merged_df = equity_series.merge(nse200, left_on='SEM_TRADING_SYMBOL', right_on='SYMBOL', how='left')
merged_df = merged_df[merged_df['LTP'].notna()]
bankniftymonthly = index_data[(index_data.SEM_INSTRUMENT_NAME == 'OPTIDX') & (index_data.SEM_EXPIRY_FLAG == "M") & (index_data.SEM_TRADING_SYMBOL.str.contains("BANKNIFTY"))]
niftymonthly = index_data[(index_data.SEM_INSTRUMENT_NAME == 'OPTIDX') & (index_data.SEM_EXPIRY_FLAG == "M") & (index_data.SEM_TRADING_SYMBOL.str.startswith("NIFTY"))]

def get_security_id(symbol):
    result = equity_series[equity_series.SEM_TRADING_SYMBOL == symbol]
    return str(result.iloc[0].SEM_SMST_SECURITY_ID) if not result.empty else None

def get_ltp_from_redis(security_id):
    ltp = redis_client.get(f"ltp:{security_id}")
    return float(ltp) if ltp else None

def get_symbol_ltp(symbol):
    security_id = get_security_id(symbol)
    return get_ltp_from_redis(security_id) if security_id else None

# Load user data
users = pd.read_csv("users.csv").to_dict("records")
user = "sonam"
CLIENT_ID = next((str(u.get("clientid")) for u in users if u.get("name") == user), None)
ACCESS_TOKEN = next((str(u.get("token")) for u in users if u.get("name") == user), None)

if not CLIENT_ID or not ACCESS_TOKEN:
    logger.error("User not found or credentials missing")
    exit(1)

# Prepare security list
items = equity_series.SEM_SMST_SECURITY_ID.to_list()
sec_list = [(marketfeed.NSE, str(i), marketfeed.Quote) for i in items]
sec_list.append((marketfeed.NSE, "27066", marketfeed.Quote))
nifty_series = [(marketfeed.NSE_FNO, str(i), marketfeed.Quote) for i in niftymonthly.SEM_SMST_SECURITY_ID]
bank_nifty_series = [(marketfeed.NSE_FNO, str(i), marketfeed.Ticker) for i in bankniftymonthly.SEM_SMST_SECURITY_ID]

def process_quote_data(quote_data):
    if 'LTT' in quote_data:
        ltt_time = datetime.strptime(quote_data['LTT'], '%H:%M:%S').time()
        quote_data['LTT'] = datetime.combine(date.today(), ltt_time)
    
    for field in ['LTP', 'avg_price', 'open', 'close', 'high', 'low']:
        if field in quote_data:
            quote_data[field] = float(quote_data[field])
    
    return quote_data

def gettimestamp():
    now = datetime.now()
    return now.hour * 60 + now.minute

def save_ltp_to_redis(security_id, ltp):
    redis_client.set(f"ltp:{security_id}", str(ltp))
    redis_client.set(f"{security_id}", ltp)

def process_data_thread(data_queue):
    while True:
        try:
            response = data_queue.get(timeout=1)
            processed_data = process_quote_data(response)
            
            if 'LTP' in processed_data and 'security_id' in processed_data:
                save_ltp_to_redis(processed_data['security_id'], processed_data['LTP'])
                logger.info(f"Saved LTP for {processed_data['security_id']}: {processed_data['LTP']}")
            
            # logger.info(f"Processed data: {processed_data}")
        except queue.Empty:
            pass
        except Exception as e:
            logger.error(f"An error occurred in process_data: {e}", exc_info=True)

async def main():
    data_queue = queue.Queue()
    
    processing_thread = threading.Thread(target=process_data_thread, args=(data_queue,))
    processing_thread.daemon = True
    processing_thread.start()

    data = marketfeed.DhanFeed(CLIENT_ID, ACCESS_TOKEN, sec_list)
    
    while gettimestamp() < 930:
        try:
            await data.connect()
            while True:
                response = await data.get_data()
                data_queue.put(response)
        except websockets.exceptions.ConnectionClosed:
            logger.error("WebSocket connection closed. Reconnecting...")
            await asyncio.sleep(RECONNECT_BASE_DELAY)
        except Exception as e:
            logger.error(f"An error occurred in main thread: {e}", exc_info=True)
            await asyncio.sleep(RECONNECT_BASE_DELAY)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Shutting down gracefully...")
    except Exception as e:
        logger.error(f"An error occurred: {e}", exc_info=True)