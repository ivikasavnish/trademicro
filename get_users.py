import pandas as pd
import os

USERS_CSV = os.path.join(os.path.dirname(__file__), "users.csv")

def get_all_users():
    if os.path.exists(USERS_CSV):
        users = pd.read_csv(USERS_CSV).to_dict("records")
        return [u.get("name") or u.get("username") for u in users]
    return []
