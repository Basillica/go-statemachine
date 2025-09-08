package statemachine

import (
	"context"
	"fmt"
	"sync"
)

// MapState iterates over an array and executes a sub-workflow for each item.
type MapState struct {
	name   string
	input  string
	result string
	next   string
	branch *StateMachine
}

func (s *MapState) GetName() string {
	return s.name
}

func (s *MapState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("Executing MapState: %s\n", s.name)

	inputArray, ok := sc.Data[s.input].([]any)
	if !ok {
		return nil, fmt.Errorf("input '%s' is not an array", s.input)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(inputArray))
	mapOutput := make([]any, len(inputArray))

	for i, item := range inputArray {
		wg.Add(1)
		go func(itemData any, index int) {
			defer wg.Done()

			branchCtx := &StateContext{Data: map[string]any{"item": itemData}}

			branchCopy := *s.branch
			branchCopy.currentState = branchCopy.states[branchCopy.startAt]

			err := branchCopy.Run(ctx, branchCtx.Data)
			if err != nil {
				errChan <- fmt.Errorf("map iteration %d failed: %w", index, err)
				return
			}
			mapOutput[index] = branchCopy.Context.Data
		}(item, i)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	sc.Data[s.result] = mapOutput

	fmt.Println("Map state finished all iterations.")
	return machine.GetState(s.next), nil
}
