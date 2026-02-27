package db

import (
	"time"
)

// PackageGCMetadata tracks package access for garbage collection
type PackageGCMetadata struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	PackageName    string     `gorm:"uniqueIndex;size:255;not null" json:"packageName"`
	Version        string     `gorm:"size:50" json:"version"`                      // Last accessed version
	LastAccessedAt *time.Time `json:"lastAccessedAt"`                              // Last access time
	AccessCount    int        `gorm:"default:0" json:"accessCount"`                // Total access count
	MarkedForGC    bool       `gorm:"default:false" json:"markedForGc"`            // Flagged for cleanup
	MarkedAt       *time.Time `json:"markedAt,omitempty"`                          // When marked
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (PackageGCMetadata) TableName() string {
	return "package_gc_metadata"
}

// OrphanedFile tracks files not referenced by any metadata
type OrphanedFile struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	FilePath     string     `gorm:"size:500;not null" json:"filePath"`
	FileSize     int64      `json:"fileSize"`
	DiscoveredAt time.Time  `json:"discoveredAt"`
	CleanedAt    *time.Time `json:"cleanedAt,omitempty"`
}

func (OrphanedFile) TableName() string {
	return "orphaned_files"
}

// PackageDeprecation tracks deprecated packages/versions
type PackageDeprecation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PackageName   string    `gorm:"uniqueIndex:idx_pkg_ver;size:255;not null" json:"packageName"`
	Version       string    `gorm:"uniqueIndex:idx_pkg_ver;size:50" json:"version"` // NULL = entire package
	Reason        string    `gorm:"type:text" json:"reason"`                         // Deprecation message
	DeprecatedBy  string    `gorm:"size:100" json:"deprecatedBy"`                    // username
	DeprecatedAt  time.Time `json:"deprecatedAt"`
}

func (PackageDeprecation) TableName() string {
	return "package_deprecations"
}

// GCPolicy defines garbage collection policy
type GCPolicy struct {
	// Age threshold: packages not accessed for this long are candidates
	MaxInactiveDays int `json:"maxInactiveDays"` // Default: 180 days (6 months)

	// Keep minimum versions per package
	MinVersionsToKeep int `json:"minVersionsToKeep"` // Default: 5

	// Dry run mode: only report, don't delete
	DryRun bool `json:"dryRun"`

	// Include deprecated packages
	IncludeDeprecated bool `json:"includeDeprecated"`

	// Minimum package age before GC eligibility
	MinPackageAgeDays int `json:"minPackageAgeDays"` // Default: 30 days
}

// DefaultGCPolicy returns the default GC policy
func DefaultGCPolicy() GCPolicy {
	return GCPolicy{
		MaxInactiveDays:   180,
		MinVersionsToKeep: 5,
		DryRun:            true,
		IncludeDeprecated: false,
		MinPackageAgeDays: 30,
	}
}
