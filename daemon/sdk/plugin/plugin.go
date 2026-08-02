package plugin

// Capability defines the type of work a plugin performs.
type Capability string

const (
	CapDiscovery     Capability = "Discovery"
	CapDiagnosis     Capability = "Diagnosis"
	CapRepair        Capability = "Repair"
	CapDeployment    Capability = "Deployment"
	CapObservation   Capability = "Observation"
	CapDocumentation Capability = "Documentation"
	CapSecurity      Capability = "Security"
	CapMonitoring    Capability = "Monitoring"
)

// Plugin represents the base interface for all daemon plugins.
type Plugin interface {
	ID() string
	Name() string
	Version() string
	Capabilities() []Capability
}

// CapabilityRegistry tracks dynamically registered plugins and their capabilities.
type CapabilityRegistry interface {
	Register(plugin Plugin) error
	GetPluginsByCapability(cap Capability) []Plugin
	AllPlugins() []Plugin
}

// MemoryCapabilityRegistry implements CapabilityRegistry.
type MemoryCapabilityRegistry struct {
	plugins []Plugin
}

// NewMemoryCapabilityRegistry instantiates a new MemoryCapabilityRegistry.
func NewMemoryCapabilityRegistry() *MemoryCapabilityRegistry {
	return &MemoryCapabilityRegistry{plugins: make([]Plugin, 0)}
}

// Register adds a plugin to the in-memory registry.
func (r *MemoryCapabilityRegistry) Register(plugin Plugin) error {
	r.plugins = append(r.plugins, plugin)
	return nil
}

// GetPluginsByCapability filters plugins by advertising capability.
func (r *MemoryCapabilityRegistry) GetPluginsByCapability(cap Capability) []Plugin {
	var matched []Plugin
	for _, p := range r.plugins {
		for _, c := range p.Capabilities() {
			if c == cap {
				matched = append(matched, p)
				break
			}
		}
	}
	return matched
}

// AllPlugins returns all registered plugins.
func (r *MemoryCapabilityRegistry) AllPlugins() []Plugin {
	return r.plugins
}

