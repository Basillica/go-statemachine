package statemachine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// The overall workflow definition from the JSON file.
type StateMachineDefinition struct {
	StartAt string                     `json:"StartAt"`
	States  map[string]json.RawMessage `json:"States"`
}

// StateType is used to unmarshal the state's type.
type StateType struct {
	Type string `json:"Type"`
}

// State definition structs for JSON parsing
type TaskStateDefinition struct {
	Name           string            `json:"-"`
	Type           string            `json:"Type"`
	Next           string            `json:"Next,omitempty"`
	End            bool              `json:"End,omitempty"`
	Retry          []RetryDefinition `json:"Retry,omitempty"`
	Catch          []CatchDefinition `json:"Catch,omitempty"`
	TimeoutSeconds int               `json:"TimeoutSeconds,omitempty"`
}

type RetryDefinition struct {
	ErrorEquals     []string `json:"ErrorEquals"`
	IntervalSeconds int      `json:"IntervalSeconds"`
	MaxAttempts     int      `json:"MaxAttempts"`
}

type CatchDefinition struct {
	ErrorEquals []string `json:"ErrorEquals"`
	Next        string   `json:"Next"`
}

type PassStateDefinition struct {
	Name string `json:"-"`
	Type string `json:"Type"`
	Next string `json:"Next"`
}

type ChoiceStateDefinition struct {
	Name    string `json:"-"`
	Type    string `json:"Type"`
	Choices []struct {
		Condition map[string]any `json:"Condition"`
		Next      string         `json:"Next"`
	} `json:"Choices"`
	Default string `json:"Default"`
}

type WaitStateDefinition struct {
	Name    string `json:"-"`
	Type    string `json:"Type"`
	Seconds int    `json:"Seconds"`
	Next    string `json:"Next"`
}

type ParallelStateDefinition struct {
	Name     string                   `json:"-"`
	Type     string                   `json:"Type"`
	Branches []StateMachineDefinition `json:"Branches"`
	Next     string                   `json:"Next"`
}

type MapStateDefinition struct {
	Name       string                 `json:"-"`
	Type       string                 `json:"Type"`
	InputPath  string                 `json:"InputPath"`
	ResultPath string                 `json:"ResultPath"`
	Next       string                 `json:"Next"`
	Iterator   StateMachineDefinition `json:"Iterator"`
}

// PassState simply passes its input to its output, optionally modifying it.
type PassState struct {
	name     string
	next     string
	modifier func(sc *StateContext)
}

func (s *PassState) GetName() string {
	return s.name
}

func (s *PassState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("Executing PassState: %s\n", s.name)
	if s.modifier != nil {
		s.modifier(sc)
	}
	return machine.GetState(s.next), nil
}

// WaitState pauses the workflow for a specified duration.
type WaitState struct {
	name    string
	seconds int
	next    string
}

func (s *WaitState) GetName() string {
	return s.name
}

func (s *WaitState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("Executing WaitState: %s. Pausing for %d seconds...\n", s.name, s.seconds)
	time.Sleep(time.Duration(s.seconds) * time.Second)
	return machine.GetState(s.next), nil
}

// FailState is a terminal state for handling errors.
type FailState struct {
	name string
}

func (s *FailState) GetName() string {
	return s.name
}

func (s *FailState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Printf("State machine failed in state: %s\n", s.name)
	return nil, fmt.Errorf("failure in state %s", s.name)
}

// EndState is the final state.
type EndState struct {
	name string
}

func (s *EndState) GetName() string {
	return s.name
}

func (s *EndState) Execute(ctx context.Context, sc *StateContext, machine *StateMachine) (State, error) {
	fmt.Println("Executing EndState...")
	return nil, nil
}
