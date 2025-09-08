package statemachine

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TaskFn defines the signature for a function to be executed by a TaskState.
type TaskFn func(ctx context.Context, sc *StateContext) error

// RetryRule defines how to handle specific errors for retries.
type RetryRule struct {
	ErrorName   string
	Interval    time.Duration
	MaxAttempts int
}

// CatchRule defines a transition for a caught error.
type CatchRule struct {
	ErrorName string
	NextState string
}

// TaskState is a concrete state that runs a given function with retry and catch logic.
type TaskState struct {
	name           string
	execute        TaskFn
	next           string
	retries        []RetryRule
	catches        []CatchRule
	end            bool
	TimeoutSeconds int
}

func (s *TaskState) GetName() string {
	return s.name
}

func (s *TaskState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("Executing TaskState: %s\n", s.name)

	var err error
	for i := 0; ; i++ {
		taskCtx := ctx
		if s.TimeoutSeconds > 0 {
			var cancel context.CancelFunc
			taskCtx, cancel = context.WithTimeout(ctx, time.Duration(s.TimeoutSeconds)*time.Second)
			defer cancel()
		}

		// Channel to signal task completion
		done := make(chan error, 1)
		go func() {
			done <- s.execute(taskCtx, sc)
		}()

		select {
		case taskErr := <-done:
			err = taskErr
		case <-taskCtx.Done():
			err = ErrTimeout
		}

		if err == nil {
			if s.end {
				return &EndState{name: "End"}, nil
			}
			return machine.GetState(s.next), nil
		}

		fmt.Printf("Task '%s' failed, attempt %d. Error: %v\n", s.name, i+1, err)

		var matchedRetryRule *RetryRule
		for _, rule := range s.retries {
			var customErr *CustomError
			if errors.As(err, &customErr) && customErr.Name == rule.ErrorName {
				matchedRetryRule = &rule
				break
			}
		}

		if matchedRetryRule == nil || i >= matchedRetryRule.MaxAttempts {
			break
		}
		time.Sleep(matchedRetryRule.Interval)
	}

	for _, catchRule := range s.catches {
		var customErr *CustomError
		if errors.As(err, &customErr) && customErr.Name == catchRule.ErrorName {
			fmt.Printf("Error '%s' caught. Transitioning to state: %s\n", catchRule.ErrorName, catchRule.NextState)
			return machine.GetState(catchRule.NextState), nil
		}
	}

	return nil, err
}
