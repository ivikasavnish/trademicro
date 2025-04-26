"""
Revision ID: 0002_create_users
Revises: 0001_create_trade_orders
Create Date: 2025-04-22 23:18:00
"""
revision = '0002_create_users'
down_revision = '0001_create_trade_orders'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa

def upgrade():
    op.create_table(
        'users',
        sa.Column('id', sa.Integer, primary_key=True, autoincrement=True),
        sa.Column('username', sa.String(32), unique=True, nullable=False),
        sa.Column('full_name', sa.String(64)),
        sa.Column('hashed_password', sa.String(128), nullable=False),
        sa.Column('disabled', sa.Boolean, nullable=False, server_default='false'),
    )

def downgrade():
    op.drop_table('users')
