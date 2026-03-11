-- 002_create_users.down.sql

-- Remove foreign keys from sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS project_id;
ALTER TABLE sessions DROP COLUMN IF EXISTS user_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
