package example

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/basillica/state-machine/statemachine"
)

var (
	useJSON = flag.Bool("json", false, "Use JSON definition instead of programmatic builder")
)

func ExampleMain() {
	flag.Parse()
	tasks := map[string]statemachine.TaskFn{
		"StartTask": func(ctx context.Context, sc *statemachine.StateContext) error {
			// Check if data is already set (for testing purposes)
			if _, ok := sc.Data["items_to_process"]; !ok {
				fmt.Println("Start: Setting up initial data.")
				sc.Data["attempts"] = 0
				sc.Data["items_to_process"] = []any{1, 2, 3}
				sc.Data["choice_value"] = "go"
				sc.Data["numeric_value"] = float64(10)
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
			fmt.Printf("Map Task: Processing item %v...\n", item)
			sc.Data["processed_item"] = item * 10
			return nil
		},
		"TestRetryCatch": func(ctx context.Context, sc *statemachine.StateContext) error {
			fmt.Println("Attempting a task that will fail...")
			currentAttempts := 0
			if val, ok := sc.Data["attempts"].(float64); ok {
				currentAttempts = int(val)
			}
			sc.Data["attempts"] = float64(currentAttempts + 1)
			return statemachine.ErrAPIBadGateway
		},
		"BranchATask": func(ctx context.Context, sc *statemachine.StateContext) error {
			fmt.Println("Executing Branch A Task.")
			return nil
		},
		"BranchBTask": func(ctx context.Context, sc *statemachine.StateContext) error {
			fmt.Println("Executing Branch B Task.")
			return nil
		},
		"DefaultTask": func(ctx context.Context, sc *statemachine.StateContext) error {
			fmt.Println("Executing Default Task.")
			return nil
		},
		"SucceedingTask": func(ctx context.Context, sc *statemachine.StateContext) error {
			return nil
		},
		"TestTimeoutTask": func(ctx context.Context, sc *statemachine.StateContext) error {
			fmt.Println("Attempting a task that will time out...")
			select {
			case <-time.After(3 * time.Second): // This task will take 3 seconds
				return nil
			case <-ctx.Done(): // If the context is canceled, it's because of a timeout
				return ctx.Err()
			}
		},
	}

	var sm *statemachine.StateMachine
	var err error

	if *useJSON {
		sm, err = statemachine.ParseStateMachine("workflow.json", tasks)
		if err != nil {
			fmt.Printf("Failed to parse state machine from JSON: %v\n", err)
			return
		}
		fmt.Println("--- Running state machine from JSON definition ---")
	} else {
		// Programmatic builder definition
		mapBranchBuilder := statemachine.NewStateMachineBuilder().
			StartAt("MapTask").
			AddTask("MapTask", tasks["MapTask"], "MapEnd", true).
			AddEnd("MapEnd")

		mapBranch, _ := mapBranchBuilder.Build()

		branchABuilder := statemachine.NewStateMachineBuilder().
			StartAt("BranchATask").
			AddTask("BranchATask", tasks["BranchATask"], "BranchAEnd", true).
			AddEnd("BranchAEnd")

		branchBBuilder := statemachine.NewStateMachineBuilder().
			StartAt("BranchBTask").
			AddTask("BranchBTask", tasks["BranchBTask"], "BranchBEnd", true).
			AddEnd("BranchBEnd")

		builder := statemachine.NewStateMachineBuilder().
			StartAt("StartTask").
			AddTask("StartTask", tasks["StartTask"], "TestPassState").
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
			AddTask("TestRetryCatch", tasks["TestRetryCatch"], "FinalEnd",
				statemachine.RetryRule{ErrorName: "API_BAD_GATEWAY", Interval: 50 * time.Millisecond, MaxAttempts: 3},
				statemachine.CatchRule{ErrorName: "API_BAD_GATEWAY", NextState: "FailState"}).
			AddTask("DefaultTask", tasks["DefaultTask"], "FinalEnd", true).
			AddTask("SucceedingTask", tasks["SucceedingTask"], "FinalEnd", true).
			AddFail("FailState").
			AddEnd("FinalEnd")

		sm, err = builder.Build()
		if err != nil {
			fmt.Printf("Failed to build state machine programmatically: %v\n", err)
			return
		}
		fmt.Println("--- Running state machine from programmatic builder ---")
	}

	if err := sm.Run(context.Background(), map[string]any{}); err != nil {
		fmt.Printf("Workflow failed: %v\n", err)
	}
	fmt.Printf("\nFinal context data: %+v\n", sm.Context.Data)
}
