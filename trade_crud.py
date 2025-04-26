from sqlalchemy.future import select
from sqlalchemy.exc import NoResultFound
from db import AsyncSessionLocal
from models import TradeOrder

# CRUD for trade orders
async def create_trade_order(data):
    async with AsyncSessionLocal() as session:
        order = TradeOrder(**data)
        session.add(order)
        await session.commit()
        await session.refresh(order)
        return order

async def get_trade_order(order_id):
    async with AsyncSessionLocal() as session:
        result = await session.execute(select(TradeOrder).where(TradeOrder.id == order_id))
        order = result.scalar_one_or_none()
        if not order:
            raise NoResultFound()
        return order

async def list_trade_orders():
    async with AsyncSessionLocal() as session:
        result = await session.execute(select(TradeOrder))
        return result.scalars().all()

async def update_trade_order(order_id, update_data):
    async with AsyncSessionLocal() as session:
        result = await session.execute(select(TradeOrder).where(TradeOrder.id == order_id))
        order = result.scalar_one_or_none()
        if not order:
            raise NoResultFound()
        for key, value in update_data.items():
            setattr(order, key, value)
        await session.commit()
        await session.refresh(order)
        return order
