package intelligence

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type AutomationOpportunity struct {
	PatternID        string   `json:"pattern_id"`
	Sequence         []string `json:"sequence"`
	OccurrencesCount int      `json:"occurrences_count"`
	AverageDuration  string   `json:"average_duration"`
	Confidence       float64  `json:"confidence"`
	OpportunityScore string   `json:"opportunity_score"`
}

type IntelligenceStateStore struct {
	mu       *sync.Mutex
	db       *sql.DB
	isShared bool
}

func NewIntelligenceStateStore(dbPath string) (*IntelligenceStateStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS intelligence_patterns (
			pattern_id TEXT PRIMARY KEY,
			sequence TEXT NOT NULL,
			occurrences_count INTEGER NOT NULL,
			average_duration TEXT,
			confidence REAL NOT NULL,
			opportunity_score TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS intelligence_predictions (
			id TEXT PRIMARY KEY,
			action TEXT NOT NULL,
			target TEXT NOT NULL,
			probability REAL NOT NULL,
			rationale TEXT,
			baseline TEXT,
			created_at TEXT NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &IntelligenceStateStore{
		mu:       &sync.Mutex{},
		db:       db,
		isShared: false,
	}, nil
}

func NewIntelligenceStateStoreFromDB(db *sql.DB) (*IntelligenceStateStore, error) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS intelligence_patterns (
			pattern_id TEXT PRIMARY KEY,
			sequence TEXT NOT NULL,
			occurrences_count INTEGER NOT NULL,
			average_duration TEXT,
			confidence REAL NOT NULL,
			opportunity_score TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS intelligence_predictions (
			id TEXT PRIMARY KEY,
			action TEXT NOT NULL,
			target TEXT NOT NULL,
			probability REAL NOT NULL,
			rationale TEXT,
			baseline TEXT,
			created_at TEXT NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return nil, err
		}
	}

	return &IntelligenceStateStore{
		mu:       &sync.Mutex{},
		db:       db,
		isShared: true,
	}, nil
}

func (s *IntelligenceStateStore) Close() error {
	if s.isShared {
		return nil
	}
	return s.db.Close()
}

func (s *IntelligenceStateStore) SaveOpportunity(opp AutomationOpportunity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	seqJSON, err := json.Marshal(opp.Sequence)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO intelligence_patterns (pattern_id, sequence, occurrences_count, average_duration, confidence, opportunity_score, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, opp.PatternID, string(seqJSON), opp.OccurrencesCount, opp.AverageDuration, opp.Confidence, opp.OpportunityScore, time.Now().Format(time.RFC3339))
	return err
}

func (s *IntelligenceStateStore) GetOpportunities() ([]AutomationOpportunity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT pattern_id, sequence, occurrences_count, average_duration, confidence, opportunity_score FROM intelligence_patterns`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AutomationOpportunity
	for rows.Next() {
		var id, seqStr, durStr, oppStr string
		var occ int
		var conf float64
		if err := rows.Scan(&id, &seqStr, &occ, &durStr, &conf, &oppStr); err != nil {
			return nil, err
		}

		var seq []string
		_ = json.Unmarshal([]byte(seqStr), &seq)

		list = append(list, AutomationOpportunity{
			PatternID:        id,
			Sequence:         seq,
			OccurrencesCount: occ,
			AverageDuration:  durStr,
			Confidence:       conf,
			OpportunityScore: oppStr,
		})
	}
	return list, nil
}
