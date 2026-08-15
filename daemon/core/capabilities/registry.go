package capabilities

import (
	"fmt"
	"sync"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

type Capability struct {
	Name          string
	Description   string
	Preconditions []string
	Inputs        []string
	Risk          RiskLevel
	Execution     string
	Verification  string
	Rollback      string
}

type Registry struct {
	mu           sync.RWMutex
	capabilities map[string]Capability
}

func NewRegistry() *Registry {
	return &Registry{capabilities: make(map[string]Capability)}
}

func (r *Registry) Register(cap Capability) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cap.Name == "" {
		return fmt.Errorf("capability name is required")
	}
	r.capabilities[cap.Name] = cap
	return nil
}

func (r *Registry) Get(name string) (Capability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cap, ok := r.capabilities[name]
	return cap, ok
}

func (r *Registry) List() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Capability, 0, len(r.capabilities))
	for _, cap := range r.capabilities {
		result = append(result, cap)
	}
	return result
}
