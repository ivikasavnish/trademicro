"""
Revision ID: 0001_create_trade_orders
Revises: 
Create Date: 2025-04-22 23:10:00
"""
revision = '0001_create_trade_orders'
down_revision = None
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa

def upgrade():
    op.create_table(
        'trade_orders',
        sa.Column('id', sa.Integer, primary_key=True, autoincrement=True),
        sa.Column('symbol', sa.String(32), nullable=False),
        sa.Column('unit', sa.Integer, nullable=False),
        sa.Column('diff', sa.Float, nullable=False),
        sa.Column('zag', sa.Integer, nullable=False),
        sa.Column('type', sa.String(16), nullable=False),
        sa.Column('user', sa.String(32), nullable=False),
        sa.Column('status', sa.String(16), nullable=False, server_default='running'),
    )

def downgrade():
    op.drop_table('trade_orders')
