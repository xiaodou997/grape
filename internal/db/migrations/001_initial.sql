-- Migration: Initial schema
-- Version: 1
-- Description: Create initial tables (users, packages, tokens, etc.)

-- Note: This migration is typically handled by GORM AutoMigrate
-- This file is provided for reference and manual database setup

-- Tables are created automatically by GORM:
-- - users
-- - packages  
-- - package_versions
-- - package_owners
-- - tokens
-- - audit_logs
-- - webhooks

INSERT INTO schema_version (version, description) VALUES (1, 'Initial schema (GORM AutoMigrate)');
