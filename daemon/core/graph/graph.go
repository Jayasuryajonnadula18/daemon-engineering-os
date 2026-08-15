package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"daemon/core/domain"
)

type GraphExportPayload struct {
	SchemaVersion string              `json:"schema_version"`
	Nodes         []domain.Module     `json:"nodes"`
	Edges         []domain.EdgeRecord `json:"edges"`
}

type KnowledgeGraph struct {
	store *SQLiteStore
}

func NewKnowledgeGraph(store *SQLiteStore) *KnowledgeGraph {
	return &KnowledgeGraph{store: store}
}

// FindDownstreamImpact traverses directed edges to find all entities affected by changes to a given target entity.
func (kg *KnowledgeGraph) FindDownstreamImpact(targetID string) ([]string, error) {
	if kg.store == nil {
		return []string{}, nil
	}

	edges, err := kg.store.GetEdges()
	if err != nil {
		return nil, err
	}

	impactSet := make(map[string]bool)
	queue := []string{strings.ToLower(targetID)}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true

		for _, e := range edges {
			if strings.ToLower(e.FromID) == curr {
				toIDLower := strings.ToLower(e.ToID)
				if !visited[toIDLower] {
					impactSet[e.ToID] = true
					queue = append(queue, toIDLower)
				}
			}
		}
	}

	impact := make([]string, 0, len(impactSet))
	for id := range impactSet {
		impact = append(impact, id)
	}
	return impact, nil
}

// FindUpstreamDependencies traverses directed edges in reverse to find upstream dependencies required by an entity.
func (kg *KnowledgeGraph) FindUpstreamDependencies(targetID string) ([]string, error) {
	if kg.store == nil {
		return []string{}, nil
	}

	edges, err := kg.store.GetEdges()
	if err != nil {
		return nil, err
	}

	depSet := make(map[string]bool)
	queue := []string{strings.ToLower(targetID)}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true

		for _, e := range edges {
			if strings.ToLower(e.ToID) == curr {
				fromIDLower := strings.ToLower(e.FromID)
				if !visited[fromIDLower] {
					depSet[e.FromID] = true
					queue = append(queue, fromIDLower)
				}
			}
		}
	}

	deps := make([]string, 0, len(depSet))
	for id := range depSet {
		deps = append(deps, id)
	}
	return deps, nil
}

// ExportJSON writes schema-versioned snapshots to .daemon/graph.json and .daemon/context.json reproducibly.
func (kg *KnowledgeGraph) ExportJSON(projectRoot string) error {
	if kg.store == nil {
		return nil
	}

	daemonDir := filepath.Join(projectRoot, ".daemon")
	if err := os.MkdirAll(daemonDir, 0755); err != nil {
		return err
	}

	nodes, err := kg.store.GetAllNodes()
	if err != nil {
		nodes = []domain.Module{}
	}
	edges, err := kg.store.GetEdges()
	if err != nil {
		edges = []domain.EdgeRecord{}
	}

	payload := GraphExportPayload{
		SchemaVersion: "1.0.0",
		Nodes:         nodes,
		Edges:         edges,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph json: %w", err)
	}

	graphPath := filepath.Join(daemonDir, "graph.json")
	if err := os.WriteFile(graphPath, data, 0644); err != nil {
		return fmt.Errorf("write graph.json: %w", err)
	}

	contextPath := filepath.Join(daemonDir, "context.json")
	if err := os.WriteFile(contextPath, data, 0644); err != nil {
		return fmt.Errorf("write context.json: %w", err)
	}

	return nil
}
