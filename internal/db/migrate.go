package db

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// RunMigrations runs all pending migrations
func RunMigrations(db *gorm.DB) error {
	// Ensure schema_version table exists
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			description VARCHAR(255)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Get current version
	var currentVersion int
	row := db.Raw("SELECT COALESCE(MAX(version), 0) FROM schema_version").Row()
	row.Scan(&currentVersion)

	// Load and sort migrations
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Run pending migrations
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		// Run migration in transaction
		tx := db.Begin()
		if err := tx.Exec(m.SQL).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}

		// Record migration
		if err := tx.Exec(
			"INSERT INTO schema_version (version, description) VALUES (?, ?)",
			m.Version, m.Description,
		).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}

		tx.Commit()
	}

	return nil
}

// loadMigrations loads all migration files
func loadMigrations() ([]Migration, error) {
	var migrations []Migration

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Parse version from filename (e.g., 002_gc_metadata.sql -> 2)
		name := entry.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) != 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		// Read migration content
		content, err := migrationsFS.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			return nil, err
		}

		// Extract description from filename
		description := strings.TrimSuffix(parts[1], ".sql")
		description = strings.ReplaceAll(description, "_", " ")

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetSchemaVersion returns the current schema version
func GetSchemaVersion(db *gorm.DB) (int, error) {
	var version int
	row := db.Raw("SELECT COALESCE(MAX(version), 0) FROM schema_version").Row()
	err := row.Scan(&version)
	return version, err
}
