package instruments

type Capability string

const (
	CapDebug               Capability = "DEBUG"
	CapStackInspection     Capability = "STACK_INSPECTION"
	CapBreakpoint          Capability = "BREAKPOINT"
	CapThreadInspection    Capability = "THREAD_INSPECTION"
	CapHeapAnalysis        Capability = "HEAP_ANALYSIS"
	CapCPUProfiling        Capability = "CPU_PROFILING"
	CapMemoryProfiling     Capability = "MEMORY_PROFILING"
	CapAllocationAnalysis  Capability = "ALLOCATION_ANALYSIS"
	CapGoroutineAnalysis   Capability = "GOROUTINE_ANALYSIS"
	CapLockContention      Capability = "LOCK_CONTENTION"
	CapRaceDetection       Capability = "RACE_DETECTION"
	CapDeadlockAnalysis    Capability = "DEADLOCK_ANALYSIS"
	CapTraceCollection     Capability = "TRACE_COLLECTION"
	CapTraceAnalysis       Capability = "TRACE_ANALYSIS"
	CapUnitTesting         Capability = "UNIT_TESTING"
	CapIntegrationTesting  Capability = "INTEGRATION_TESTING"
	CapRegressionTesting   Capability = "REGRESSION_TESTING"
	CapBenchmarking        Capability = "BENCHMARKING"
	CapFuzzing             Capability = "FUZZING"
	CapCoverage            Capability = "COVERAGE"
	CapStaticAnalysis      Capability = "STATIC_ANALYSIS"
	CapSecurityAnalysis    Capability = "SECURITY_ANALYSIS"
	CapDependencyAnalysis  Capability = "DEPENDENCY_ANALYSIS"
	CapBuild               Capability = "BUILD"
	CapPackageAnalysis     Capability = "PACKAGE_ANALYSIS"
	CapProcessInspection   Capability = "PROCESS_INSPECTION"
	CapPortInspection      Capability = "PORT_INSPECTION"
	CapContainerInspection Capability = "CONTAINER_INSPECTION"
	CapLogAnalysis         Capability = "LOG_ANALYSIS"
	CapDatabaseInspection  Capability = "DATABASE_INSPECTION"
	CapNetworkInspection   Capability = "NETWORK_INSPECTION"
	CapArchitectureAnalysis Capability = "ARCHITECTURE_ANALYSIS"
	CapCodeNavigation      Capability = "CODE_NAVIGATION"
)

type ToolAvailabilityState string

const (
	StateAdapterExists      ToolAvailabilityState = "ADAPTER_EXISTS"
	StateToolDiscovered     ToolAvailabilityState = "TOOL_DISCOVERED"
	StateToolInstalled      ToolAvailabilityState = "TOOL_INSTALLED"
	StateToolHealthUnknown  ToolAvailabilityState = "TOOL_HEALTH_UNKNOWN"
	StateProjectCompatible  ToolAvailabilityState = "PROJECT_COMPATIBLE"
	StateCapabilityAvailable ToolAvailabilityState = "CAPABILITY_AVAILABLE"
)

// ProjectCapabilityProfile represents the detected capabilities map for a target technology stack.
type ProjectCapabilityProfile struct {
	Languages      []string                         `json:"languages"`
	Runtimes       []string                         `json:"runtimes"`
	BuildSystems   []string                         `json:"build_systems"`
	Matrix         map[Capability]map[string]string `json:"matrix"` // Capability -> Tool -> State
}

func NewProjectCapabilityProfile() *ProjectCapabilityProfile {
	return &ProjectCapabilityProfile{
		Languages:    []string{},
		Runtimes:     []string{},
		BuildSystems: []string{},
		Matrix:       make(map[Capability]map[string]string),
	}
}
