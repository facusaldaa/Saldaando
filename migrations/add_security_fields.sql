-- Migration: Add security fields to lobbies and users tables
-- Run this if you have an existing database

-- Add language field to users table (if it doesn't exist)
-- Note: SQLite doesn't support IF NOT EXISTS for ALTER TABLE
-- You may need to recreate the table or handle this in application code

-- Add invite_token and group_chat_id to lobbies table
-- Note: For SQLite, you may need to recreate the table or use a migration tool
-- The application will handle this automatically for new installations

-- For existing installations, you can:
-- 1. Backup your data
-- 2. Drop and recreate tables (data will be lost)
-- 3. Or use a migration tool like golang-migrate

-- The application will generate tokens automatically when creating new lobbies
-- Existing lobbies without tokens will need to be recreated

