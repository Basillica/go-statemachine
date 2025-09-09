package statemachine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ParseStateMachine reads a JSON file and builds a StateMachine.
func ParseStateMachine(filePath string, tasks map[string]TaskFn) (*StateMachine, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var def StateMachineDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON: %w", err)
	}

	states := make(map[string]State)
	for name, rawState := range def.States {
		var stateType StateType
		if err := json.Unmarshal(rawState, &stateType); err != nil {
			return nil, fmt.Errorf("could not determine state type for '%s': %w", name, err)
		}

		switch stateType.Type {
		case "Task":
			var taskDef TaskStateDefinition
			if err := json.Unmarshal(rawState, &taskDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal task state '%s': %w", name, err)
			}
			taskDef.Name = name
			task := &TaskState{name: taskDef.Name, next: taskDef.Next, end: taskDef.End, TimeoutSeconds: taskDef.TimeoutSeconds}
			if taskFn, ok := tasks[name]; ok {
				task.execute = taskFn
			}

			for _, rule := range taskDef.Retry {
				task.retries = append(task.retries, RetryRule{
					ErrorName:   rule.ErrorEquals[0],
					Interval:    time.Duration(rule.IntervalSeconds) * time.Second,
					MaxAttempts: rule.MaxAttempts,
				})
			}
			for _, rule := range taskDef.Catch {
				task.catches = append(task.catches, CatchRule{
					ErrorName: rule.ErrorEquals[0],
					NextState: rule.Next,
				})
			}
			states[name] = task
		case "Pass":
			var passDef PassStateDefinition
			if err := json.Unmarshal(rawState, &passDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal pass state '%s': %w", name, err)
			}
			passDef.Name = name
			states[name] = &PassState{name: passDef.Name, next: passDef.Next}
		case "Map":
			var mapDef MapStateDefinition
			if err := json.Unmarshal(rawState, &mapDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal map state '%s': %w", name, err)
			}
			mapDef.Name = name
			inputKey := strings.TrimPrefix(mapDef.InputPath, "$.")
			resultKey := strings.TrimPrefix(mapDef.ResultPath, "$.")

			subMachine, err := parseSubMachine(mapDef.Iterator, tasks)
			if err != nil {
				return nil, fmt.Errorf("could not parse Map iterator for state '%s': %w", name, err)
			}
			states[name] = &MapState{
				name:   mapDef.Name,
				input:  inputKey,
				result: resultKey,
				next:   mapDef.Next,
				branch: subMachine,
			}
		case "Choice":
			var choiceDef ChoiceStateDefinition
			if err := json.Unmarshal(rawState, &choiceDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal choice state '%s': %w", name, err)
			}
			choiceDef.Name = name
			var choices []ChoiceRule
			for _, rule := range choiceDef.Choices {
				choices = append(choices, ChoiceRule{Condition: rule.Condition, Next: rule.Next})
			}
			states[name] = &ChoiceState{name: choiceDef.Name, choices: choices, defaultState: choiceDef.Default}
		case "Wait":
			var waitDef WaitStateDefinition
			if err := json.Unmarshal(rawState, &waitDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal wait state '%s': %w", name, err)
			}
			waitDef.Name = name
			states[name] = &WaitState{name: waitDef.Name, seconds: waitDef.Seconds, next: waitDef.Next}
		case "Parallel":
			var parallelDef ParallelStateDefinition
			if err := json.Unmarshal(rawState, &parallelDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal parallel state '%s': %w", name, err)
			}
			parallelDef.Name = name
			var branches []*StateMachine
			for _, branchDef := range parallelDef.Branches {
				branch, err := parseSubMachine(branchDef, tasks)
				if err != nil {
					return nil, fmt.Errorf("could not parse Parallel branch: %w", err)
				}
				branches = append(branches, branch)
			}
			states[name] = &ParallelState{name: parallelDef.Name, branches: branches, next: parallelDef.Next}
		case "End":
			states[name] = &EndState{name: name}
		case "Fail":
			states[name] = &FailState{name: name}
		default:
			return nil, fmt.Errorf("unknown state type '%s' for state '%s'", stateType.Type, name)
		}
	}

	return &StateMachine{
		states:  states,
		startAt: def.StartAt,
	}, nil
}

// parseSubMachine is a helper function for nested state machines (like in Map or Parallel states).
func parseSubMachine(def StateMachineDefinition, tasks map[string]TaskFn) (*StateMachine, error) {
	states := make(map[string]State)
	for name, rawState := range def.States {
		var stateType StateType
		if err := json.Unmarshal(rawState, &stateType); err != nil {
			return nil, fmt.Errorf("could not determine sub-state type for '%s': %w", name, err)
		}

		switch stateType.Type {
		case "Task":
			var taskDef TaskStateDefinition
			if err := json.Unmarshal(rawState, &taskDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal sub-task state '%s': %w", name, err)
			}
			taskDef.Name = name
			task := &TaskState{name: taskDef.Name, next: taskDef.Next, end: taskDef.End, TimeoutSeconds: taskDef.TimeoutSeconds}
			if taskFn, ok := tasks[name]; ok {
				task.execute = taskFn
			}
			states[name] = task
		case "Choice":
			var choiceDef ChoiceStateDefinition
			if err := json.Unmarshal(rawState, &choiceDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal sub-choice state '%s': %w", name, err)
			}
			choiceDef.Name = name
			var choices []ChoiceRule
			for _, rule := range choiceDef.Choices {
				choices = append(choices, ChoiceRule{Condition: rule.Condition, Next: rule.Next})
			}
			states[name] = &ChoiceState{name: choiceDef.Name, choices: choices, defaultState: choiceDef.Default}
		case "Wait":
			var waitDef WaitStateDefinition
			if err := json.Unmarshal(rawState, &waitDef); err != nil {
				return nil, fmt.Errorf("could not unmarshal sub-wait state '%s': %w", name, err)
			}
			waitDef.Name = name
			states[name] = &WaitState{name: waitDef.Name, seconds: waitDef.Seconds, next: waitDef.Next}
		case "End":
			states[name] = &EndState{name: name}
		default:
			return nil, fmt.Errorf("unknown sub-state type '%s' for state '%s'", stateType.Type, name)
		}
	}

	return &StateMachine{
		states:  states,
		startAt: def.StartAt,
	}, nil
}
