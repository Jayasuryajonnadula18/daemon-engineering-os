package analysis

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// AnalysisCache tracks per-file analysis state in SQLite.
// SQLite remains canonical. This is DERIVED analysis metadata.
type AnalysisCache struct {
	mu sync.RWMutex
	db *sql.DB
}

type CacheEntry struct {
	FilePath          string    `json:"file_path"`
	FileHash          string    `json:"file_hash"`          // SHA-256 of file content
	ASTHash           string    `json:"ast_hash"`           // Hash of AST tree properties
	AnalyzerVersion   string    `json:"analyzer_version"`   // semver string e.g. "1.0.0"
	DependencyState   string    `json:"dependency_state"`   // Hash or version representing project dependencies
	TwinVersion       string    `json:"twin_version"`       // Version of the engineering twin model
	FindingCount      int       `json:"finding_count"`
	FindingsGenerated string    `json:"findings_generated"` // JSON string representation of findings list
	AnalyzerStatus    string    `json:"analyzer_status"`    // JSON string representation of analyzer status
	AnalyzedAt        time.Time `json:"analyzed_at"`
	Stale             bool      `json:"stale"`
}

type CacheStats struct {
	TotalEntries int   `json:"total_entries"`
	StaleEntries int   `json:"stale_entries"`
	SizeBytes    int64 `json:"size_bytes"`
}

func NewAnalysisCache(dbPath string) (*AnalysisCache, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open analysis_cache db: %w", err)
	}

	// Schema that tracks all required properties.
	schema := `
	CREATE TABLE IF NOT EXISTS analysis_cache (
		file_path          TEXT PRIMARY KEY,
		file_hash          TEXT NOT NULL,
		ast_hash           TEXT NOT NULL DEFAULT '',
		analyzer_version   TEXT NOT NULL,
		dependency_state   TEXT NOT NULL DEFAULT '',
		twin_version       TEXT NOT NULL DEFAULT '',
		finding_count      INTEGER NOT NULL DEFAULT 0,
		findings_generated TEXT NOT NULL DEFAULT '',
		analyzer_status    TEXT NOT NULL DEFAULT '',
		analyzed_at        DATETIME NOT NULL,
		stale              INTEGER NOT NULL DEFAULT 0
	);`

	// Verify columns exist or recreate table to preserve schema integrity.
	_, checkErr := db.Exec("SELECT ast_hash, dependency_state, twin_version, findings_generated, analyzer_status FROM analysis_cache LIMIT 1")
	if checkErr != nil {
		// Drop and recreate derived metadata cache if outdated schema.
		_, _ = db.Exec("DROP TABLE IF EXISTS analysis_cache")
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create analysis_cache table: %w", err)
	}

	return &AnalysisCache{db: db}, nil
}

// Close closes the underlying database.
func (c *AnalysisCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// IsStale checks if the file has changed since the last analysis run.
// returns stale, currentHash, cachedFindings, err
func (c *AnalysisCache) IsStale(filePath, analyzerVersion, dependencyState, twinVersion string) (bool, string, []Finding, error) {
	// Calculate current file hash
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return true, "", nil, err
	}
	if fileInfo.IsDir() {
		return false, "", nil, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return true, "", nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return true, "", nil, err
	}
	currentHash := hex.EncodeToString(h.Sum(nil))

	c.mu.RLock()
	defer c.mu.RUnlock()

	row := c.db.QueryRow(`
		SELECT file_hash, analyzer_version, dependency_state, twin_version, findings_generated, stale 
		FROM analysis_cache WHERE file_path = ?`, filePath)

	var storedHash, storedVersion, storedDep, storedTwin, storedFindings string
	var stale int
	if err := row.Scan(&storedHash, &storedVersion, &storedDep, &storedTwin, &storedFindings, &stale); err != nil {
		if err == sql.ErrNoRows {
			return true, currentHash, nil, nil
		}
		return true, currentHash, nil, err
	}

	// Invalidation rules:
	// 1. Explicitly marked stale
	if stale == 1 {
		return true, currentHash, nil, nil
	}
	// 2. CHANGED file hash
	if currentHash != storedHash {
		return true, currentHash, nil, nil
	}
	// 3. ANALYZER VERSION CHANGE
	if analyzerVersion != storedVersion {
		return true, currentHash, nil, nil
	}
	// 4. DEPENDENCY CHANGE
	if dependencyState != storedDep {
		return true, currentHash, nil, nil
	}
	// 5. TWIN VERSION CHANGE
	if twinVersion != storedTwin {
		return true, currentHash, nil, nil
	}

	// Cache Hit: Unpack findings
	var findings []Finding
	if storedFindings != "" && storedFindings != "null" {
		if err := json.Unmarshal([]byte(storedFindings), &findings); err != nil {
			// fallback on corruption: treat as stale
			return true, currentHash, nil, nil
		}
	}

	return false, currentHash, findings, nil
}

func (c *AnalysisCache) SetEntry(e CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	staleInt := 0
	if e.Stale {
		staleInt = 1
	}
	query := `INSERT OR REPLACE INTO analysis_cache (
		file_path, file_hash, ast_hash, analyzer_version, dependency_state, twin_version, 
		finding_count, findings_generated, analyzer_status, analyzed_at, stale
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := c.db.Exec(query,
		e.FilePath, e.FileHash, e.ASTHash, e.AnalyzerVersion, e.DependencyState, e.TwinVersion,
		e.FindingCount, e.FindingsGenerated, e.AnalyzerStatus, e.AnalyzedAt, staleInt,
	)
	return err
}

func (c *AnalysisCache) InvalidateAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec("UPDATE analysis_cache SET stale = 1")
	return err
}

func (c *AnalysisCache) GetStats() (CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var stats CacheStats
	row := c.db.QueryRow("SELECT COUNT(*), COALESCE(SUM(stale), 0) FROM analysis_cache")
	var total, stale int
	if err := row.Scan(&total, &stale); err != nil {
		return stats, err
	}
	stats.TotalEntries = total
	stats.StaleEntries = stale

	row = c.db.QueryRow("PRAGMA page_count;")
	var pageCount int64
	if err := row.Scan(&pageCount); err == nil {
		row = c.db.QueryRow("PRAGMA page_size;")
		var pageSize int64
		if err := row.Scan(&pageSize); err == nil {
			stats.SizeBytes = pageCount * pageSize
		}
	}
	return stats, nil
}
