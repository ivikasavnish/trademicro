from sqlalchemy import Column, Integer, String, Float, DateTime, ForeignKey, Enum
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func
import enum

Base = declarative_base()

class TradeRequest(Base):
    __tablename__ = "trade_requests"
    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    symbol = Column(String(32), nullable=False)
    unit = Column(Integer, nullable=False)
    diff = Column(Float, nullable=False)
    zag = Column(Integer, nullable=False)
    type = Column(String(8), nullable=False)
    created_at = Column(DateTime, server_default=func.now(), nullable=False)

class TradeFulfillment(Base):
    __tablename__ = "trade_fulfillments"
    id = Column(Integer, primary_key=True, autoincrement=True)
    request_id = Column(Integer, ForeignKey("trade_requests.id"), nullable=False)
    fulfilled_at = Column(DateTime, server_default=func.now(), nullable=False)
    status = Column(String(32), nullable=False)  # e.g. 'FILLED', 'REJECTED', etc.
    details = Column(String(256), nullable=True)

class TradeStateEnum(enum.Enum):
    PENDING = "PENDING"
    FILLED = "FILLED"
    REJECTED = "REJECTED"
    CANCELLED = "CANCELLED"

class TradeCurrentState(Base):
    __tablename__ = "trade_current_states"
    id = Column(Integer, primary_key=True, autoincrement=True)
    request_id = Column(Integer, ForeignKey("trade_requests.id"), nullable=False, unique=True)
    state = Column(Enum(TradeStateEnum), nullable=False)
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now(), nullable=False)
