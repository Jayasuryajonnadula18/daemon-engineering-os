package storage

import (
	"database/sql"
	"daemon/core/domain"
	"time"
)

type DatabaseProvider interface {
	DB() *sql.DB
}

// GraphStore defines methods to write and read the Engineering Knowledge Graph.
type GraphStore interface {
	AddNode(nodeType, id, name string, properties map[string]string) error
	AddEdge(fromType, fromID, toType, toID, relation string) error
	GetNodes(nodeType string) ([]domain.Module, error)      // maps generic nodes to module structure or list
	GetAllNodes() ([]domain.Module, error)
	GetEdges() ([]domain.EdgeRecord, error)
	GetServices() ([]domain.Service, error)
	GetDependencies() ([]domain.Dependency, error)
	GetAPIs() ([]domain.API, error)
	Clear() error
}

// MemoryStore records historical incidents, accepted fixes, and deployment logs.
type MemoryStore interface {
	AddIncident(incident *domain.Incident) error
	GetIncidents() ([]domain.Incident, error)
	ResolveIncident(id string) error
	AddDeployment(deployment *domain.Deployment) error
	GetDeployments() ([]domain.Deployment, error)
	AddRecommendation(rec *domain.Recommendation) error
	GetRecommendations() ([]domain.Recommendation, error)
}

// ConfigStore represents project local configuration parameters.
type ConfigStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
	All() (map[string]string, error)
}

// MetricsStore saves performance telemetry and operational counters.
type MetricsStore interface {
	RecordMetric(name string, value float64, timestamp time.Time) error
}

// CacheStore maintains runtime key-value states.
type CacheStore interface {
	Get(key string) ([]byte, error)
	Set(key string, val []byte, ttl time.Duration) error
	Delete(key string) error
}

// AuditStore registers logs of administrative/policy decisions.
type AuditStore interface {
	LogAction(action string, allowed bool, reason string, timestamp time.Time) error
	GetLogs() ([]string, error)
}

