-- Migration: Add package_gc_metadata table
-- Version: 2
-- Description: Add metadata for garbage collection

-- Track package download/access for GC decisions
CREATE TABLE IF NOT EXISTS package_gc_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    package_name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(50),
    last_accessed_at DATETIME,  -- Last time package/version was accessed
    access_count INTEGER DEFAULT 0,
    marked_for_gc BOOLEAN DEFAULT 0,  -- Flagged for cleanup
    marked_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gc_package ON package_gc_metadata(package_name);
CREATE INDEX IF NOT EXISTS idx_gc_last_accessed ON package_gc_metadata(last_accessed_at);
CREATE INDEX IF NOT EXISTS idx_gc_marked ON package_gc_metadata(marked_for_gc);

-- Track orphaned files (files not referenced by any metadata)
CREATE TABLE IF NOT EXISTS orphaned_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER,
    discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    cleaned_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_orphaned_cleaned ON orphaned_files(cleaned_at);

INSERT INTO schema_version (version, description) VALUES (2, 'Add GC metadata tables');
