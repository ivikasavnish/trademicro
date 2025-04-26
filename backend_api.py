# FastAPI backend scaffold for trading management
# Requirements: fastapi, uvicorn, sqlalchemy, asyncpg, aioredis, python-jose, passlib

from fastapi import FastAPI, Depends, HTTPException, status, WebSocket, WebSocketDisconnect
from typing import List, Dict, Optional, Set
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from pydantic import BaseModel
from typing import Optional, List
from jose import JWTError, jwt
from passlib.context import CryptContext
import redis.asyncio as aioredis  # Modern async Redis client for Python 3.12+
import asyncpg
import os
from datetime import datetime, timedelta

from fastapi.middleware.cors import CORSMiddleware

from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from db import get_db
from models_broker import BrokerToken

app = FastAPI()

# --- WebSocket Connection Manager ---
class ConnectionManager:
    def __init__(self):
        self.active_connections: Set[WebSocket] = set()

    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.add(websocket)

    def disconnect(self, websocket: WebSocket):
        self.active_connections.discard(websocket)

    async def broadcast(self, message: dict):
        to_remove = set()
        for connection in self.active_connections:
            try:
                await connection.send_json(message)
            except Exception:
                to_remove.add(connection)
        for conn in to_remove:
            self.active_connections.discard(conn)

manager = ConnectionManager()

# Allow CORS for local development (React/Vite)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173", "http://127.0.0.1:5173"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# --- Config ---
SECRET_KEY = "your_secret_key_here"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60

# --- Auth ---
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/token")

# --- Database (PostgreSQL) ---
DATABASE_URL = os.environ.get("DATABASE_URL", "postgresql://postgres:password@localhost/trademicro")
REDIS_URL = os.environ.get("REDIS_URL", "redis://localhost:6379/0")

# --- Models ---
class Token(BaseModel):
    access_token: str
    token_type: str

class TokenData(BaseModel):
    username: Optional[str] = None

class User(BaseModel):
    id: Optional[int] = None
    username: str
    full_name: Optional[str] = None
    disabled: Optional[bool] = None

class UserInDB(User):
    hashed_password: str

class TradeSession(BaseModel):
    id: Optional[int] = None
    symbol: str
    unit: int
    diff: float
    zag: int
    type: str
    user: str
    status: Optional[str] = None

# --- Dummy User Store for MVP ---
fake_users_db = {
    "sonam": {
        "username": "sonam",
        "full_name": "Sonam Trader",
        "hashed_password": pwd_context.hash("testpass"),
        "disabled": False,
    },
}

def verify_password(plain_password, hashed_password):
    return pwd_context.verify(plain_password, hashed_password)

# --- Database user lookup for endpoints that require user.id ---
from sqlalchemy.ext.asyncio import AsyncSession
from models_user import User as UserModel

async def get_user_from_db(db: AsyncSession, username: str) -> Optional[User]:
    result = await db.execute(select(UserModel).where(UserModel.username == username))
    user_row = result.scalar_one_or_none()
    if user_row:
        return User(
            id=user_row.id,
            username=user_row.username,
            full_name=getattr(user_row, 'full_name', None),
            disabled=getattr(user_row, 'disabled', None)
        )
    return None

def get_user(db, username: str):
    if username in db:
        user_dict = db[username]
        return UserInDB(**user_dict)

def authenticate_user(db, username: str, password: str):
    user = get_user(db, username)
    if not user:
        return False
    if not verify_password(password, user.hashed_password):
        return False
    return user

def create_access_token(data: dict):
    from datetime import datetime, timedelta
    to_encode = data.copy()
    expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

from fastapi import Request

async def get_current_user(token: str = Depends(oauth2_scheme), db: AsyncSession = Depends(get_db)):
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise credentials_exception
        token_data = TokenData(username=username)
    except JWTError:
        raise credentials_exception
    # Try DB user first
    db_user = await get_user_from_db(db, token_data.username)
    if db_user:
        return db_user
    # Fallback to fake_users_db for legacy endpoints
    user = get_user(fake_users_db, username=token_data.username)
    if user is None:
        raise credentials_exception
    # Patch in a dummy id for legacy users
    user.id = 1
    return user

# --- Auth Endpoints ---
@app.post("/token", response_model=Token)
async def login_for_access_token(form_data: OAuth2PasswordRequestForm = Depends()):
    user = authenticate_user(fake_users_db, form_data.username, form_data.password)
    if not user:
        raise HTTPException(status_code=400, detail="Incorrect username or password")
    access_token = create_access_token(data={"sub": user.username})
    return {"access_token": access_token, "token_type": "bearer"}

@app.get("/users/me", response_model=User)
async def read_users_me(current_user: User = Depends(get_current_user)):
    return current_user

# --- Trading Endpoints ---
from trade_store import TradeSessionStore

@app.websocket("/ws/trades")
async def websocket_endpoint(websocket: WebSocket):
    await manager.connect(websocket)
    try:
        while True:
            # We don't expect messages from client, just keep alive
            await websocket.receive_text()
    except WebSocketDisconnect:
        manager.disconnect(websocket)

@app.post("/trade/start", response_model=TradeSession)
async def start_trade(session: TradeSession, current_user: User = Depends(get_current_user)):
    session.status = "running"
    # Only use provided fields for input (ignore id/status if not sent)
    session_dict = session.dict(exclude_unset=True)
    session_id = TradeSessionStore.add_session(session_dict)
    session_dict["id"] = session_id
    session_dict["status"] = "running"
    # Broadcast new trade to all WebSocket clients
    await manager.broadcast({"event": "trade_started", "trade": session_dict})
    return session_dict

@app.post("/trade/stop/{session_id}")
async def stop_trade(session_id: int, current_user: User = Depends(get_current_user)):
    # TODO: Stop trading background task
    return {"status": "stopped", "session_id": session_id}

@app.get("/trade/status/{session_id}", response_model=TradeSession)
async def trade_status(session_id: int, current_user: User = Depends(get_current_user)):
    session = TradeSessionStore.get_session(session_id)
    return session

@app.get("/symbols")
async def get_symbols(current_user: User = Depends(get_current_user)):
    # TODO: Load from CSV or DB
    return ["HINDZINC", "RELIANCE", "SBIN"]

# --- List Users Endpoint ---
from get_users import get_all_users
from trade_store import TradeSessionStore

@app.get("/users")
async def list_users(current_user: User = Depends(get_current_user)):
    return get_all_users()

# --- List All Trade Orders ---
@app.get("/trade/orders")
async def list_trades(current_user: User = Depends(get_current_user)):
    return TradeSessionStore.load_sessions()

# --- Broker Token Endpoints ---
from pydantic import BaseModel

class BrokerTokenCreate(BaseModel):
    client_id: str
    broker_token: str
    expires_at: datetime

@app.post("/broker-token")
async def create_broker_token(data: BrokerTokenCreate, current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)):
    token = BrokerToken(
        user_id=current_user.id,
        client_id=data.client_id,
        broker_token=data.broker_token,
        expires_at=data.expires_at,
    )
    db.add(token)
    await db.commit()
    await db.refresh(token)
    return {"id": token.id}

@app.get("/broker-token")
async def list_broker_tokens(current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)):
    q = await db.execute(select(BrokerToken).where(BrokerToken.user_id == current_user.id))
    tokens = q.scalars().all()
    return [
        {
            "id": t.id,
            "client_id": t.client_id,
            "broker_token": t.broker_token,
            "expires_at": t.expires_at,
            "created_at": t.created_at,
        } for t in tokens
    ]

@app.get("/broker-token/{token_id}")
async def get_broker_token(token_id: int, current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)):
    q = await db.execute(select(BrokerToken).where(BrokerToken.id == token_id, BrokerToken.user_id == current_user.id))
    token = q.scalar_one_or_none()
    if not token:
        raise HTTPException(status_code=404, detail="Not found")
    return {
        "id": token.id,
        "client_id": token.client_id,
        "broker_token": token.broker_token,
        "expires_at": token.expires_at,
        "created_at": token.created_at,
    }

# --- Health Check ---
@app.get("/")
async def root():
    return {"status": "ok"}
