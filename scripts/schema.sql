-- Grape Database Schema
-- Version: 1.0.0
-- Database: SQLite / PostgreSQL / MySQL compatible

-- ============================================
-- Users Table
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255),
    password VARCHAR(255) NOT NULL,  -- bcrypt hashed
    role VARCHAR(50) DEFAULT 'developer',  -- admin, developer, readonly
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Default admin user (password: admin)
-- INSERT INTO users (username, email, password, role) 
-- VALUES ('admin', 'admin@grape.local', '$2a$10$...', 'admin');

-- ============================================
-- Packages Table (for quick queries)
-- ============================================
CREATE TABLE IF NOT EXISTS packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    description VARCHAR(500),
    private BOOLEAN DEFAULT 0,
    latest VARCHAR(50),  -- latest version
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_packages_name ON packages(name);

-- ============================================
-- Package Versions Table
-- ============================================
CREATE TABLE IF NOT EXISTS package_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    package_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    tarball VARCHAR(500),  -- tarball filename
    shasum VARCHAR(100),
    publisher VARCHAR(100),  -- username of publisher
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(package_name, version)
);

CREATE INDEX IF NOT EXISTS idx_package_versions_package ON package_versions(package_name);

-- ============================================
-- Package Owners Table (Package-level ACL)
-- ============================================
CREATE TABLE IF NOT EXISTS package_owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    package_name VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(package_name, username)
);

CREATE INDEX IF NOT EXISTS idx_package_owners_package ON package_owners(package_name);
CREATE INDEX IF NOT EXISTS idx_package_owners_user ON package_owners(username);

-- ============================================
-- Tokens Table (CI/CD persistent tokens)
-- ============================================
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 hash
    readonly BOOLEAN DEFAULT 0,
    expires_at DATETIME,  -- NULL = never expires
    last_used DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tokens_hash ON tokens(token_hash);

-- ============================================
-- Audit Logs Table
-- ============================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action VARCHAR(50) NOT NULL,  -- login, logout, publish, unpublish, user_create, user_delete, config_update, backup_create, backup_restore
    username VARCHAR(100),
    ip VARCHAR(50),
    detail VARCHAR(500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_username ON audit_logs(username);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- ============================================
-- Webhooks Table
-- ============================================
CREATE TABLE IF NOT EXISTS webhooks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255),  -- HMAC secret
    events VARCHAR(500),  -- comma-separated event types
    enabled BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_delivery_at DATETIME
);

-- ============================================
-- Schema Version Table (for migrations)
-- ============================================
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    description VARCHAR(255)
);

-- Initialize schema version
INSERT INTO schema_version (version, description) 
VALUES (1, 'Initial schema');

-- ============================================
-- Views for common queries
-- ============================================

-- Package statistics view
CREATE VIEW IF NOT EXISTS v_package_stats AS
SELECT 
    p.name,
    p.description,
    p.private,
    p.latest,
    p.created_at,
    p.updated_at,
    COUNT(DISTINCT pv.version) as version_count,
    GROUP_CONCAT(DISTINCT po.username) as owners
FROM packages p
LEFT JOIN package_versions pv ON p.name = pv.package_name
LEFT JOIN package_owners po ON p.name = po.package_name
GROUP BY p.id;

-- User activity view
CREATE VIEW IF NOT EXISTS v_user_activity AS
SELECT 
    u.username,
    u.email,
    u.role,
    u.created_at,
    u.last_login,
    COUNT(DISTINCT t.id) as token_count,
    COUNT(DISTINCT pv.id) as packages_published
FROM users u
LEFT JOIN tokens t ON u.id = t.user_id
LEFT JOIN package_versions pv ON u.username = pv.publisher
GROUP BY u.id;
