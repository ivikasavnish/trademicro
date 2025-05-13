CREATE TABLE IF NOT EXISTS family_members (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    pin VARCHAR(10),
    portfolio_id INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create foreign key constraint to users table
ALTER TABLE family_members ADD CONSTRAINT fk_family_members_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create index
CREATE INDEX idx_family_members_user_id ON family_members(user_id);