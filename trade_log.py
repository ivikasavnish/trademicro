from dhanhq import dhanhq
import traceback
import time, datetime, json, os, sys, requests, redis, uuid, pandas as pd
from collections import deque, Counter
from parsearg import parse
# from reporting import register, ping
import websocket
import logging
import math
# from dhanfeed_sync import get_security_id, get_ltp_from_redis
logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(levelname)s - Line %(lineno)d: %(message)s')
def round_to_nearest_half(number):
    """Rounds a number to the nearest 0.5 increment with 1 decimal place if <= 1000.
    If number > 1000, returns the floor integer."""
    if number > 1000:
        return math.floor(number)
    else:
        return math.floor(number * 2) / 2

priceq = deque(maxlen=1000)

def find_price_position_in_deque(priceq, given_price):
    recent_priceq = deque(list(priceq)[-60:])
    if len(recent_priceq) < 30:
        return -1, 0
    prices_list = list(recent_priceq)
    price_frequencies = Counter(prices_list)
    sorted_frequencies = price_frequencies.most_common()
    for position, (price, frequency) in enumerate(sorted_frequencies, start=1):
        if price == given_price:
            return position, frequency
    return -1, 0 

class Object:
    def toJSON(self):
        return json.dumps(self, default=lambda o: o.__dict__, sort_keys=True, indent=4)

def send_ws(payload: dict):
    if ws is not None:
        me = Object()
        me.data = payload
        ws.send(me.toJSON())
    else:
        ws.connect("wss://invest.servloci.in/ws?room=trade", header={
            "X-Authhorization": "yb3TudagFA8ap6OT2ox.V4sNrMZMGOcAfB5r14HUyfMI1PbvMsqs"
        })


redis_host = os.environ.get('REDIS_HOST', 'localhost')
redis_port = os.environ.get('REDIS_PORT', 6379)
r = redis.Redis(host=redis_host, port=redis_port, db=0)

data = pd.read_csv("old.symbols.csv")
nse_data = data[data.SEM_EXM_EXCH_ID == "NSE"]
equity_data = nse_data[nse_data.SEM_SEGMENT == "E"]
equity_series = equity_data[equity_data.SEM_SERIES == "EQ"]
index_data = nse_data[nse_data.SEM_SEGMENT == "D"]
bankniftymonthly = index_data[(index_data.SEM_INSTRUMENT_NAME == 'OPTIDX') & (index_data.SEM_EXPIRY_FLAG == "M") & (index_data.SEM_TRADING_SYMBOL.str.contains("BANKNIFTY"))]
niftymonthly = index_data[(index_data.SEM_INSTRUMENT_NAME == 'OPTIDX') & (index_data.SEM_EXPIRY_FLAG == "M") & (index_data.SEM_TRADING_SYMBOL.str.startswith("NIFTY"))]

# Create a case-insensitive dictionary for symbol lookup
symbol_to_token = {row.SEM_TRADING_SYMBOL.upper(): str(row.SEM_SMST_SECURITY_ID) for _, row in equity_series.iterrows()}
# symbol_to_token.append(("SWIGGY", "27066"))
def get_ltp(symbol):
    instrument_token = symbol_to_token.get(symbol.upper())
    if instrument_token:
        ltp_data = r.get(f"ltp:{instrument_token}")
        if ltp_data:
            return float(ltp_data)
    return None

sdf = pd.read_csv("filtered.csv")
sdf.columns = ["a", "b", "token", "d", "e", "sym", "g", "h"]

def info(sym):
    return sdf[sdf.sym == sym + "-EQ"].to_dict(orient="records")[0]

def gettimestamp():
    now = datetime.datetime.now()
    return now.hour * 60 + now.minute

def get_seconds_until_915():
    now = datetime.datetime.now()
    target_time = datetime.datetime(now.year, now.month, now.day, 9, 15)
    if now > target_time:
        target_time += datetime.timedelta(days=1)
    return int((target_time - now).total_seconds())

class OrderItem:
    def __init__(self, symbol, security_id, quantity, order_type, product, price, transaction_type):
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

    def register(self):
        logging.info(f"Registering order: {self.__dict__}")

    def execute(self):
        logging.info(f"Executing order: {self}")
        try:
            time.sleep(1)
            self.action = dhan.place_order(security_id=self.security_id, exchange_segment="NSE_EQ",
                                           transaction_type=self.transaction_type, quantity=self.quantity,
                                           order_type=self.order_type, product_type=self.product, price=self.price)
            if self.action.get("status") == "success":
                logging.info("Order placed successfully")
                self.orderId = self.action.get("data").get("orderId")
                self.status = self.action.get("data").get("orderStatus")
                self.order_at = gettimestamp()
                r.rpush("orders", self.orderId)
            else:
                logging.error("Order not placed")
                logging.error(self.action, type(self.action))
                self.be_closed = True
        except Exception as e:
            logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
            logging.error(traceback.format_exc())

    def cancel(self):
        time.sleep(1)
        order = dhan.cancel_order(self.orderId)
        logging.info(order)
        if order.get("status") == "success":
            self.be_closed = True

    def direct_confirm(self, orderid):
        return

    def confirm(self):
        if self.status != "TRADED" and self.reverse is None:
            logging.info(f"{self.orderId}, {self.status}")
            oid = self.orderId
            if r.get(oid) is None:
                if self.close_at is None:
                    self.close_at = gettimestamp() + 300
                elif gettimestamp() > self.close_at:
                    self.be_closed = True
                return
            logging.info(f"Confirming order {self.orderId}, {r.get(self.orderId)}")
            if r.get(self.orderId) == b"TRADED":
                logging.info("Order confirmed")
                self.status = "TRADED"
                self.exit()
            if r.get(self.orderId) == b"CANCELLED":
                logging.info("Order cancelled")
                self.be_closed = True
        if self.status == "TRADED" and self.reverse is not None:
            logging.info(f"{self.reverse}, {self.reverse_order.status}")
            if r.get(self.reverse) == b"TRADED":
                logging.info("Order confirmed")
                self.reverse_order.status = "TRADED"
                self.be_closed = True
                priceq.append(round_to_nearest_half(self.price))
                logging.info(f"Price queue: {priceq}")
            if r.get(self.reverse) == b"CANCELLED":
                logging.info("Order cancelled")
                self.be_closed = True
        if self.status == "TRANSIT" and self.reverse is None:
            if self.order_at + 1 <= gettimestamp():
                logging.info("Order not confirmed, cancelling order")
                self.cancel()

    def exit(self):
        exit_price = round(self.price * 1.007, 1)
        self.reverse_order = OrderItem(symbol=self.symbol, security_id=self.security_id, quantity=self.quantity,
                                       order_type="LIMIT", product=self.product, price=exit_price,
                                       transaction_type="SELL")
        try:
            self.reverse_order.execute()
            self.reverse = self.reverse_order.orderId
        except Exception as e:
            logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
            logging.error(traceback.format_exc())

    def __str__(self):
        return str(self.__dict__)

    def __repr__(self):
        return str(self.__dict__)

class Bucket:
    broker = None
    user = None

    def __init__(self, symbol, productType, quantity=1, orderType="LIMIT", price=0, transaction_type="BUY", diff=0.5, zag=1):
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

    def entry(self):
        price = get_ltp(self.symbol)
        if price is None:
            logging.info("Price not available")
            return
        pos, freq = find_price_position_in_deque(priceq, round_to_nearest_half(price))
        logging.info(f"Price position: {pos}, Price frequency: {freq} depth: {len(self.items)}")
        if not self.items:
            try:
                # quantity = max(self.quantity, self.zag -pos + 1)
                if pos > 1:
                    quantity = max(self.quantity, self.zag - pos + 1)
                else:
                    quantity = self.quantity
                orderitem = OrderItem(self.symbol, self.securityId, quantity, self.orderType, self.productType, price, self.transaction_type)
                self.items.append(orderitem)
                self.items[-1].execute()
            except Exception as e:
                logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
                logging.error(traceback.format_exc())
        else:
            last_price = self.items[-1].price
            diff = self.diff + (math.floor(math.sqrt(len(self.items))) * 0.0001)
            order_price = round(last_price * (1-diff), 1) - 0.05
            intimation_price = last_price * (1-diff+0.0005)
            logging.info(f"Last price: {last_price}, Order price: {order_price}, Intimation price: {intimation_price}, diff: {diff}, effective diff: {order_price/last_price}")
            if price < intimation_price:
                try:
                    quantity = len(self.items) % self.zag  + self.quantity
                    if pos > 1:
                        quantity = max(self.quantity, self.zag-pos+1)
                    orderitem = OrderItem(self.symbol, self.securityId, quantity, self.orderType, self.productType, order_price, self.transaction_type)
                    self.items.append(orderitem)
                    self.items[-1].execute()
                except Exception as e:
                    logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
                    logging.error(traceback.format_exc())
            logging.info("Already in position")

    def exit(self):
        pass

    def clear(self):
        logging.info("clearing process")
        if self.items and self.items[-1].be_closed:
            self.items.pop()

    def confirmation(self):
        for item in self.items:
            try:
                item.confirm()
            except Exception as e:
                logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
                logging.error(traceback.format_exc())

    def run(self):
        try:
            self.entry()
        except Exception as e:
            logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
            return False
        try:
            self.confirmation()
        except Exception as e:
            logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
            return False
        try:
            self.clear()
        except Exception as e:
            logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
            return False

if __name__ == "__main__":
    args = parse()
    user = args.user
    productType = args.type
    userdata = None
    users = pd.read_csv("users.csv").to_dict("records")
    for u in users:
        if u.get("name") == user:
            userdata = u
            break
    if userdata is not None:
        logging.info(f"User found: {userdata}")
        dhan = dhanhq(userdata.get("user"), userdata.get("token"))

        logging.info(dhan.get_fund_limits())
        bucket = Bucket(symbol=args.symbol.upper(), quantity=args.unit, diff=args.diff, productType=productType, zag=args.zag)
        logging.info(bucket.__dict__)
        bucket.uuid = str(uuid.uuid4())
        bucket.broker = "dhan"
        logging.info(dhan.__dict__)
        bucket.user = userdata.get("user")

        # register(bucket.__dict__)

        while gettimestamp() < 555:
            logging.info(f"Waiting for market to open approx {get_seconds_until_915()} seconds left")
            time.sleep(1)
        while gettimestamp() < 930:
            bucket.run()
            # ping_status = ping(bucket.uuid)
            # if ping_status.get("active") == True:
            #     bucket.run()
            # elif ping_status.get("active") is None:
            #     break
            # else:
            #     logging.info("Not active")
            try:
                pass
                # send_ws(bucket.__dict__)
            except Exception as e:
                logging.error(f'Error on line {sys.exc_info()[-1].tb_lineno}, {type(e).__name__}: {e}')
                logging.error(str(e))
                continue
