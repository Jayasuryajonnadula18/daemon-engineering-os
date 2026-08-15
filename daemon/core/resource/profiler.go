package resource

import (
	"runtime"
	"sync"
	"time"
)

type HardwareMetrics struct {
	CPUCores         int     `json:"cpu_cores"`
	CPUUtilization   float64 `json:"cpu_utilization"`
	TotalMemoryMB    uint64  `json:"total_memory_mb"`
	AvailableMemoryMB uint64 `json:"available_memory_mb"`
	GPUAvailable     bool    `json:"gpu_available"`
	FreeDiskGB       uint64  `json:"free_disk_gb"`
	ObservedAt       time.Time `json:"observed_at"`
}

type Profiler struct {
	mu           sync.RWMutex
	lastMetrics  HardwareMetrics
	simulated    bool
	simCPULoad   float64
	simAvailRAM  uint64
}

func NewProfiler() *Profiler {
	p := &Profiler{
		simCPULoad:  0.25, // default 25% CPU
		simAvailRAM: 8192, // default 8GB RAM
	}
	p.Refresh()
	return p
}

func (p *Profiler) SetSimulatedLoad(cpu float64, availRAM uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.simulated = true
	p.simCPULoad = cpu
	p.simAvailRAM = availRAM
	p.lastMetrics.CPUUtilization = cpu
	p.lastMetrics.AvailableMemoryMB = availRAM
}

func (p *Profiler) Refresh() HardwareMetrics {
	p.mu.Lock()
	defer p.mu.Unlock()

	cores := runtime.NumCPU()

	cpuUtil := 0.30
	availRAM := uint64(8192)

	if p.simulated {
		cpuUtil = p.simCPULoad
		availRAM = p.simAvailRAM
	}

	metrics := HardwareMetrics{
		CPUCores:          cores,
		CPUUtilization:    cpuUtil,
		TotalMemoryMB:     16384,
		AvailableMemoryMB: availRAM,
		GPUAvailable:      false,
		FreeDiskGB:        120,
		ObservedAt:        time.Now(),
	}

	p.lastMetrics = metrics
	return metrics
}

func (p *Profiler) GetMetrics() HardwareMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastMetrics
}
