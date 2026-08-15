package instruments

import (
	"context"
	"fmt"
	"sync"
)

type InstrumentRegistry struct {
	mu          sync.RWMutex
	instruments map[string]EngineeringInstrument
}

func NewInstrumentRegistry() *InstrumentRegistry {
	return &InstrumentRegistry{
		instruments: make(map[string]EngineeringInstrument),
	}
}

func (ir *InstrumentRegistry) Register(inst EngineeringInstrument) error {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	id := inst.Identity().ID
	if id == "" {
		return fmt.Errorf("instrument identity ID is required")
	}

	ir.instruments[id] = inst
	return nil
}

func (ir *InstrumentRegistry) Unregister(id string) error {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	if _, exists := ir.instruments[id]; !exists {
		return fmt.Errorf("instrument %s not registered", id)
	}

	delete(ir.instruments, id)
	return nil
}

func (ir *InstrumentRegistry) Get(id string) (EngineeringInstrument, bool) {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	inst, exists := ir.instruments[id]
	return inst, exists
}

func (ir *InstrumentRegistry) List() []EngineeringInstrument {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	list := make([]EngineeringInstrument, 0, len(ir.instruments))
	for _, inst := range ir.instruments {
		list = append(list, inst)
	}
	return list
}

// FindByID returns the registered instrument with the given ID, or nil if not found.
func (ir *InstrumentRegistry) FindByID(id string) EngineeringInstrument {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	return ir.instruments[id]
}

func (ir *InstrumentRegistry) FindByCapability(cap Capability) []EngineeringInstrument {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	var matched []EngineeringInstrument
	for _, inst := range ir.instruments {
		for _, c := range inst.Capabilities() {
			if c == cap {
				matched = append(matched, inst)
				break
			}
		}
	}
	return matched
}

func (ir *InstrumentRegistry) FindCompatible(ctx context.Context, env Environment) []EngineeringInstrument {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	var matched []EngineeringInstrument
	for _, inst := range ir.instruments {
		res := inst.Detect(ctx, env)
		if res.Compatible {
			matched = append(matched, inst)
		}
	}
	return matched
}
