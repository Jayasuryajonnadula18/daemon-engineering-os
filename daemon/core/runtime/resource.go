package runtime

import (
	"context"
)

// ResourceManager keeps track of OS and network resources used during tasks.
type ResourceManager interface {
	RegisterProcess(pid int) error
	RegisterTempFile(path string) error
	Cleanup(ctx context.Context) error
}

// MemoryResourceManager registers resources in local slices for bulk cleanup.
type MemoryResourceManager struct {
	processes []int
	tempFiles []string
}

// NewMemoryResourceManager instantiates a new MemoryResourceManager.
func NewMemoryResourceManager() *MemoryResourceManager {
	return &MemoryResourceManager{
		processes: make([]int, 0),
		tempFiles: make([]string, 0),
	}
}

// RegisterProcess tracks a child process ID.
func (m *MemoryResourceManager) RegisterProcess(pid int) error {
	m.processes = append(m.processes, pid)
	return nil
}

// RegisterTempFile tracks a temporary filepath.
func (m *MemoryResourceManager) RegisterTempFile(path string) error {
	m.tempFiles = append(m.tempFiles, path)
	return nil
}

// Cleanup performs termination on resources.
func (m *MemoryResourceManager) Cleanup(ctx context.Context) error {
	// In the MVP, this serves as a placeholder interface
	return nil
}

