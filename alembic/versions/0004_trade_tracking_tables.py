"""
Revision ID: 0004_trade_tracking_tables
Revises: 0003_create_broker_tokens
Create Date: 2025-04-22 23:27:45
"""
revision = '0004_trade_tracking_tables'
down_revision = '0003_create_broker_tokens'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa
import enum

class TradeStateEnum(str, enum.Enum):
    PENDING = "PENDING"
    FILLED = "FILLED"
    REJECTED = "REJECTED"
    CANCELLED = "CANCELLED"

def upgrade():
    op.create_table(
        'trade_requests',
        sa.Column('id', sa.Integer, primary_key=True, autoincrement=True),
        sa.Column('user_id', sa.Integer, sa.ForeignKey('users.id'), nullable=False),
        sa.Column('symbol', sa.String(32), nullable=False),
        sa.Column('unit', sa.Integer, nullable=False),
        sa.Column('diff', sa.Float, nullable=False),
        sa.Column('zag', sa.Integer, nullable=False),
        sa.Column('type', sa.String(8), nullable=False),
        sa.Column('created_at', sa.DateTime, server_default=sa.func.now(), nullable=False),
    )
    op.create_table(
        'trade_fulfillments',
        sa.Column('id', sa.Integer, primary_key=True, autoincrement=True),
        sa.Column('request_id', sa.Integer, sa.ForeignKey('trade_requests.id'), nullable=False),
        sa.Column('fulfilled_at', sa.DateTime, server_default=sa.func.now(), nullable=False),
        sa.Column('status', sa.String(32), nullable=False),
        sa.Column('details', sa.String(256), nullable=True),
    )
    op.create_table(
        'trade_current_states',
        sa.Column('id', sa.Integer, primary_key=True, autoincrement=True),
        sa.Column('request_id', sa.Integer, sa.ForeignKey('trade_requests.id'), nullable=False, unique=True),
        sa.Column('state', sa.Enum(TradeStateEnum), nullable=False),
        sa.Column('updated_at', sa.DateTime, server_default=sa.func.now(), onupdate=sa.func.now(), nullable=False),
    )

def downgrade():
    op.drop_table('trade_current_states')
    op.drop_table('trade_fulfillments')
    op.drop_table('trade_requests')
