package debug_test

import (
	"context"
	"os"
	"testing"
	"time"

	"daemon/core/debug/integration"
	"daemon/core/events"
	"daemon/core/evolution"
	"daemon/core/graph"
	"daemon/core/instruments"
	"daemon/core/twin"
)

func TestIntegration_EventsAdapter(t *testing.T) {
	bus := events.NewMemoryEventBus()
	bus.Publish(events.Event{
		ID:        "evt-1",
		Type:      "service_alert",
		Source:    "orders-api",
		EntityID:  "orders",
		Timestamp: time.Now(),
	})

	adapter := integration.NewEventsAdapter(bus)
	evs, err := adapter.GatherEventsEvidence()
	if err != nil {
		t.Fatalf("failed to gather events evidence: %v", err)
	}

	if len(evs) == 0 {
		t.Fatal("expected gathered event evidence, got 0")
	}

	if evs[0].EntityID != "orders" || evs[0].Type != instruments.EvidenceEvent {
		t.Errorf("unexpected gathered evidence properties: %+v", evs[0])
	}
}

func TestIntegration_HistoryAdapter(t *testing.T) {
	ledger, err := evolution.NewFixLedger(":memory:")
	if err != nil {
		t.Fatalf("failed to create fix ledger: %v", err)
	}

	err = ledger.RecordFix(evolution.FixLedgerEntry{
		ActionID:           "act-123",
		PatternID:          "pat-999",
		ErrorSignature:     "connection pool exhausted",
		RootCause:          "unclosed response body",
		FixSummary:         "add defer body close",
		VerificationResult: "VERIFIED",
		Environment:        "test",
		Timestamp:          time.Now(),
		EvidenceIDs:        []string{"ev-1"},
	})
	if err != nil {
		t.Fatalf("failed to record fix: %v", err)
	}

	adapter := integration.NewHistoryAdapter(ledger)
	evs, err := adapter.GatherHistoryEvidence("connection pool exhausted")
	if err != nil {
		t.Fatalf("failed to gather history evidence: %v", err)
	}

	if len(evs) == 0 {
		t.Fatal("expected gathered history evidence, got 0")
	}

	if evs[0].EntityID != "act-123" || evs[0].Type != instruments.EvidenceHistory {
		t.Errorf("unexpected gathered history evidence: %+v", evs[0])
	}
}

func TestIntegration_GraphAdapter(t *testing.T) {
	dbPath := "test_graph_adapter.db"
	defer os.Remove(dbPath)

	store, err := graph.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite store: %v", err)
	}
	defer store.Close()

	_ = store.AddNode("Repository", "repo-1", "DaemonCore", map[string]string{})
	_ = store.AddEdge("Repository", "repo-1", "Repository", "repo-2", "depends_on")

	kg := graph.NewKnowledgeGraph(store)
	adapter := integration.NewGraphAdapter(kg)

	evs, err := adapter.GatherGraphEvidence("repo-2")
	if err != nil {
		t.Fatalf("failed to gather graph evidence: %v", err)
	}

	if len(evs) == 0 {
		t.Fatal("expected dependency relation evidence, got 0")
	}
}

func TestIntegration_TwinAdapter(t *testing.T) {
	dbPath := "test_twin_adapter.db"
	defer os.Remove(dbPath)

	store, err := graph.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create sqlite store: %v", err)
	}
	defer store.Close()

	_ = store.AddNode("service", "service-123", "CheckoutService", map[string]string{
		"port":   "8080",
		"status": "active",
	})

	model := twin.NewTwinModel(store)
	adapter := integration.NewTwinAdapter(model)

	evs, err := adapter.GatherTwinEvidence(context.Background(), "CheckoutService")
	if err != nil {
		t.Fatalf("failed to gather twin evidence: %v", err)
	}

	if len(evs) == 0 {
		t.Fatal("expected twin match evidence, got 0")
	}
}
