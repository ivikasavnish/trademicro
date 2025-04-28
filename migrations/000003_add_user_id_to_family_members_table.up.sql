-- Add user_id column to family_members and set up foreign key constraint
ALTER TABLE family_members
ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;

-- Optionally, set a default user_id for existing rows or handle nulls as needed
-- UPDATE family_members SET user_id = 1 WHERE user_id IS NULL; -- if you want to set a default
