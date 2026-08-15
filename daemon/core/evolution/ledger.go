package evolution

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type FixLedgerEntry struct {
	ActionID           string    `json:"action_id"`
	PatternID          string    `json:"pattern_id"`
	ErrorSignature     string    `json:"error_signature"`
	RootCause          string    `json:"root_cause"`
	FixSummary         string    `json:"fix_summary"`
	VerificationResult string    `json:"verification_result"`
	RollbackResult     string    `json:"rollback_result"`
	Environment        string    `json:"environment"`
	Timestamp          time.Time `json:"timestamp"`
	EvidenceIDs        []string  `json:"evidence_ids"`
}

type FixLedger struct {
	mu sync.RWMutex
	db *sql.DB
}

func NewFixLedger(dbPath string) (*FixLedger, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open fix_ledger db: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS fix_ledger (
		action_id TEXT PRIMARY KEY,
		pattern_id TEXT,
		error_signature TEXT,
		root_cause TEXT,
		fix_summary TEXT,
		verification_result TEXT,
		rollback_result TEXT,
		environment TEXT,
		timestamp DATETIME,
		evidence_ids TEXT
	);`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create fix_ledger table: %w", err)
	}

	return &FixLedger{db: db}, nil
}

func (fl *FixLedger) RecordFix(entry FixLedgerEntry) error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// Redact potential secrets
	rootCause := redactSecrets(entry.RootCause)
	fixSummary := redactSecrets(entry.FixSummary)

	evJoined := strings.Join(entry.EvidenceIDs, ",")

	query := `INSERT OR REPLACE INTO fix_ledger (
		action_id, pattern_id, error_signature, root_cause, fix_summary,
		verification_result, rollback_result, environment, timestamp, evidence_ids
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := fl.db.Exec(query,
		entry.ActionID, entry.PatternID, entry.ErrorSignature, rootCause, fixSummary,
		entry.VerificationResult, entry.RollbackResult, entry.Environment, entry.Timestamp, evJoined,
	)
	return err
}

func (fl *FixLedger) GetEntries() ([]FixLedgerEntry, error) {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	rows, err := fl.db.Query(`SELECT action_id, pattern_id, error_signature, root_cause, fix_summary, verification_result, rollback_result, environment, timestamp, evidence_ids FROM fix_ledger ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []FixLedgerEntry
	for rows.Next() {
		var e FixLedgerEntry
		var evJoined string
		if err := rows.Scan(&e.ActionID, &e.PatternID, &e.ErrorSignature, &e.RootCause, &e.FixSummary, &e.VerificationResult, &e.RollbackResult, &e.Environment, &e.Timestamp, &evJoined); err != nil {
			continue
		}
		if evJoined != "" {
			e.EvidenceIDs = strings.Split(evJoined, ",")
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func redactSecrets(input string) string {
	secrets := []string{"password=", "token=", "bearer ", "secret=", "api_key="}
	res := input
	for _, s := range secrets {
		if strings.Contains(strings.ToLower(res), s) {
			res = "[REDACTED_SECRET]"
			break
		}
	}
	return res
}
