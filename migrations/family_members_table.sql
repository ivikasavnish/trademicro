-- SQL DDL for creating the family_members table

CREATE TABLE IF NOT EXISTS family_members (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    pin TEXT,
    portfolio_id BIGINT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    client_token VARCHAR(255),
    client_id VARCHAR(255),
    user_id BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create an index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_family_members_user_id ON family_members(user_id);

-- Add comments to the table and columns for documentation
COMMENT ON TABLE family_members IS 'Stores information about family members connected to main user accounts';
COMMENT ON COLUMN family_members.id IS 'Primary key';
COMMENT ON COLUMN family_members.name IS 'Name of the family member';
COMMENT ON COLUMN family_members.email IS 'Email address of the family member';
COMMENT ON COLUMN family_members.phone IS 'Phone number of the family member';
COMMENT ON COLUMN family_members.pin IS 'PIN/password for the family member';
COMMENT ON COLUMN family_members.portfolio_id IS 'Portfolio ID for the family member';
COMMENT ON COLUMN family_members.is_active IS 'Whether the family member account is active';
COMMENT ON COLUMN family_members.created_at IS 'Timestamp when the record was created';
COMMENT ON COLUMN family_members.updated_at IS 'Timestamp when the record was last updated';
COMMENT ON COLUMN family_members.client_token IS 'Broker client token for the family member';
COMMENT ON COLUMN family_members.client_id IS 'Broker client ID for the family member';
COMMENT ON COLUMN family_members.user_id IS 'Foreign key referencing the main user account';
