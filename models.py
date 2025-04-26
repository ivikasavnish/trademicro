from sqlalchemy import Column, Integer, String, Float, Text
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class TradeOrder(Base):
    __tablename__ = "trade_orders"
    id = Column(Integer, primary_key=True, autoincrement=True)
    symbol = Column(String(32), nullable=False)
    unit = Column(Integer, nullable=False)
    diff = Column(Float, nullable=False)
    zag = Column(Integer, nullable=False)
    type = Column(String(16), nullable=False)
    user = Column(String(32), nullable=False)
    status = Column(String(16), default="running", nullable=False)
    # Add more fields as needed (timestamps, etc)
