import threading
import datetime
import logging
import uuid
import pandas as pd
from collections import deque, Counter
import redis
import time

# --- Utility and Data Setup (from trade_log.py) ---
redis_host = 'localhost'
redis_port = 6379
r = redis.Redis(host=redis_host, port=redis_port, db=0)

data = pd.read_csv("old.symbols.csv")
nse_data = data[data.SEM_EXM_EXCH_ID == "NSE"]
equity_data = nse_data[nse_data.SEM_SEGMENT == "E"]
equity_series = equity_data[equity_data.SEM_SERIES == "EQ"]
symbol_to_token = {row.SEM_TRADING_SYMBOL.upper(): str(row.SEM_SMST_SECURITY_ID) for _, row in equity_series.iterrows()}

import random

def get_ltp(symbol, mock=False):
    if mock:
        # Return a random plausible price for demo
        return round(random.uniform(100, 500), 2)
    instrument_token = symbol_to_token.get(symbol.upper())
    if instrument_token:
        ltp_data = r.get(f"ltp:{instrument_token}")
        if ltp_data:
            return float(ltp_data)
    return None

def get_today():
    return datetime.date.today().isoformat()

# --- Trading Logic Classes (from trade_log.py, refactored) ---
class OrderItem:
    def __init__(self, symbol, security_id, quantity, order_type, product, price, transaction_type, dhan=None):
        self.uuid = str(uuid.uuid4())
        self.symbol = symbol
        self.security_id = security_id
        self.quantity = quantity
        self.order_type = order_type
        self.product = product
        self.price = price
        self.transaction_type = transaction_type
        self.action = None
        self.orderId = None
        self.status = None
        self.reverse = None
        self.reverse_order = None
        self.reverse_status = None
        self.be_closed = False
        self.order_at = None
        self.close_at = None
        self.dhan = dhan  # Dhan API client

    def execute(self):
        logging.info(f"Executing order: {self}")
        if getattr(self, 'mock', False):
            # Simulate order placement
            time.sleep(0.5)
            self.action = {"status": "success", "data": {"orderId": str(uuid.uuid4()), "orderStatus": "MOCKED"}}
            self.orderId = self.action["data"]["orderId"]
            self.status = self.action["data"]["orderStatus"]
            self.order_at = int(time.time())
            logging.info(f"[MOCK] Order placed: {self.orderId}")
            return
        if self.dhan is None:
            logging.warning("No Dhan client provided, skipping order execution.")
            self.action = {"status": "skipped"}
            self.be_closed = True
            return
        try:
            time.sleep(1)
            self.action = self.dhan.place_order(
                security_id=self.security_id, exchange_segment="NSE_EQ",
                transaction_type=self.transaction_type, quantity=self.quantity,
                order_type=self.order_type, product_type=self.product, price=self.price)
            if self.action.get("status") == "success":
                logging.info("Order placed successfully")
                self.orderId = self.action.get("data").get("orderId")
                self.status = self.action.get("data").get("orderStatus")
                self.order_at = int(time.time())
                r.rpush("orders", self.orderId)
            else:
                logging.error("Order not placed")
                logging.error(self.action)
                self.be_closed = True
        except Exception as e:
            logging.error(f'Error executing order: {e}')
            self.be_closed = True


    def confirm(self):
        if self.status != "TRADED" and self.reverse is None:
            oid = self.orderId
            if r.get(oid) is None:
                if self.close_at is None:
                    self.close_at = int(time.time()) + 300
                elif int(time.time()) > self.close_at:
                    self.be_closed = True
                return
            if r.get(self.orderId) == b"TRADED":
                logging.info("Order confirmed")
                self.status = "TRADED"
                self.exit()
            if r.get(self.orderId) == b"CANCELLED":
                logging.info("Order cancelled")
                self.be_closed = True
        if self.status == "TRANSIT" and self.reverse is None:
            if self.order_at and self.order_at + 1 <= int(time.time()):
                logging.info("Order not confirmed, cancelling order")
                self.cancel()

    def cancel(self):
        if getattr(self, 'mock', False):
            logging.info(f"[MOCK] Cancelled order: {self.orderId}")
            self.be_closed = True
            return
        if self.dhan is None or not self.orderId:
            return
        time.sleep(1)
        order = self.dhan.cancel_order(self.orderId)
        logging.info(order)
        if order.get("status") == "success":
            self.be_closed = True

    def exit(self):
        exit_price = round(self.price * 1.007, 1)
        self.reverse_order = OrderItem(
            symbol=self.symbol, security_id=self.security_id, quantity=self.quantity,
            order_type="LIMIT", product=self.product, price=exit_price,
            transaction_type="SELL", dhan=self.dhan)
        try:
            self.reverse_order.execute()
            self.reverse = self.reverse_order.orderId
        except Exception as e:
            logging.error(f'Error creating exit order: {e}')

    def __str__(self):
        return str(self.__dict__)

class Bucket:
    def __init__(self, symbol, productType, quantity=1, orderType="LIMIT", price=0, transaction_type="BUY", diff=0.5, zag=1, dhan=None, mock=False):
        self.items = []
        self.symbol = symbol
        self.quantity = quantity
        self.orderType = orderType
        self.productType = productType
        self.price = price
        self.transaction_type = transaction_type
        self.securityId = symbol_to_token.get(symbol.upper())
        self.diff = max(diff/100, 0.001)
        self.zag = zag
        self.dhan = dhan
        self.mock = mock

    def entry(self):
        price = get_ltp(self.symbol, mock=self.mock)
        if price is None:
            logging.info("Price not available")
            return
        # For brevity, not using priceq/frequency logic here, but can be added
        if not self.items:
            try:
                quantity = self.quantity
                orderitem = OrderItem(self.symbol, self.securityId, quantity, self.orderType, self.productType, price, self.transaction_type, dhan=self.dhan)
                orderitem.mock = self.mock
                self.items.append(orderitem)
                self.items[-1].execute()
            except Exception as e:
                logging.error(f'Error creating order: {e}')
        else:
            last_price = self.items[-1].price
            order_price = round(last_price * (1-self.diff), 1) - 0.05
            try:
                orderitem = OrderItem(self.symbol, self.securityId, self.quantity, self.orderType, self.productType, order_price, self.transaction_type, dhan=self.dhan)
                orderitem.mock = self.mock
                self.items.append(orderitem)
                self.items[-1].execute()
            except Exception as e:
                logging.error(f'Error creating order: {e}')

    def confirmation(self):
        for item in self.items:
            try:
                item.confirm()
            except Exception as e:
                logging.error(f'Error confirming order: {e}')

    def clear(self):
        logging.info("clearing process")
        if self.items and self.items[-1].be_closed:
            self.items.pop()

    def run(self):
        try:
            self.entry()
        except Exception as e:
            logging.error(f'Error in entry: {e}')
            return False
        try:
            self.confirmation()
        except Exception as e:
            logging.error(f'Error in confirmation: {e}')
            return False
        try:
            self.clear()
        except Exception as e:
            logging.error(f'Error in clear: {e}')
            return False

# --- Order Thread/Manager ---
class ManagedTradeOrder(threading.Thread):
    def __init__(self, owner, symbol, args, dhan=None, mock=False):
        super().__init__()
        self.owner = owner
        self.symbol = symbol.upper()
        self.args = args.copy()  # dict: unit, diff, zag, type, etc.
        self.uuid = str(uuid.uuid4())
        self.status = 'initialized'
        self.running = threading.Event()
        self.running.set()
        self._lock = threading.Lock()
        self.created_at = datetime.datetime.now()
        self.log = logging.getLogger(f"Order-{self.owner}-{self.symbol}")
        self.dhan = dhan  # Pass a dhan API client if available
        self.bucket = None
        self.mock = mock

    def run(self):
        self.status = 'running'
        self.log.info(f"Started trade order thread for {self.owner} {self.symbol}")
        # Initialize bucket with current args
        self._init_bucket()
        try:
            while self.running.is_set():
                with self._lock:
                    # Update bucket args if changed
                    self._init_bucket(update_only=True)
                    self.bucket.run()
                time.sleep(2)  # Polling interval
        except Exception as e:
            self.status = 'error'
            self.log.error(f"Error: {e}")
        self.status = 'stopped'
        self.log.info(f"Stopped trade order thread for {self.owner} {self.symbol}")

    def stop(self):
        self.status = 'stopping'
        self.running.clear()

    def update_args(self, new_args):
        with self._lock:
            self.args.update(new_args)
            self.log.info(f"Updated args: {self.args}")

    def _init_bucket(self, update_only=False):
        # Create or update the Bucket instance with the latest args
        if self.bucket is None or not update_only:
            self.bucket = Bucket(
                symbol=self.symbol,
                productType=self.args.get('type', 'INTRADAY'),
                quantity=self.args.get('unit', 1),
                orderType=self.args.get('orderType', 'LIMIT'),
                price=self.args.get('price', 0),
                transaction_type=self.args.get('transaction_type', 'BUY'),
                diff=self.args.get('diff', 0.5),
                zag=self.args.get('zag', 1),
                dhan=self.dhan,
                mock=self.mock
            )
        else:
            # Update bucket parameters live
            self.bucket.quantity = self.args.get('unit', self.bucket.quantity)
            self.bucket.diff = max(self.args.get('diff', self.bucket.diff*100)/100, 0.001)
            self.bucket.zag = self.args.get('zag', self.bucket.zag)
            self.bucket.productType = self.args.get('type', self.bucket.productType)
            self.bucket.orderType = self.args.get('orderType', self.bucket.orderType)
            self.bucket.price = self.args.get('price', self.bucket.price)
            self.bucket.transaction_type = self.args.get('transaction_type', self.bucket.transaction_type)

    def get_status(self):
        return {
            'owner': self.owner,
            'symbol': self.symbol,
            'uuid': self.uuid,
            'status': self.status,
            'args': self.args,
            'created_at': self.created_at.isoformat(),
            'bucket_items': [str(item) for item in self.bucket.items] if self.bucket else [],
        }

# --- Order Registry ---
class TradeOrderManager:
    def __init__(self):
        self.registry = {}  # (owner, symbol, date): ManagedTradeOrder
        self._lock = threading.Lock()

    def order_key(self, owner, symbol):
        return (owner, symbol.upper(), get_today())

    def start_order(self, owner, symbol, args):
        key = self.order_key(owner, symbol)
        with self._lock:
            if key in self.registry and self.registry[key].is_alive():
                return False, 'already_running'
            order = ManagedTradeOrder(owner, symbol, args)
            self.registry[key] = order
            order.start()
            return True, order.uuid

    def stop_order(self, owner, symbol):
        key = self.order_key(owner, symbol)
        with self._lock:
            order = self.registry.get(key)
            if order and order.is_alive():
                order.stop()
                return True, 'stopped'
            return False, 'not_found'

    def update_order(self, owner, symbol, new_args):
        key = self.order_key(owner, symbol)
        with self._lock:
            order = self.registry.get(key)
            if order and order.is_alive():
                order.update_args(new_args)
                return True, 'updated'
            return False, 'not_found'

    def get_status(self, owner, symbol):
        key = self.order_key(owner, symbol)
        with self._lock:
            order = self.registry.get(key)
            if order:
                return order.get_status()
            return None

    def list_orders(self):
        with self._lock:
            return [order.get_status() for order in self.registry.values()]


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    logging.info("TradeOrderManager process started and waiting for requests...")
    manager = TradeOrderManager()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logging.info("TradeOrderManager process interrupted and exiting.")
