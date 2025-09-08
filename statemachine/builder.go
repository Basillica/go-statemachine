package statemachine

import "fmt"

// StateMachineBuilder provides a fluent API for defining the state machine.
type StateMachineBuilder struct {
	states  map[string]State
	startAt string
}

func NewStateMachineBuilder() *StateMachineBuilder {
	return &StateMachineBuilder{
		states: make(map[string]State),
	}
}

func (b *StateMachineBuilder) StartAt(name string) *StateMachineBuilder {
	b.startAt = name
	return b
}

func (b *StateMachineBuilder) AddTask(name string, fn TaskFn, nextState string, options ...any) *StateMachineBuilder {
	task := &TaskState{name: name, execute: fn, next: nextState}
	for _, opt := range options {
		if retry, ok := opt.(RetryRule); ok {
			task.retries = append(task.retries, retry)
		} else if catch, ok := opt.(CatchRule); ok {
			task.catches = append(task.catches, catch)
		} else if timeout, ok := opt.(int); ok {
			task.TimeoutSeconds = timeout
		} else if end, ok := opt.(bool); ok && end {
			task.end = true
		}
	}
	b.states[name] = task
	return b
}

func (b *StateMachineBuilder) AddPass(name string, nextState string, modifier func(sc *StateContext)) *StateMachineBuilder {
	b.states[name] = &PassState{name: name, next: nextState, modifier: modifier}
	return b
}

func (b *StateMachineBuilder) AddMap(name string, inputKey, resultKey string, branch *StateMachine, nextState string) *StateMachineBuilder {
	b.states[name] = &MapState{name: name, input: inputKey, result: resultKey, branch: branch, next: nextState}
	return b
}

func (b *StateMachineBuilder) AddChoice(name string, choices []ChoiceRule, defaultState string) *StateMachineBuilder {
	b.states[name] = &ChoiceState{name: name, choices: choices, defaultState: defaultState}
	return b
}

func (b *StateMachineBuilder) AddWait(name string, seconds int, nextState string) *StateMachineBuilder {
	b.states[name] = &WaitState{name: name, seconds: seconds, next: nextState}
	return b
}

func (b *StateMachineBuilder) AddParallel(name string, branches []*StateMachine, nextState string) *StateMachineBuilder {
	b.states[name] = &ParallelState{name: name, branches: branches, next: nextState}
	return b
}

func (b *StateMachineBuilder) AddFail(name string) *StateMachineBuilder {
	b.states[name] = &FailState{name: name}
	return b
}

func (b *StateMachineBuilder) AddEnd(name string) *StateMachineBuilder {
	b.states[name] = &EndState{name: name}
	return b
}

func (b *StateMachineBuilder) Build() (*StateMachine, error) {
	startState, ok := b.states[b.startAt]
	if !ok {
		return nil, fmt.Errorf("start state '%s' not found", b.startAt)
	}
	return &StateMachine{
		states:       b.states,
		currentState: startState,
		startAt:      b.startAt,
	}, nil
}

func (b *StateMachineBuilder) BuildOrDie() *StateMachine {
	sm, err := b.Build()
	if err != nil {
		panic(err)
	}
	return sm
}
