package twin

import (
	"context"
	"fmt"
	"strings"

	"daemon/core/storage"
)

// TwinModel is the central system representation backing search and diagnostics.
type TwinModel struct {
	graphStore storage.GraphStore
}

// NewTwinModel instantiates a new TwinModel.
func NewTwinModel(gs storage.GraphStore) *TwinModel {
	return &TwinModel{graphStore: gs}
}

// SearchResult holds search outcome context.
type SearchResult struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Context string `json:"context"`
}

// Search matches query strings across services, dependencies, and APIs inside the twin model.
func (t *TwinModel) Search(ctx context.Context, query string) ([]SearchResult, error) {
	query = strings.ToLower(query)
	var results []SearchResult

	services, err := t.graphStore.GetServices()
	if err == nil {
		for _, s := range services {
			if strings.Contains(strings.ToLower(s.Name), query) || strings.Contains(strings.ToLower(s.ID), query) {
				results = append(results, SearchResult{
					Type:    "service",
					ID:      s.ID,
					Name:    s.Name,
					Context: fmt.Sprintf("Service Port: %d, Status: %s", s.Port, s.Status),
				})
			}
		}
	}

	deps, err := t.graphStore.GetDependencies()
	if err == nil {
		for _, d := range deps {
			if strings.Contains(strings.ToLower(d.Name), query) || strings.Contains(strings.ToLower(d.ID), query) {
				results = append(results, SearchResult{
					Type:    "dependency",
					ID:      d.ID,
					Name:    d.Name,
					Context: fmt.Sprintf("Dependency Version: %s, Outdated: %t", d.Version, d.IsOutdated),
				})
			}
		}
	}

	apis, err := t.graphStore.GetAPIs()
	if err == nil {
		for _, a := range apis {
			if strings.Contains(strings.ToLower(a.Path), query) || strings.Contains(strings.ToLower(a.ID), query) {
				results = append(results, SearchResult{
					Type:    "api",
					ID:      a.ID,
					Name:    a.Path,
					Context: fmt.Sprintf("API endpoint exposed by: %s (Method: %s)", a.Service, a.Method),
				})
			}
		}
	}

	// Query custom resource type nodes
	resourceTypes := []string{"Repository", "Container", "Deployment", "Tunnel"}
	for _, rt := range resourceTypes {
		nodes, err := t.graphStore.GetNodes(rt)
		if err == nil {
			for _, n := range nodes {
				if strings.Contains(strings.ToLower(n.Name), query) || strings.Contains(strings.ToLower(n.ID), query) {
					results = append(results, SearchResult{
						Type:    strings.ToLower(rt),
						ID:      n.ID,
						Name:    n.Name,
						Context: fmt.Sprintf("Active integration resource of type %s.", rt),
					})
				}
			}
		}
	}

	return results, nil
}

// UpdateIncremental applies a real-time event mutation to the twin model in SQLite.
func (t *TwinModel) UpdateIncremental(ctx context.Context, entityType string, id string, name string, attrs map[string]string) error {
	if t.graphStore == nil {
		return nil
	}
	return t.graphStore.AddNode(entityType, id, name, attrs)
}

