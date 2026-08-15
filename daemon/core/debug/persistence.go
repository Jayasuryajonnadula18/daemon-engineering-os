package debug

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type DebugStore struct {
	mu       *sync.Mutex
	db       *sql.DB
	isShared bool
}

func NewDebugStore(dbPath string) (*DebugStore, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")

	query := `CREATE TABLE IF NOT EXISTS debug_investigations (
		id TEXT PRIMARY KEY,
		problem TEXT,
		status TEXT,
		reason TEXT,
		started_at TEXT,
		completed_at TEXT,
		ai_enhanced INTEGER,
		insufficient_context INTEGER,
		findings TEXT,
		hypotheses TEXT,
		experiments TEXT,
		evidence TEXT,
		root_causes TEXT,
		recommendations TEXT,
		verification TEXT,
		confidence REAL,
		budget TEXT,
		iterations INTEGER,
		files_inspected INTEGER,
		tests_executed INTEGER,
		ai_requests_count INTEGER,
		duration_ms INTEGER
	);`

	if _, err := db.Exec(query); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DebugStore{
		mu:       &sync.Mutex{},
		db:       db,
		isShared: false,
	}, nil
}

func NewDebugStoreFromDB(db *sql.DB) (*DebugStore, error) {
	query := `CREATE TABLE IF NOT EXISTS debug_investigations (
		id TEXT PRIMARY KEY,
		problem TEXT,
		status TEXT,
		reason TEXT,
		started_at TEXT,
		completed_at TEXT,
		ai_enhanced INTEGER,
		insufficient_context INTEGER,
		findings TEXT,
		hypotheses TEXT,
		experiments TEXT,
		evidence TEXT,
		root_causes TEXT,
		recommendations TEXT,
		verification TEXT,
		confidence REAL,
		budget TEXT,
		iterations INTEGER,
		files_inspected INTEGER,
		tests_executed INTEGER,
		ai_requests_count INTEGER,
		duration_ms INTEGER
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &DebugStore{
		mu:       &sync.Mutex{},
		db:       db,
		isShared: true,
	}, nil
}

func (ds *DebugStore) Close() error {
	if ds.isShared {
		return nil
	}
	return ds.db.Close()
}

func (ds *DebugStore) SaveInvestigation(inv *Investigation) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	findingsJSON, _ := json.Marshal(inv.Findings)
	hypothesesJSON, _ := json.Marshal(inv.Hypotheses)
	experimentsJSON, _ := json.Marshal(inv.Experiments)
	evidenceJSON, _ := json.Marshal(inv.Evidence)
	rootCausesJSON, _ := json.Marshal(inv.RootCauses)
	recommendationsJSON, _ := json.Marshal(inv.Recommendations)
	verificationJSON, _ := json.Marshal(inv.Verification)
	budgetJSON, _ := json.Marshal(inv.Budget)

	var completedAtStr string
	if inv.CompletedAt != nil {
		completedAtStr = inv.CompletedAt.Format(time.RFC3339)
	}

	aiInt := 0
	if inv.AIEnhanced {
		aiInt = 1
	}

	insufficientInt := 0
	if inv.InsufficientContext {
		insufficientInt = 1
	}

	query := `INSERT OR REPLACE INTO debug_investigations (
		id, problem, status, reason, started_at, completed_at, ai_enhanced, insufficient_context,
		findings, hypotheses, experiments, evidence, root_causes, recommendations, verification,
		confidence, budget, iterations, files_inspected, tests_executed, ai_requests_count, duration_ms
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := ds.db.Exec(query,
		inv.ID, inv.Problem, string(inv.Status), inv.Reason, inv.StartedAt.Format(time.RFC3339), completedAtStr,
		aiInt, insufficientInt, string(findingsJSON), string(hypothesesJSON), string(experimentsJSON),
		string(evidenceJSON), string(rootCausesJSON), string(recommendationsJSON), string(verificationJSON),
		inv.Confidence, string(budgetJSON), inv.Iterations, inv.FilesInspected, inv.TestsExecuted,
		inv.AIRequestsCount, inv.DurationMs,
	)

	return err
}

func (ds *DebugStore) getInvestigationUnlocked(id string) (*Investigation, error) {
	query := `SELECT id, problem, status, reason, started_at, completed_at, ai_enhanced, insufficient_context,
		findings, hypotheses, experiments, evidence, root_causes, recommendations, verification,
		confidence, budget, iterations, files_inspected, tests_executed, ai_requests_count, duration_ms 
		FROM debug_investigations WHERE id = ?`

	var sID, problem, statusStr, reason, startedStr, completedStr string
	var aiInt, insufficientInt, iterations, filesInspected, testsExecuted, aiRequestsCount int
	var durationMs int64
	var confidence float64
	var findingsStr, hypothesesStr, experimentsStr, evidenceStr, rootCausesStr, recommendationsStr, verificationStr, budgetStr string

	err := ds.db.QueryRow(query, id).Scan(
		&sID, &problem, &statusStr, &reason, &startedStr, &completedStr, &aiInt, &insufficientInt,
		&findingsStr, &hypothesesStr, &experimentsStr, &evidenceStr, &rootCausesStr, &recommendationsStr, &verificationStr,
		&confidence, &budgetStr, &iterations, &filesInspected, &testsExecuted, &aiRequestsCount, &durationMs,
	)
	if err != nil {
		return nil, err
	}

	startedAt, _ := time.Parse(time.RFC3339, startedStr)
	var completedAt *time.Time
	if completedStr != "" {
		t, err := time.Parse(time.RFC3339, completedStr)
		if err == nil {
			completedAt = &t
		}
	}

	inv := &Investigation{
		ID:                  sID,
		Problem:             problem,
		Status:              InvestigationState(statusStr),
		Reason:              reason,
		StartedAt:           startedAt,
		CompletedAt:         completedAt,
		AIEnhanced:          aiInt == 1,
		InsufficientContext: insufficientInt == 1,
		Confidence:          confidence,
		Iterations:          iterations,
		FilesInspected:      filesInspected,
		TestsExecuted:       testsExecuted,
		AIRequestsCount:     aiRequestsCount,
		DurationMs:          durationMs,
	}

	_ = json.Unmarshal([]byte(findingsStr), &inv.Findings)
	_ = json.Unmarshal([]byte(hypothesesStr), &inv.Hypotheses)
	_ = json.Unmarshal([]byte(experimentsStr), &inv.Experiments)
	_ = json.Unmarshal([]byte(evidenceStr), &inv.Evidence)
	_ = json.Unmarshal([]byte(rootCausesStr), &inv.RootCauses)
	_ = json.Unmarshal([]byte(recommendationsStr), &inv.Recommendations)
	_ = json.Unmarshal([]byte(verificationStr), &inv.Verification)
	_ = json.Unmarshal([]byte(budgetStr), &inv.Budget)

	return inv, nil
}

func (ds *DebugStore) GetInvestigation(id string) (*Investigation, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	return ds.getInvestigationUnlocked(id)
}

// FindDuplicate finds if there's an existing recent investigation with the same normalized problem
func (ds *DebugStore) FindDuplicate(problem string) (*Investigation, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	query := `SELECT id FROM debug_investigations WHERE problem = ? ORDER BY started_at DESC LIMIT 1`
	var id string
	err := ds.db.QueryRow(query, problem).Scan(&id)
	if err != nil {
		return nil, err
	}

	return ds.getInvestigationUnlocked(id)
}
