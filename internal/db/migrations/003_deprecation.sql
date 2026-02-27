-- Migration: Add package_deprecation table
-- Version: 3
-- Description: Support package deprecation (like npm deprecate)

-- Track deprecated packages/versions
CREATE TABLE IF NOT EXISTS package_deprecations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    package_name VARCHAR(255) NOT NULL,
    version VARCHAR(50),  -- NULL means entire package is deprecated
    reason TEXT,  -- Deprecation message
    deprecated_by VARCHAR(100),  -- username
    deprecated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(package_name, version)
);

CREATE INDEX IF NOT EXISTS idx_deprecations_package ON package_deprecations(package_name);

INSERT INTO schema_version (version, description) VALUES (3, 'Add package deprecation support');
