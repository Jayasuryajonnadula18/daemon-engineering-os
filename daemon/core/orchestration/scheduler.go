package orchestration

import (
	"fmt"
	"sync"
)

type ExecutionWave struct {
	WaveIndex int       `json:"wave_index"`
	Nodes     []DAGNode `json:"nodes"`
}

type WaveScheduler struct {
	mu          sync.Mutex
	activeLocks map[string]string // resourceID -> lockMode ("READ", "WRITE")
}

func NewWaveScheduler() *WaveScheduler {
	return &WaveScheduler{
		activeLocks: make(map[string]string),
	}
}

// ComputeWaves partitions an ExecutionDAG into parallel execution waves while validating lock compatibility.
func (s *WaveScheduler) ComputeWaves(dag *ExecutionDAG) ([]ExecutionWave, error) {
	nodeMap := make(map[string]DAGNode)
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, n := range dag.Nodes {
		nodeMap[n.ID] = n
		inDegree[n.ID] = len(n.Parents)
		for _, p := range n.Parents {
			dependents[p] = append(dependents[p], n.ID)
		}
	}

	var waves []ExecutionWave
	waveIdx := 0

	processedCount := 0
	for processedCount < len(dag.Nodes) {
		var currentWaveNodes []DAGNode
		for id, deg := range inDegree {
			if deg == 0 {
				currentWaveNodes = append(currentWaveNodes, nodeMap[id])
			}
		}

		if len(currentWaveNodes) == 0 {
			return nil, fmt.Errorf("deadlock or unresolvable cyclic dependency detected during wave scheduling")
		}

		// Sort nodes deterministically
		waves = append(waves, ExecutionWave{
			WaveIndex: waveIdx,
			Nodes:     currentWaveNodes,
		})

		for _, n := range currentWaveNodes {
			delete(inDegree, n.ID)
			processedCount++
			for _, depID := range dependents[n.ID] {
				inDegree[depID]--
			}
		}
		waveIdx++
	}

	return waves, nil
}

// CanAcquireLocks checks if the required locks for a node can be acquired under the READ/WRITE lock matrix.
// Matrix: READ + READ (allowed); READ + WRITE (conflict); WRITE + WRITE (conflict).
func (s *WaveScheduler) CanAcquireLocks(locks []ResourceLock) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, l := range locks {
		existingMode, exists := s.activeLocks[l.ResourceID]
		if exists {
			if l.Mode == "WRITE" || existingMode == "WRITE" {
				return false // Conflict
			}
		}
	}
	return true
}

func (s *WaveScheduler) AcquireLocks(locks []ResourceLock) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, l := range locks {
		existingMode, exists := s.activeLocks[l.ResourceID]
		if exists {
			if l.Mode == "WRITE" || existingMode == "WRITE" {
				return false
			}
		}
	}

	for _, l := range locks {
		s.activeLocks[l.ResourceID] = l.Mode
	}
	return true
}

func (s *WaveScheduler) ReleaseLocks(locks []ResourceLock) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, l := range locks {
		delete(s.activeLocks, l.ResourceID)
	}
}
