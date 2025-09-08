package statemachine

import (
	"context"
	"fmt"
	"sync"
)

// ParallelState executes multiple independent branches concurrently.
type ParallelState struct {
	name     string
	branches []*StateMachine
	next     string
}

func (s *ParallelState) GetName() string {
	return s.name
}

func (s *ParallelState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("Executing ParallelState: %s\n", s.name)
	var wg sync.WaitGroup
	errChan := make(chan error, len(s.branches))
	branchOutputs := make([]any, len(s.branches))

	for i, branch := range s.branches {
		wg.Add(1)
		go func(branch *StateMachine, index int) {
			defer wg.Done()
			branchCopy := *branch
			branchCopy.currentState = branchCopy.states[branchCopy.startAt]
			err := branchCopy.Run(ctx, make(map[string]any))
			if err != nil {
				errChan <- fmt.Errorf("parallel branch %d failed: %w", index, err)
				return
			}
			branchOutputs[index] = branchCopy.Context.Data
		}(branch, i)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	sc.Data["parallel_output"] = branchOutputs
	fmt.Println("Parallel state finished all branches.")
	return machine.GetState(s.next), nil
}
