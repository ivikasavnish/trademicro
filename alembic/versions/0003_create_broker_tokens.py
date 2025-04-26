"""
Revision ID: 0003_create_broker_tokens
Revises: 0002_create_users
Create Date: 2025-04-22 23:20:00
"""
revision = '0003_create_broker_tokens'
down_revision = '0002_create_users'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa

def upgrade():
    op.create_table(
        'broker_tokens',
        sa.Column('id', sa.Integer, primary_key=True, autoincrement=True),
        sa.Column('user_id', sa.Integer, sa.ForeignKey('users.id'), nullable=False),
        sa.Column('client_id', sa.String(64), nullable=False),
        sa.Column('broker_token', sa.String(256), nullable=False),
        sa.Column('expires_at', sa.DateTime, nullable=False),
        sa.Column('created_at', sa.DateTime, server_default=sa.func.now(), nullable=False),
    )

def downgrade():
    op.drop_table('broker_tokens')
