import json
import os
from fastapi import HTTPException
from pydantic import BaseModel
from typing import List

TRADES_FILE = os.path.join(os.path.dirname(__file__), "trade_sessions.json")

class TradeSessionStore:
    @staticmethod
    def load_sessions() -> list:
        if os.path.exists(TRADES_FILE):
            with open(TRADES_FILE, "r") as f:
                return json.load(f)
        return []

    @staticmethod
    def save_sessions(sessions: list):
        with open(TRADES_FILE, "w") as f:
            json.dump(sessions, f, indent=2)

    @staticmethod
    def add_session(session: dict) -> int:
        sessions = TradeSessionStore.load_sessions()
        session_id = len(sessions) + 1
        session["id"] = session_id
        sessions.append(session)
        TradeSessionStore.save_sessions(sessions)
        return session_id

    @staticmethod
    def get_session(session_id: int) -> dict:
        sessions = TradeSessionStore.load_sessions()
        for s in sessions:
            if s["id"] == session_id:
                return s
        raise HTTPException(status_code=404, detail="Session not found")
