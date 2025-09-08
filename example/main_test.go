// package main

// import (
// 	"context"
// 	"fmt"
// 	"testing"
// 	"time"
// )

// // A separate set of tasks for testing to ensure no side effects
// var testTasks = map[string]TaskFn{
// 	"StartTask": func(ctx context.Context, sc *StateContext) error {
// 		// This task no longer sets the initial data if it's already present from the test case
// 		if _, ok := sc.Data["items_to_process"]; !ok {
// 			sc.Data["items_to_process"] = []any{1, 2, 3}
// 			sc.Data["choice_value"] = "go"
// 			sc.Data["attempts"] = float64(0)
// 		}
// 		return nil
// 	},
// 	"MapTask": func(ctx context.Context, sc *StateContext) error {
// 		item, ok := sc.Data["item"].(float64)
// 		if !ok {
// 			if itemInt, ok := sc.Data["item"].(int); ok {
// 				item = float64(itemInt)
// 			} else {
// 				return fmt.Errorf("invalid item type: %T", sc.Data["item"])
// 			}
// 		}
// 		sc.Data["processed_item"] = item * 10
// 		return nil
// 	},
// 	"BranchATask": func(ctx context.Context, sc *StateContext) error {
// 		sc.Data["branch_a_status"] = "finished"
// 		return nil
// 	},
// 	"BranchBTask": func(ctx context.Context, sc *StateContext) error {
// 		sc.Data["branch_b_status"] = "finished"
// 		return nil
// 	},
// 	"TestRetryCatch": func(ctx context.Context, sc *StateContext) error {
// 		currentAttempts, _ := sc.Data["attempts"].(float64)
// 		sc.Data["attempts"] = currentAttempts + 1
// 		return ErrAPIBadGateway
// 	},
// 	"DefaultTask": func(ctx context.Context, sc *StateContext) error {
// 		sc.Data["default_path"] = "taken"
// 		return nil
// 	},
// 	"SucceedingTask": func(ctx context.Context, sc *StateContext) error {
// 		sc.Data["succeeding_path"] = "taken"
// 		return nil
// 	},
// }

// // buildTestStateMachine is now a helper function to create a new, clean instance for each test.
// func buildTestStateMachine() *StateMachine {
// 	mapBranchBuilder := NewStateMachineBuilder().
// 		StartAt("MapTask").
// 		AddTask("MapTask", testTasks["MapTask"], "MapEnd", true).
// 		AddEnd("MapEnd")
// 	mapBranch, _ := mapBranchBuilder.Build()

// 	branchABuilder := NewStateMachineBuilder().
// 		StartAt("BranchATask").
// 		AddTask("BranchATask", testTasks["BranchATask"], "BranchAEnd", true).
// 		AddEnd("BranchAEnd")
// 	branchBBuilder := NewStateMachineBuilder().
// 		StartAt("BranchBTask").
// 		AddTask("BranchBTask", testTasks["BranchBTask"], "BranchBEnd", true).
// 		AddEnd("BranchBEnd")

// 	builder := NewStateMachineBuilder().
// 		StartAt("StartTask").
// 		AddTask("StartTask", testTasks["StartTask"], "TestPassState").
// 		AddPass("TestPassState", "TestMapState", func(sc *StateContext) {
// 			sc.Data["message"] = "Data has been passed successfully."
// 		}).
// 		AddMap("TestMapState", "items_to_process", "map_output", mapBranch, "TestChoiceState").
// 		AddChoice("TestChoiceState", []ChoiceRule{
// 			{
// 				Condition: map[string]any{"StringEquals": "go"},
// 				Next:      "TestParallelState",
// 			},
// 			{
// 				Condition: map[string]any{"NumericEquals": float64(10)},
// 				Next:      "SucceedingTask",
// 			},
// 		}, "DefaultTask").
// 		AddParallel("TestParallelState", []*StateMachine{branchABuilder.BuildOrDie(), branchBBuilder.BuildOrDie()}, "WaitState").
// 		AddWait("WaitState", 1, "TestRetryCatch").
// 		AddTask("TestRetryCatch", testTasks["TestRetryCatch"], "FinalEnd",
// 			RetryRule{ErrorName: "API_BAD_GATEWAY", Interval: 50 * time.Millisecond, MaxAttempts: 3},
// 			CatchRule{ErrorName: "API_BAD_GATEWAY", NextState: "FailState"}).
// 		AddTask("DefaultTask", testTasks["DefaultTask"], "FinalEnd", true).
// 		AddTask("SucceedingTask", testTasks["SucceedingTask"], "FinalEnd", true).
// 		AddFail("FailState").
// 		AddEnd("FinalEnd")

// 	sm, _ := builder.Build()
// 	return sm
// }

// // TestProgrammaticStateMachine now uses sub-tests with fresh state machines.
// func TestProgrammaticStateMachine(t *testing.T) {
// 	t.Run("Full workflow execution with expected failure (StringEquals path)", func(t *testing.T) {
// 		sm := buildTestStateMachine()
// 		err := sm.Run(context.Background(), map[string]any{})
// 		if err == nil {
// 			t.Fatal("Expected workflow to fail, but it succeeded")
// 		}

// 		if sm.context.Data["attempts"] != float64(4) {
// 			t.Errorf("Expected 4 attempts, got %v", sm.context.Data["attempts"])
// 		}
// 	})

// 	t.Run("ChoiceState takes the default path", func(t *testing.T) {
// 		sm := buildTestStateMachine()
// 		initialData := map[string]any{
// 			"items_to_process": []any{1},
// 			"choice_value":     "other",
// 		}

// 		err := sm.Run(context.Background(), initialData)
// 		if err != nil {
// 			t.Fatalf("Unexpected error: %v", err)
// 		}

// 		if _, ok := sm.context.Data["default_path"]; !ok {
// 			t.Errorf("Expected 'default_path' to be in context, but it wasn't")
// 		}
// 	})

// 	t.Run("ChoiceState takes the NumericEquals path", func(t *testing.T) {
// 		sm := buildTestStateMachine()
// 		initialData := map[string]any{
// 			"items_to_process": []any{1},
// 			"choice_value":     float64(10),
// 		}

// 		err := sm.Run(context.Background(), initialData)
// 		if err != nil {
// 			t.Fatalf("Unexpected error: %v", err)
// 		}
// 		if _, ok := sm.context.Data["succeeding_path"]; !ok {
// 			t.Errorf("Expected 'succeeding_path' to be in context, but it wasn't")
// 		}
// 	})
// }

// // TestJSONStateMachine remains unchanged as it already creates a new instance.
// func TestJSONStateMachine(t *testing.T) {
// 	t.Run("Full workflow execution from JSON with expected failure", func(t *testing.T) {
// 		sm, err := ParseStateMachine("workflow.json", testTasks)
// 		if err != nil {
// 			t.Fatalf("Failed to parse JSON file: %v", err)
// 		}

// 		err = sm.Run(context.Background(), map[string]any{})
// 		if err == nil {
// 			t.Fatal("Expected workflow to fail, but it succeeded")
// 		}

// 		if sm.context.Data["attempts"] != float64(4) {
// 			t.Errorf("Expected 4 attempts, got %v", sm.context.Data["attempts"])
// 		}
// 	})
// }

package example

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basillica/state-machine/statemachine"
)

// A separate set of tasks for testing to ensure no side effects
var testTasks = map[string]statemachine.TaskFn{
	"StartTask": func(ctx context.Context, sc *statemachine.StateContext) error {
		// This task no longer sets the initial data if it's already present from the test case
		if _, ok := sc.Data["items_to_process"]; !ok {
			sc.Data["items_to_process"] = []any{1, 2, 3}
			sc.Data["choice_value"] = "go"
			sc.Data["attempts"] = float64(0)
		}
		return nil
	},
	"MapTask": func(ctx context.Context, sc *statemachine.StateContext) error {
		item, ok := sc.Data["item"].(float64)
		if !ok {
			if itemInt, ok := sc.Data["item"].(int); ok {
				item = float64(itemInt)
			} else {
				return fmt.Errorf("invalid item type: %T", sc.Data["item"])
			}
		}
		sc.Data["processed_item"] = item * 10
		return nil
	},
	"BranchATask": func(ctx context.Context, sc *statemachine.StateContext) error {
		sc.Data["branch_a_status"] = "finished"
		return nil
	},
	"BranchBTask": func(ctx context.Context, sc *statemachine.StateContext) error {
		sc.Data["branch_b_status"] = "finished"
		return nil
	},
	"TestRetryCatch": func(ctx context.Context, sc *statemachine.StateContext) error {
		currentAttempts, _ := sc.Data["attempts"].(float64)
		sc.Data["attempts"] = currentAttempts + 1
		return statemachine.ErrAPIBadGateway
	},
	"DefaultTask": func(ctx context.Context, sc *statemachine.StateContext) error {
		sc.Data["default_path"] = "taken"
		return nil
	},
	"SucceedingTask": func(ctx context.Context, sc *statemachine.StateContext) error {
		sc.Data["succeeding_path"] = "taken"
		return nil
	},
	"TestTimeoutTask": func(ctx context.Context, sc *statemachine.StateContext) error {
		select {
		case <-time.After(3 * time.Second): // This is a slow task
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	},
}

// buildTestStateMachine is now a helper function to create a new, clean instance for each test.
func buildTestStateMachine() *statemachine.StateMachine {
	mapBranchBuilder := statemachine.NewStateMachineBuilder().
		StartAt("MapTask").
		AddTask("MapTask", testTasks["MapTask"], "MapEnd", true).
		AddEnd("MapEnd")
	mapBranch, _ := mapBranchBuilder.Build()

	branchABuilder := statemachine.NewStateMachineBuilder().
		StartAt("BranchATask").
		AddTask("BranchATask", testTasks["BranchATask"], "BranchAEnd", true).
		AddEnd("BranchAEnd")
	branchBBuilder := statemachine.NewStateMachineBuilder().
		StartAt("BranchBTask").
		AddTask("BranchBTask", testTasks["BranchBTask"], "BranchBEnd", true).
		AddEnd("BranchBEnd")

	builder := statemachine.NewStateMachineBuilder().
		StartAt("StartTask").
		AddTask("StartTask", testTasks["StartTask"], "TestPassState").
		AddPass("TestPassState", "TestMapState", func(sc *statemachine.StateContext) {
			sc.Data["message"] = "Data has been passed successfully."
		}).
		AddMap("TestMapState", "items_to_process", "map_output", mapBranch, "TestChoiceState").
		AddChoice("TestChoiceState", []statemachine.ChoiceRule{
			{
				Condition: map[string]any{"StringEquals": "go"},
				Next:      "TestParallelState",
			},
			{
				Condition: map[string]any{"NumericEquals": float64(10)},
				Next:      "SucceedingTask",
			},
		}, "DefaultTask").
		AddParallel("TestParallelState", []*statemachine.StateMachine{branchABuilder.BuildOrDie(), branchBBuilder.BuildOrDie()}, "WaitState").
		AddWait("WaitState", 1, "TestRetryCatch").
		AddTask("TestRetryCatch", testTasks["TestRetryCatch"], "FinalEnd",
			statemachine.RetryRule{ErrorName: "API_BAD_GATEWAY", Interval: 50 * time.Millisecond, MaxAttempts: 3},
			statemachine.CatchRule{ErrorName: "API_BAD_GATEWAY", NextState: "FailState"}).
		AddTask("DefaultTask", testTasks["DefaultTask"], "FinalEnd", true).
		AddTask("SucceedingTask", testTasks["SucceedingTask"], "FinalEnd", true).
		AddFail("FailState").
		AddEnd("FinalEnd")

	sm, _ := builder.Build()
	return sm
}

func TestProgrammaticStateMachine(t *testing.T) {
	t.Run("Full workflow execution with expected failure (StringEquals path)", func(t *testing.T) {
		sm := buildTestStateMachine()
		err := sm.Run(context.Background(), map[string]any{})
		if err == nil {
			t.Fatal("Expected workflow to fail, but it succeeded")
		}

		if sm.Context.Data["attempts"] != float64(4) {
			t.Errorf("Expected 4 attempts, got %v", sm.Context.Data["attempts"])
		}
	})

	t.Run("ChoiceState takes the default path", func(t *testing.T) {
		sm := buildTestStateMachine()
		initialData := map[string]any{
			"items_to_process": []any{1},
			"choice_value":     "other",
		}

		err := sm.Run(context.Background(), initialData)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if _, ok := sm.Context.Data["default_path"]; !ok {
			t.Errorf("Expected 'default_path' to be in context, but it wasn't")
		}
	})

	t.Run("ChoiceState takes the NumericEquals path", func(t *testing.T) {
		sm := buildTestStateMachine()
		initialData := map[string]any{
			"items_to_process": []any{1},
			"choice_value":     float64(10),
		}

		err := sm.Run(context.Background(), initialData)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if _, ok := sm.Context.Data["succeeding_path"]; !ok {
			t.Errorf("Expected 'succeeding_path' to be in context, but it wasn't")
		}
	})

	t.Run("TaskState times out correctly", func(t *testing.T) {
		sm := statemachine.NewStateMachineBuilder().
			StartAt("TestTimeoutTask").
			AddTask("TestTimeoutTask", testTasks["TestTimeoutTask"], "End",
				1, // Set timeout to 1 second
				statemachine.CatchRule{ErrorName: "TIMEOUT", NextState: "FailState"}).
			AddFail("FailState").
			AddEnd("End").
			BuildOrDie()

		err := sm.Run(context.Background(), map[string]any{})

		if err == nil {
			t.Fatal("Expected workflow to fail due to timeout, but it succeeded.")
		}

		expectedError := "state 'FailState' failed: failure in state FailState"
		if err.Error() != expectedError {
			t.Errorf("Expected error: %q, but got %q", expectedError, err.Error())
		}
	})
}

func TestJSONStateMachine(t *testing.T) {
	t.Run("Full workflow execution from JSON with expected failure", func(t *testing.T) {
		sm, err := statemachine.ParseStateMachine("workflow.json", testTasks)
		if err != nil {
			t.Fatalf("Failed to parse JSON file: %v", err)
		}

		err = sm.Run(context.Background(), map[string]any{})
		if err == nil {
			t.Fatal("Expected workflow to fail, but it succeeded")
		}

		if sm.Context.Data["attempts"] != float64(4) {
			t.Errorf("Expected 4 attempts, got %v", sm.Context.Data["attempts"])
		}
	})
}
