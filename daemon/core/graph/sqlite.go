package graph

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"daemon/core/domain"
	_ "modernc.org/sqlite"
)

// SQLiteStore coordinates database persistence for the Knowledge Graph and Memory store.
type SQLiteStore struct {
	mu *sync.Mutex
	db *sql.DB
}

// NewSQLiteStore opens or creates a new SQLite database at dbPath and runs table creation queries.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			type TEXT NOT NULL,
			id TEXT NOT NULL,
			name TEXT NOT NULL,
			properties TEXT,
			PRIMARY KEY (type, id)
		);`,
		`CREATE TABLE IF NOT EXISTS edges (
			from_type TEXT NOT NULL,
			from_id TEXT NOT NULL,
			to_type TEXT NOT NULL,
			to_id TEXT NOT NULL,
			relation TEXT NOT NULL,
			PRIMARY KEY (from_type, from_id, to_type, to_id, relation)
		);`,
		`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			message TEXT,
			severity TEXT,
			resolved INTEGER,
			detected_at TEXT,
			resolved_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			env TEXT,
			status TEXT,
			version TEXT,
			started_at TEXT,
			ended_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS recommendations (
			id TEXT PRIMARY KEY,
			category TEXT,
			message TEXT,
			fixable INTEGER,
			rationale TEXT
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &SQLiteStore{
		mu: &sync.Mutex{},
		db: db,
	}, nil
}

// Close releases the SQLite connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ==========================================
// GraphStore Implementation
// ==========================================

// AddNode creates or updates a node.
func (s *SQLiteStore) AddNode(nodeType, id, name string, properties map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	propsJSON, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO nodes (type, id, name, properties) VALUES (?, ?, ?, ?)`
	_, err = s.db.Exec(query, nodeType, id, name, string(propsJSON))
	return err
}

// AddEdge creates a relationship link.
func (s *SQLiteStore) AddEdge(fromType, fromID, toType, toID, relation string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT OR REPLACE INTO edges (from_type, from_id, to_type, to_id, relation) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, fromType, fromID, toType, toID, relation)
	return err
}

// GetNodes retrieves module representations.
func (s *SQLiteStore) GetNodes(nodeType string) ([]domain.Module, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, name, properties FROM nodes WHERE type = ?`
	rows, err := s.db.Query(query, nodeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []domain.Module
	for rows.Next() {
		var id, name, propsStr string
		if err := rows.Scan(&id, &name, &propsStr); err != nil {
			return nil, err
		}

		var props map[string]string
		_ = json.Unmarshal([]byte(propsStr), &props)

		var imports []string
		if impVal, exists := props["imports"]; exists {
			_ = json.Unmarshal([]byte(impVal), &imports)
		}

		modules = append(modules, domain.Module{
			ID:      id,
			Name:    name,
			Type:    nodeType,
			Path:    props["path"],
			Imports: imports,
		})
	}
	return modules, nil
}

// GetEdges returns all relationship edges from the edges table.
func (s *SQLiteStore) GetEdges() ([]domain.EdgeRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT from_type, from_id, to_type, to_id, relation FROM edges`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []domain.EdgeRecord
	for rows.Next() {
		var e domain.EdgeRecord
		if err := rows.Scan(&e.FromType, &e.FromID, &e.ToType, &e.ToID, &e.Relation); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, nil
}

// GetAllNodes returns all nodes regardless of type.
func (s *SQLiteStore) GetAllNodes() ([]domain.Module, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT type, id, name, properties FROM nodes`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []domain.Module
	for rows.Next() {
		var nodeType, id, name, propsStr string
		if err := rows.Scan(&nodeType, &id, &name, &propsStr); err != nil {
			return nil, err
		}

		var props map[string]string
		_ = json.Unmarshal([]byte(propsStr), &props)

		modules = append(modules, domain.Module{
			ID:   id,
			Name: name,
			Type: nodeType,
			Path: props["path"],
		})
	}
	return modules, nil
}

// GetServices returns services mapped from nodes table.
func (s *SQLiteStore) GetServices() ([]domain.Service, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, name, properties FROM nodes WHERE type = 'service'`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var id, name, propsStr string
		if err := rows.Scan(&id, &name, &propsStr); err != nil {
			return nil, err
		}

		var props map[string]string
		_ = json.Unmarshal([]byte(propsStr), &props)

		port, _ := strconv.Atoi(props["port"])

		var dependsOn []string
		if depVal, exists := props["depends_on"]; exists {
			_ = json.Unmarshal([]byte(depVal), &dependsOn)
		}

		services = append(services, domain.Service{
			ID:        id,
			Name:      name,
			Port:      port,
			Status:    props["status"],
			DependsOn: dependsOn,
		})
	}
	return services, nil
}

// GetDependencies returns application dependency definitions.
func (s *SQLiteStore) GetDependencies() ([]domain.Dependency, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, name, properties FROM nodes WHERE type = 'dependency'`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependencies []domain.Dependency
	for rows.Next() {
		var id, name, propsStr string
		if err := rows.Scan(&id, &name, &propsStr); err != nil {
			return nil, err
		}

		var props map[string]string
		_ = json.Unmarshal([]byte(propsStr), &props)

		dependencies = append(dependencies, domain.Dependency{
			ID:         id,
			Name:       name,
			Version:    props["version"],
			Type:       props["type"],
			IsOutdated: props["is_outdated"] == "true",
		})
	}
	return dependencies, nil
}

// GetAPIs returns registered routing entrypoints.
func (s *SQLiteStore) GetAPIs() ([]domain.API, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, name, properties FROM nodes WHERE type = 'api'`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apis []domain.API
	for rows.Next() {
		var id, name, propsStr string
		if err := rows.Scan(&id, &name, &propsStr); err != nil {
			return nil, err
		}

		var props map[string]string
		_ = json.Unmarshal([]byte(propsStr), &props)

		apis = append(apis, domain.API{
			ID:      id,
			Path:    props["path"],
			Method:  props["method"],
			Service: props["service"],
		})
	}
	return apis, nil
}

// Clear truncates node and edge records.
func (s *SQLiteStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err1 := s.db.Exec("DELETE FROM nodes")
	_, err2 := s.db.Exec("DELETE FROM edges")
	if err1 != nil {
		return err1
	}
	return err2
}

// ==========================================
// MemoryStore Implementation
// ==========================================

// AddIncident records a project failure case.
func (s *SQLiteStore) AddIncident(incident *domain.Incident) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resolvedInt := 0
	if incident.Resolved {
		resolvedInt = 1
	}

	query := `INSERT OR REPLACE INTO incidents (id, message, severity, resolved, detected_at, resolved_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, incident.ID, incident.Message, incident.Severity, resolvedInt, incident.DetectedAt.Format(time.RFC3339), incident.ResolvedAt.Format(time.RFC3339))
	return err
}

// GetIncidents returns historical failure incidents.
func (s *SQLiteStore) GetIncidents() ([]domain.Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, message, severity, resolved, detected_at, resolved_at FROM incidents`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var id, msg, sev, detStr, resStr string
		var resInt int
		if err := rows.Scan(&id, &msg, &sev, &resInt, &detStr, &resStr); err != nil {
			return nil, err
		}

		det, _ := time.Parse(time.RFC3339, detStr)
		res, _ := time.Parse(time.RFC3339, resStr)

		incidents = append(incidents, domain.Incident{
			ID:         id,
			Message:    msg,
			Severity:   sev,
			Resolved:   resInt == 1,
			DetectedAt: det,
			ResolvedAt: res,
		})
	}
	return incidents, nil
}

// ResolveIncident markers an incident as resolved.
func (s *SQLiteStore) ResolveIncident(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE incidents SET resolved = 1, resolved_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now().Format(time.RFC3339), id)
	return err
}

// AddDeployment saves deployment metadata logs.
func (s *SQLiteStore) AddDeployment(deployment *domain.Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT OR REPLACE INTO deployments (id, env, status, version, started_at, ended_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, deployment.ID, deployment.Env, deployment.Status, deployment.Version, deployment.StartedAt.Format(time.RFC3339), deployment.EndedAt.Format(time.RFC3339))
	return err
}

// GetDeployments retrieves historical deployment summaries.
func (s *SQLiteStore) GetDeployments() ([]domain.Deployment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, env, status, version, started_at, ended_at FROM deployments`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []domain.Deployment
	for rows.Next() {
		var id, env, status, ver, startStr, endStr string
		if err := rows.Scan(&id, &env, &status, &ver, &startStr, &endStr); err != nil {
			return nil, err
		}

		start, _ := time.Parse(time.RFC3339, startStr)
		end, _ := time.Parse(time.RFC3339, endStr)

		deployments = append(deployments, domain.Deployment{
			ID:        id,
			Env:       env,
			Status:    status,
			Version:   ver,
			StartedAt: start,
			EndedAt:   end,
		})
	}
	return deployments, nil
}

// AddRecommendation logs recommendations.
func (s *SQLiteStore) AddRecommendation(rec *domain.Recommendation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fixableInt := 0
	if rec.Fixable {
		fixableInt = 1
	}

	query := `INSERT OR REPLACE INTO recommendations (id, category, message, fixable, rationale) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, rec.ID, rec.Category, rec.Message, fixableInt, rec.Rationale)
	return err
}

// GetRecommendations returns recommendations from database.
func (s *SQLiteStore) GetRecommendations() ([]domain.Recommendation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `SELECT id, category, message, fixable, rationale FROM recommendations`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []domain.Recommendation
	for rows.Next() {
		var id, cat, msg, rat string
		var fixInt int
		if err := rows.Scan(&id, &cat, &msg, &fixInt, &rat); err != nil {
			return nil, err
		}

		recs = append(recs, domain.Recommendation{
			ID:        id,
			Category:  cat,
			Message:   msg,
			Fixable:   fixInt == 1,
			Rationale: rat,
		})
	}
	return recs, nil
}

