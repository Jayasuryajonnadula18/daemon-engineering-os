package agent

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type AgentBudget struct {
	MaxIterations        int     `json:"max_iterations"`
	MaxToolCalls         int     `json:"max_tool_calls"`
	MaxDuration          float64 `json:"max_duration_seconds"`
	MaxTokens            int     `json:"max_tokens"`
	MaxEstimatedCost     float64 `json:"max_estimated_cost"`
	MaxParallelTools     int     `json:"max_parallel_tools"`
	MaxRetries           int     `json:"max_retries"`
	MaxExperiments       int     `json:"max_experiments"`
	MaxWrites            int     `json:"max_writes"`
	MaxNetworkOperations int     `json:"max_network_operations"`
	MaxRiskLevel         string  `json:"max_risk_level"`
}

type AgentSession struct {
	ID                string            `json:"id"`
	Intent            string            `json:"intent"`
	State             string            `json:"state"`
	Budget            AgentBudget       `json:"budget"`
	PlanRef           string            `json:"plan_ref"`
	ExecutionID       string            `json:"execution_id"`
	FinalResult       string            `json:"final_result"`
	Failure           string            `json:"failure"`
	CancellationState bool              `json:"cancellation_state"`
	AIEnhanced        bool              `json:"ai_enhanced"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type SessionStore struct {
	mu       *sync.Mutex
	db       *sql.DB
	isShared bool
}

func NewSessionStore(dbPath string) (*SessionStore, error) {
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

	query := `CREATE TABLE IF NOT EXISTS agent_sessions (
		id TEXT PRIMARY KEY,
		intent TEXT,
		state TEXT,
		budget TEXT,
		plan_ref TEXT,
		execution_id TEXT,
		final_result TEXT,
		failure TEXT,
		cancellation_state INTEGER,
		ai_enhanced INTEGER,
		created_at TEXT,
		updated_at TEXT
	);`

	if _, err := db.Exec(query); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SessionStore{
		mu: &sync.Mutex{},
		db: db,
	}, nil
}

func NewSessionStoreFromDB(db *sql.DB) *SessionStore {
	return &SessionStore{
		mu:       &sync.Mutex{},
		db:       db,
		isShared: true,
	}
}

func (ss *SessionStore) Close() error {
	if ss.isShared {
		return nil
	}
	return ss.db.Close()
}

func (ss *SessionStore) SaveSession(s AgentSession) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	budgetJSON, err := json.Marshal(s.Budget)
	if err != nil {
		return err
	}

	cancelInt := 0
	if s.CancellationState {
		cancelInt = 1
	}

	aiInt := 0
	if s.AIEnhanced {
		aiInt = 1
	}

	query := `INSERT OR REPLACE INTO agent_sessions 
		(id, intent, state, budget, plan_ref, execution_id, final_result, failure, cancellation_state, ai_enhanced, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = ss.db.Exec(query, s.ID, s.Intent, s.State, string(budgetJSON), s.PlanRef, s.ExecutionID, s.FinalResult, s.Failure, cancelInt, aiInt, s.CreatedAt.Format(time.RFC3339), s.UpdatedAt.Format(time.RFC3339))
	return err
}

func (ss *SessionStore) GetSession(id string) (*AgentSession, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	query := `SELECT id, intent, state, budget, plan_ref, execution_id, final_result, failure, cancellation_state, ai_enhanced, created_at, updated_at 
		FROM agent_sessions WHERE id = ?`

	var sID, intent, state, budgetStr, planRef, execID, finalRes, failure, createdStr, updatedStr string
	var cancelInt, aiInt int

	err := ss.db.QueryRow(query, id).Scan(&sID, &intent, &state, &budgetStr, &planRef, &execID, &finalRes, &failure, &cancelInt, &aiInt, &createdStr, &updatedStr)
	if err != nil {
		return nil, err
	}

	var budget AgentBudget
	_ = json.Unmarshal([]byte(budgetStr), &budget)

	cTime, _ := time.Parse(time.RFC3339, createdStr)
	uTime, _ := time.Parse(time.RFC3339, updatedStr)

	return &AgentSession{
		ID:                sID,
		Intent:            intent,
		State:             state,
		Budget:            budget,
		PlanRef:           planRef,
		ExecutionID:       execID,
		FinalResult:       finalRes,
		Failure:           failure,
		CancellationState: cancelInt == 1,
		AIEnhanced:        aiInt == 1,
		CreatedAt:         cTime,
		UpdatedAt:         uTime,
	}, nil
}

func (ss *SessionStore) ListSessions() ([]AgentSession, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	query := `SELECT id, intent, state, budget, plan_ref, execution_id, final_result, failure, cancellation_state, ai_enhanced, created_at, updated_at 
		FROM agent_sessions`

	rows, err := ss.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AgentSession
	for rows.Next() {
		var sID, intent, state, budgetStr, planRef, execID, finalRes, failure, createdStr, updatedStr string
		var cancelInt, aiInt int

		if err := rows.Scan(&sID, &intent, &state, &budgetStr, &planRef, &execID, &finalRes, &failure, &cancelInt, &aiInt, &createdStr, &updatedStr); err != nil {
			return nil, err
		}

		var budget AgentBudget
		_ = json.Unmarshal([]byte(budgetStr), &budget)

		cTime, _ := time.Parse(time.RFC3339, createdStr)
		uTime, _ := time.Parse(time.RFC3339, updatedStr)

		list = append(list, AgentSession{
			ID:                sID,
			Intent:            intent,
			State:             state,
			Budget:            budget,
			PlanRef:           planRef,
			ExecutionID:       execID,
			FinalResult:       finalRes,
			Failure:           failure,
			CancellationState: cancelInt == 1,
			AIEnhanced:        aiInt == 1,
			CreatedAt:         cTime,
			UpdatedAt:         uTime,
		})
	}
	return list, nil
}
