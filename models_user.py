from sqlalchemy import Column, Integer, String, Boolean
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class User(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True, autoincrement=True)
    username = Column(String(32), unique=True, nullable=False)
    full_name = Column(String(64), nullable=True)
    hashed_password = Column(String(128), nullable=False)
    disabled = Column(Boolean, default=False, nullable=False)
