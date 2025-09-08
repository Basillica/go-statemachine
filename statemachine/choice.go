package statemachine

import (
	"context"
	"fmt"
	"strings"
)

// ChoiceRule defines a condition and the next state to transition to.
type ChoiceRule struct {
	Condition map[string]any
	Next      string
}

// ChoiceState implements conditional branching.
type ChoiceState struct {
	name         string
	choices      []ChoiceRule
	defaultState string
}

func (s *ChoiceState) GetName() string {
	return s.name
}

func (s *ChoiceState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("Executing ChoiceState: %s\n", s.name)
	for _, rule := range s.choices {
		if s.evaluateCondition(rule.Condition, sc) {
			fmt.Printf("Condition met. Transitioning to %s\n", rule.Next)
			return machine.GetState(rule.Next), nil
		}
	}
	fmt.Printf("No conditions met. Transitioning to default state %s\n", s.defaultState)
	return machine.GetState(s.defaultState), nil
}

func (s *ChoiceState) evaluateCondition(condition map[string]any, sc *StateContext) bool {
	var inputValue any

	// First, try to get the input value from the "InputPath" field, if it exists (JSON case)
	if path, okPath := condition["InputPath"].(string); okPath {
		// Use a new variable for the `ok` from the map lookup to avoid shadowing
		if val, okData := sc.Data[strings.TrimPrefix(path, "$.")]; okData {
			inputValue = val
		} else {
			// Key from InputPath not found, so the condition cannot be evaluated.
			return false
		}
	} else {
		// If InputPath is not found, assume it's the programmatic case
		// and the key to check is `choice_value` as defined in the main function.
		if val, okData := sc.Data["choice_value"]; okData {
			inputValue = val
		} else {
			// Hardcoded key not found, return false.
			return false
		}
	}

	// Now, evaluate the condition based on the comparison operator
	for key, val := range condition {
		switch key {
		case "StringEquals":
			expected, ok := val.(string)
			inputVal, ok2 := inputValue.(string)
			return ok && ok2 && inputVal == expected
		case "NumericEquals":
			expected, ok := val.(float64)
			inputVal, ok2 := inputValue.(float64)
			return ok && ok2 && inputVal == expected
		case "BooleanEquals":
			expected, ok := val.(bool)
			inputVal, ok2 := inputValue.(bool)
			return ok && ok2 && inputVal == expected
		}
	}
	return false
}
