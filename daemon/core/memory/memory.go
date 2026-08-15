package memory

import (
	"sync"
	"time"

	"daemon/core/domain"
)

// KnowledgeRecord represents a structured historical engineering outcome.
type KnowledgeRecord struct {
	ID             string            `json:"id"`
	ErrorSignature string            `json:"error_signature"`
	Finding        string            `json:"finding"`
	Evidence       string            `json:"evidence"`
	Fix            string            `json:"fix"`
	Action         string            `json:"action"`
	Outcome        string            `json:"outcome"`
	Rollback       string            `json:"rollback"`
	Duration       time.Duration     `json:"duration"`
	Scope          string            `json:"scope"` // "personal", "project", "organization", "generic"
	Confidence     float64           `json:"confidence"`
	Provenance     domain.Provenance `json:"provenance"`
}

// MemoryEngine manages durable queryable workflow memory and fix logs.
type MemoryEngine struct {
	mu      sync.RWMutex
	records map[string][]KnowledgeRecord
}

// NewMemoryEngine creates a new MemoryEngine instance.
func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		records: make(map[string][]KnowledgeRecord),
	}
}

// RecordFix logs an executed fix outcome under its error signature.
func (me *MemoryEngine) RecordFix(record KnowledgeRecord) {
	me.mu.Lock()
	defer me.mu.Unlock()

	sig := record.ErrorSignature
	me.records[sig] = append(me.records[sig], record)
}

// QueryRecords retrieves recorded knowledge matching an error signature.
func (me *MemoryEngine) QueryRecords(signature string) []KnowledgeRecord {
	me.mu.RLock()
	defer me.mu.RUnlock()

	if recs, ok := me.records[signature]; ok {
		res := make([]KnowledgeRecord, len(recs))
		copy(res, recs)
		return res
	}
	return []KnowledgeRecord{}
}
