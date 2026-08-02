package integrations

import (
	"context"
)

// Lifecycle represents the current integration state.
type Lifecycle string

const (
	StateRegistered    Lifecycle = "Registered"
	StateConfigured    Lifecycle = "Configured"
	StateAuthenticating Lifecycle = "Authenticating"
	StateConnected     Lifecycle = "Connected"
	StateSynchronizing Lifecycle = "Synchronizing"
	StateObserving     Lifecycle = "Observing"
	StateHealthy       Lifecycle = "Healthy"
	StateDegraded      Lifecycle = "Degraded"
	StateDisconnected  Lifecycle = "Disconnected"
	StateReconnecting  Lifecycle = "Reconnecting"
	StateFailed        Lifecycle = "Failed"
)

// Capability flags declared by connectors.
type Capability string

const (
	CapRead       Capability = "Read"
	CapWrite      Capability = "Write"
	CapObserve    Capability = "Observe"
	CapSearch     Capability = "Search"
	CapEvents     Capability = "Events"
	CapAutomation Capability = "Automation"
)

// Resource represents a provider-independent Engineering Resource stored in the twin graph.
type Resource struct {
	Type    string            `json:"type"`
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Metrics map[string]string `json:"metrics"`
}

// Connector defines standard integration methods.
type Connector interface {
	ID() string
	Capabilities() []Capability
	Connect(ctx context.Context) error
	Authenticate(ctx context.Context) (bool, error)
	Discover(ctx context.Context) ([]Resource, error)
	Synchronize(ctx context.Context) error
	Observe(ctx context.Context, eventChan chan<- string) error
	Execute(ctx context.Context, action string, args []string) (string, error)
	Health(ctx context.Context) (Lifecycle, int, error) // state, latency (ms), error
	Disconnect(ctx context.Context) error
	Reconnect(ctx context.Context) error
}
