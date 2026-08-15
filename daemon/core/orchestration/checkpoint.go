package orchestration

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type NodeCheckpoint struct {
	ExecutionID     string     `json:"execution_id"`
	DAGID           string     `json:"dag_id"`
	NodeID          string     `json:"node_id"`
	Attempt         int        `json:"attempt"`
	Status          NodeStatus `json:"status"`
	InputHash       string     `json:"input_hash"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	OutputRef       string     `json:"output_ref"`
	VerificationRef string     `json:"verification_ref"`
}

type CheckpointStore struct {
	mu *sync.Mutex
	db *sql.DB
}

func NewCheckpointStore(dbPath string) (*CheckpointStore, error) {
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

	query := `CREATE TABLE IF NOT EXISTS node_checkpoints (
		execution_id TEXT NOT NULL,
		dag_id TEXT NOT NULL,
		node_id TEXT NOT NULL,
		attempt INTEGER NOT NULL,
		status TEXT NOT NULL,
		input_hash TEXT,
		started_at TEXT NOT NULL,
		completed_at TEXT,
		output_ref TEXT,
		verification_ref TEXT,
		PRIMARY KEY (execution_id, node_id)
	);`

	if _, err := db.Exec(query); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &CheckpointStore{
		mu: &sync.Mutex{},
		db: db,
	}, nil
}

func (cs *CheckpointStore) Close() error {
	return cs.db.Close()
}

func (cs *CheckpointStore) SaveCheckpoint(cp NodeCheckpoint) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var compStr string
	if cp.CompletedAt != nil {
		compStr = cp.CompletedAt.Format(time.RFC3339)
	}

	query := `INSERT OR REPLACE INTO node_checkpoints 
		(execution_id, dag_id, node_id, attempt, status, input_hash, started_at, completed_at, output_ref, verification_ref)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := cs.db.Exec(query, cp.ExecutionID, cp.DAGID, cp.NodeID, cp.Attempt, string(cp.Status), cp.InputHash, cp.StartedAt.Format(time.RFC3339), compStr, cp.OutputRef, cp.VerificationRef)
	return err
}

func (cs *CheckpointStore) GetCheckpoints(executionID string) ([]NodeCheckpoint, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	query := `SELECT execution_id, dag_id, node_id, attempt, status, 
		COALESCE(input_hash, ''), started_at, COALESCE(completed_at, ''), 
		COALESCE(output_ref, ''), COALESCE(verification_ref, '') 
		FROM node_checkpoints WHERE execution_id = ?`

	rows, err := cs.db.Query(query, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []NodeCheckpoint
	for rows.Next() {
		var execID, dagID, nodeID, statusStr, inHash, startStr, compStr, outRef, verRef string
		var att int

		if err := rows.Scan(&execID, &dagID, &nodeID, &att, &statusStr, &inHash, &startStr, &compStr, &outRef, &verRef); err != nil {
			return nil, err
		}

		sTime, _ := time.Parse(time.RFC3339, startStr)
		var cTime *time.Time
		if compStr != "" {
			t, _ := time.Parse(time.RFC3339, compStr)
			cTime = &t
		}

		list = append(list, NodeCheckpoint{
			ExecutionID:     execID,
			DAGID:           dagID,
			NodeID:          nodeID,
			Attempt:         att,
			Status:          NodeStatus(statusStr),
			InputHash:       inHash,
			StartedAt:       sTime,
			CompletedAt:     cTime,
			OutputRef:       outRef,
			VerificationRef: verRef,
		})
	}
	return list, nil
}
