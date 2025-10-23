-- Script to make the first user an admin
-- Run this against your production database

UPDATE users 
SET is_admin = true 
WHERE id = (
    SELECT id 
    FROM users 
    ORDER BY created_at ASC 
    LIMIT 1
);

-- Verify the change
SELECT id, email, first_name, last_name, is_admin, created_at 
FROM users 
ORDER BY created_at ASC 
LIMIT 1;