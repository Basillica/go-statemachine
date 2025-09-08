package statemachine

import (
	"context"
	"fmt"
	"time"
)

// -----------------------------------------------------------------------------
// Core State Machine Components
// -----------------------------------------------------------------------------

// The shared context that will be passed between states.
type StateContext struct {
	Data map[string]any
}

// State defines the interface for each step in the state machine.
type State interface {
	Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error)
	GetName() string
}

// StateMachine manages the execution flow and holds all states.
type StateMachine struct {
	states       map[string]State
	currentState State
	Context      *StateContext
	startAt      string
}

// GetState retrieves a state by its name.
func (sm *StateMachine) GetState(name string) State {
	return sm.states[name]
}

// Run executes the state machine.
func (sm *StateMachine) Run(ctx context.Context, initialData map[string]any) error {
	sm.Context = &StateContext{Data: initialData}
	sm.currentState = sm.states[sm.startAt]

	for sm.currentState != nil {
		nextState, err := sm.currentState.Execute(ctx, sm.Context, sm)
		if err != nil {
			return fmt.Errorf("state '%s' failed: %w", sm.currentState.GetName(), err)
		}
		sm.currentState = nextState
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
