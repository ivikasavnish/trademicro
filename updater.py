import time,datetime
import redis
import datetime, time
from dhanhq import dhanhq
import pandas as pd
import logging
logging.basicConfig(level=logging.DEBUG)
# import options
users = pd.read_csv("users.csv").to_dict("records")
dhanclientarr = [dhanhq(user["clientid"],user["token"]) for user in users]
dhantokenarr = [user["token"] for user in users]
now = datetime.datetime.now()
import json
import uuid, os
import requests

def getdata(token):
    url = "https://api.dhan.co/orders"

    headers = {
        "access-token": token
    }

    response = requests.request("GET", url, headers=headers)
    return response.json()

# def getpositions():
#     url = "https://api.dhan.co/orders"

#     headers = {
#         "access-token": token
#     }

#     response = requests.request("GET", url, headers=headers)
#     data = response.json()

# def funds(token):
#     url = "https://api.dhan.co/orders"

#     headers = {
#         "access-token": token
#     }

#     response = requests.request("GET", url, headers=headers)
#     data = response.json()

# def holdings(token):
#     url = "https://api.dhan.co/orders"

#     headers = {
#         "access-token": token
#     }

#     response = requests.request("GET", url, headers=headers)
#     return response.json()

mode = "FULL_PULL"
redis_host = os.environ.get('REDIS_HOST', 'localhost')
redis_port = os.environ.get('REDIS_PORT', 6379)
cache = r = redis.Redis(host=redis_host, port=redis_port, db=0)
cache.set("mode",mode)
import logging
logging.basicConfig(level=logging.DEBUG)
def gettimestamp():
    now = datetime.datetime.now()
    min_index = now.hour * 60 + now.minute
    return min_index
def get_seconds_until_915():
    now = datetime.datetime.now()
    target_time = datetime.datetime(now.year, now.month, now.day, 9, 15)
    if now > target_time:
        # If current time is past 9:15, calculate time for next day's 9:15
        target_time += datetime.timedelta(days=1)
    remaining_seconds = int((target_time - now).total_seconds())
    return remaining_seconds
while gettimestamp() < 555:
    
    print("Waiting for market to open approx", get_seconds_until_915() , "seconds left")
    time.sleep(1)
while gettimestamp() < 930:
    time.sleep(0.5)
    logging.info("Mode changed to " +  r.get("mode").decode('utf-8'))
    for dhanclient in dhantokenarr:
        try:
            orders  = getdata(dhanclient)
            for order in orders:
                cache.set(order['orderId'], order['orderStatus'])
                print(order['orderId'], order['orderStatus'],int(time.time()))
            # logging.info(orders)
           
            # print(len(orders.get('data')))
            # for o in orders.get('data'):
            #     o['_id'] = o['orderId']
            #     o['status'] = o["orderStatus"]
            #     cache.set(o['_id'], o['status'])
            #     print(o['_id'], o['status'],int(time.time()))
            # options.run()
        except Exception as e:
            print(e)
